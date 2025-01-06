package bootstrap

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	cf "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cfTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type Aws struct {
	bucketPrefix string
	cfClient     CloudFormationClient
	cfWaitClient CloudFormationStackWaiter
}

func NewAws(prefix string) Aws {
	awsBackend := Aws{}
	awsBackend.bucketPrefix = prefix

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("eu-central-1"))
	if err != nil {
		log.Fatal(err)
	}

	awsBackend.cfClient = cf.NewFromConfig(cfg)
	awsBackend.cfWaitClient = cf.NewStackCreateCompleteWaiter(awsBackend.cfClient)

	return awsBackend
}

type CloudFormationClient interface {
	ListStacks(ctx context.Context, params *cf.ListStacksInput, optFns ...func(*cf.Options)) (*cf.ListStacksOutput, error)
	DescribeStacks(ctx context.Context, params *cf.DescribeStacksInput, optFns ...func(*cf.Options)) (*cf.DescribeStacksOutput, error)
	CreateStack(ctx context.Context, params *cf.CreateStackInput, optFns ...func(*cf.Options)) (*cf.CreateStackOutput, error)
}

type CloudFormationStackWaiter interface {
	Wait(ctx context.Context, params *cf.DescribeStacksInput, maxWaitDur time.Duration, optFns ...func(*cf.StackCreateCompleteWaiterOptions)) error
}

// Check if CloudFormation with a tag TF_BOOTSTRAP = true exists.
// Check if Bucket Exists.
// Run ClodFormation to create the bucket.
// Generate the backend.tf file.
// Show boootstrapped regions.

// Check if CloudFormation with a tag TF_BOOTSTRAP = true exists.
func (a *Aws) checkIfBootstrapped() bool {
	activeStatuses := []cfTypes.StackStatus{
		cfTypes.StackStatusCreateComplete,
		cfTypes.StackStatusUpdateComplete,
		cfTypes.StackStatusUpdateRollbackComplete,
	}

	output, err := a.cfClient.ListStacks(context.Background(), &cf.ListStacksInput{
		StackStatusFilter: activeStatuses,
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, stackSummary := range output.StackSummaries {
		stackOutput, err := a.cfClient.DescribeStacks(context.Background(), &cf.DescribeStacksInput{
			StackName: stackSummary.StackName,
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, stack := range stackOutput.Stacks {
			for _, tag := range stack.Tags {
				if *tag.Key == "TF_BOOTSTRAP" && *tag.Value == "true" {
					log.Printf("Found bootstrapped stack: %s", *stack.StackName)
					return true
				}
			}
		}
	}
	return false
}

func (a *Aws) generateTemplate() string {
	const templateBody = `
AWSTemplateFormatVersion: '2010-09-09'
Description: CloudFormation template to create an S3 bucket with versioning enabled

Resources:
  S3Bucket:
    Type: 'AWS::S3::Bucket'
    Properties:
      BucketName: !Join
        - "-"
        - - {{ .BucketPrefix }}
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
`
	tmpl, err := template.New("cfTemplate").Parse(templateBody)
	if err != nil {
		log.Fatal(err)
	}

	var renderedTemplateWritter bytes.Buffer
	err = tmpl.Execute(&renderedTemplateWritter, struct {
		BucketPrefix string
	}{
		BucketPrefix: a.bucketPrefix,
	})
	if err != nil {
		log.Fatal(err)
	}
	return renderedTemplateWritter.String()

}

func (a *Aws) executeTemplate(templateBody string) {
	stackName := "TFBootstraper"

	_, err := a.cfClient.CreateStack(context.Background(), &cf.CreateStackInput{
		StackName:    aws.String(stackName),
		TemplateBody: &templateBody,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Stack creation initiated, waiting for completion... 5 min max")

	err = a.cfWaitClient.Wait(context.Background(), &cf.DescribeStacksInput{
		StackName: aws.String(stackName),
	}, time.Minute*5)
	if err != nil {
		log.Fatal(err)
	}

	stackOutput, err := a.cfClient.DescribeStacks(context.Background(), &cf.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, stack := range stackOutput.Stacks {
		fmt.Printf("Stack Name: %s\n", *stack.StackName)
		for _, output := range stack.Outputs {
			if *output.OutputKey == "BucketName" {
				fmt.Printf("Bucket Name: %s\n", *output.OutputValue)
			}
		}
	}

	log.Println("Stack creation complete")
}

func (a *Aws) Run() {
	if a.checkIfBootstrapped() {
		log.Fatal("TF Bootstrapper stack is already exist")
	}
	templateBody := a.generateTemplate()
	a.executeTemplate(templateBody)
}

// stackName, err := a.cfClient.CreateStack(context.Background(), &cf.CreateStackInput{
// 	StackName:    aws.String(stackName),
// 	TemplateBody: &templateBody,
// })
// if err != nil {
// 	log.Fatal(err)
// }
//
