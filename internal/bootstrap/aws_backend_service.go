package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	cf "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type CloudFormationClient interface {
	ListStacks(ctx context.Context, params *cf.ListStacksInput, optFns ...func(*cf.Options)) (*cf.ListStacksOutput, error)
	DescribeStacks(ctx context.Context, params *cf.DescribeStacksInput, optFns ...func(*cf.Options)) (*cf.DescribeStacksOutput, error)
	CreateStack(ctx context.Context, params *cf.CreateStackInput, optFns ...func(*cf.Options)) (*cf.CreateStackOutput, error)
}

type CloudFormationStackWaiter interface {
	Wait(ctx context.Context, params *cf.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cf.StackCreateCompleteWaiterOptions)) error
}

type AwsBackendService struct {
	cfClient     CloudFormationClient
	cfWaitClient CloudFormationStackWaiter
	region       string
	stackName    string
	templateBody string
}

func NewAwsBackendService(region string) (*AwsBackendService, error) {

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}
	cfClient := cf.NewFromConfig(cfg)

	return &AwsBackendService{
		cfClient: cfClient,
		cfWaitClient: cf.NewStackCreateCompleteWaiter(cfClient),
		region:               region,
		stackName:            "TFBoot",
		templateBody: `
AWSTemplateFormatVersion: '2010-09-09'
Description: CloudFormation template to create an S3 for a Terraform backend

Parameters:
  BucketPrefix:
    Type: String
    Description: The prefix for the S3 bucket name

Resources:
  S3Bucket:
    Type: 'AWS::S3::Bucket'
    Properties:
      BucketName:
        !Join
          - "-"
          -
            - !Ref BucketPrefix
            - !Select
                - 0
                - !Split
                    - "-"
                    - !Select
                        - 2
                        - !Split
                            - "/"
                            - !Ref "AWS::StackId"
      VersioningConfiguration:
        Status: Enabled

Outputs:
  BucketName:
    Description: Name of the S3 bucket
    Value: !Ref S3Bucket
`,
	}, nil
}


func (s *AwsBackendService) createStack(ctx context.Context, namePrefix string) error {
	_, err := s.cfClient.CreateStack(ctx, &cf.CreateStackInput{
		StackName:    aws.String(s.stackName),
		TemplateBody: aws.String(s.templateBody),
		Tags: []types.Tag{
			{Key: aws.String("TFBoot"), Value: aws.String("true")},
		},
		Parameters: []types.Parameter{
			{ParameterKey: aws.String("BucketPrefix"), ParameterValue: aws.String(namePrefix)},
		},
	})
	if err != nil {
		return err
	}
	fmt.Println("Bootstrap process has been started...")
	return nil
}


func (s *AwsBackendService) waitForBootstrapped(ctx context.Context) error {
	for {
		time.Sleep(2 * time.Second)

		resp, err := s.cfClient.DescribeStacks(ctx, &cf.DescribeStacksInput{
			StackName: aws.String(s.stackName),
		})
		if err != nil {
			return err
		}

		if len(resp.Stacks) == 0 {
			return fmt.Errorf("stack %s not found", s.stackName)
		}

		stack := resp.Stacks[0]
		fmt.Printf("Current status: %s\n", stack.StackStatus)
		
		switch stack.StackStatus {
			case types.StackStatusCreateComplete:
				fmt.Printf("Stack %s created successfully\n", s.stackName)
				return nil
			case types.StackStatusCreateFailed, types.StackStatusRollbackComplete:
				return fmt.Errorf("stack creation failed with status: %s", stack.StackStatus)
		}

	}
}

func (s *AwsBackendService) checkIfBootstrapped(ctx context.Context) (bool, error) {
	_, err := s.cfClient.DescribeStacks(ctx, &cf.DescribeStacksInput{
		StackName: aws.String(s.stackName),
	})
	if err != nil  {
		return false, nil
	}
	return true, nil
}

func (s *AwsBackendService) getBucketName(ctx context.Context) (string, error) {
	resp, err := s.cfClient.DescribeStacks(ctx, &cf.DescribeStacksInput{
		StackName: aws.String(s.stackName),
	})
	if err != nil {
		return "", err
	}

	if len(resp.Stacks) == 0 {
		return "", fmt.Errorf("stack %s not found", s.stackName)
	}

	stack := resp.Stacks[0]
	for _, output := range stack.Outputs {
		if *output.OutputKey == "BucketName" {
			if output.OutputValue == nil {
				return "", fmt.Errorf("bucket name is empty in %s", s.stackName)
			}
			return *output.OutputValue, nil
		}
	}

	return "", fmt.Errorf("bucket name not found in stack outputs")
}

func (s *AwsBackendService) Run(ctx context.Context, namePrefix string) error {
	bootstrapped, err := s.checkIfBootstrapped(ctx)
	if err != nil {
		return err
	}
	if bootstrapped {
		fmt.Println("The stack is already bootstrapped. No further action is required")
		return nil
	}
	

	if err := s.createStack(ctx, namePrefix); err != nil {
		return err
	}

	if err := s.waitForBootstrapped(ctx); err != nil {
		return err
	}

	bucketName, err := s.getBucketName(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Bucket:", bucketName)
	return nil
}


func (s *AwsBackendService) Show(ctx context.Context) error {
	bootstrapped, err := s.checkIfBootstrapped(ctx)
	if err != nil {
		return err
	}
	if bootstrapped {
		bucketName, err := s.getBucketName(ctx)
		if err != nil {
			return err
		}

		fmt.Println(bucketName)
		return nil
	}

	return nil
}