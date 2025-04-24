package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/stretchr/testify/assert"
)

type mockCloudFormationClient struct {
	ListStacksFunc     func(ctx context.Context, params *cloudformation.ListStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.ListStacksOutput, error)
	DescribeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	CreateStackFunc    func(ctx context.Context, params *cloudformation.CreateStackInput, optFns ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
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
		describeStacksFunc func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
		expectedResult     bool
	}{
		{
			name: "Stack with TF_BOOTSTRAP=true tag",
			describeStacksFunc: func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
				return &cloudformation.DescribeStacksOutput{
					Stacks: []types.Stack{
						{
							StackName: aws.String("TFBoot"),
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
			name: "Check if stack is not exists",
			describeStacksFunc: func(ctx context.Context, params *cloudformation.DescribeStacksInput, optFns ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
				return nil, fmt.Errorf("Error")
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCFClient := &mockCloudFormationClient{
				DescribeStacksFunc: tt.describeStacksFunc,
			}

			awsBackend := &AwsBackendService{
				cfClient: mockCFClient,
			}
			ctx := context.Background()

			result, _ := awsBackend.checkIfBootstrapped(ctx)
			assert.Equal(t, tt.expectedResult, result)
			
		})
	}
}
