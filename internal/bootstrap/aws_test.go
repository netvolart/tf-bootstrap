package bootstrap

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
)

type mockCloudFormationClient struct {
	ListStacksFunc     func(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
	DescribeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	CreateStackFunc  func(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
}

func (m *mockCloudFormationClient) ListStacks(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
	return m.ListStacksFunc(ctx, params, optFns...)
}

func (m *mockCloudFormationClient) DescribeStacks(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	return m.DescribeStacksFunc(ctx, params, optFns...)
}

func (m *mockCloudFormationClient) CreateStack(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	return m.CreateStackFunc(ctx, params, optFns...)
}


func Test_checkIfBootstrapped(t *testing.T) {
	tests := []struct {
		name               string
		listStacksFunc     func(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
		describeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
		expectedResult     bool
	}{
		{
			name: "Stack with TF_BOOTSTRAP=true tag",
			listStacksFunc: func(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
				return &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("test-stack"),
						},
					},
				}, nil
			},
			describeStacksFunc: func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
				return &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName: aws.String("test-stack"),
							Tags: []types.Tag{
								{
									Key:   aws.String("TF_BOOTSTRAP"),
									Value: aws.String("true"),
								},
							},
						},
					},
				}, nil
			},
			expectedResult: true,
		},
		{
			name: "Stack without TF_BOOTSTRAP=true tag",
			listStacksFunc: func(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error) {
				return &cloudformation.ListStacksOutput{
					StackSummaries: []types.StackSummary{
						{
							StackName: aws.String("test-stack"),
						},
					},
				}, nil
			},
			describeStacksFunc: func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
				return &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName: aws.String("test-stack"),
							Tags: []types.Tag{
								{
									Key:   aws.String("TF_BOOTSTRAP"),
									Value: aws.String("false"),
								},
							},
						},
					},
				}, nil
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCFClient := &mockCloudFormationClient{
				ListStacksFunc:     tt.listStacksFunc,
				DescribeStacksFunc: tt.describeStacksFunc,
			}

			awsBackend := &Aws{
				cfClient: mockCFClient,
			}

			result := awsBackend.checkIfBootstrapped()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func Test_generateTemplate(t *testing.T) {
	tests := []struct {
		name       string
		BucketName string
	}{
		{
			name:       "Valid bucket name",
			BucketName: "my-unique-bucket-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			awsBackend := &Aws{
				bucketPrefix: tt.BucketName,
			}
			result := awsBackend.generateTemplate()
			assert.Contains(t, result, tt.BucketName)
			assert.Contains(t, result, "AWSTemplateFormatVersion: '2010-09-09'")
			assert.Contains(t, result, "Description: CloudFormation template to create an S3 bucket with versioning enabled")
			assert.Contains(t, result, "Type: 'AWS::S3::Bucket'")
			assert.Contains(t, result, "VersioningConfiguration:\n        Status: Enabled")
			assert.Contains(t, result, "Outputs:\n  BucketName:\n    Description: Name of the S3 bucket\n    Value: !Ref S3Bucket")
		})
	}
}
