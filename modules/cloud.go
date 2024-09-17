package modules

import (
  "context"
  "fmt"
  "strings"
  "os"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/service/s3"
)

type CloudProvider int


const (
	AWS CloudProvider =  iota
	GCP
);

type CloudStorageClient interface{
	UploadFile(ctx context.Context,localFilePath string,remoteFilePath string) error
}

type AWSStorageClient struct{
	S3Client *s3.Client
	bucket string
}

func NewAWSStorageClientLoadDefault(ctx context.Context,bucket string) (*AWSStorageClient,error){
	cfg,err := config.LoadDefaultConfig(ctx)
	if err!=nil{
		return nil,fmt.Errorf("failed to load shared config, set up or try another options : %v",err);
	}
	
	client := s3.NewFromConfig(cfg)

	return &AWSStorageClient{
		S3Client : client,
		bucket : bucket,
	},nil
	
}

func NewAWSStorageClient(ctx context.Context,accessKeyID,secretAccessKey,region,bucket string)(*AWSStorageClient, error){
	cfg,err := config.LoadDefaultConfig(ctx)
	if err!=nil{
		return nil,fmt.Errorf("failed to load shared config, set up or try another options : %v",err);
	}
	
	client := s3.NewFromConfig(cfg)

	return &AWSStorageClient{
		S3Client : client,
		bucket : bucket,
	},nil
	
}

func (a *AWSStorageClient) UploadFile(ctx context.Context,localFilePath,remoteFilePath string) error{
	file,err := os.Open(localFilePath)
	if err!=nil{
		return 	fmt.Errorf("failed to open file : %v",err)
	}
	defer file.Close()

	_,err = a.S3Client.PutObject(ctx,&s3.PutObjectInput{
		Bucket : aws.String(a.bucket),
		Key : aws.String(remoteFilePath),
		Body : file,
	})
	
	if err!=nil{
		return fmt.Errorf("failed to upload file : %v",err)
	}

	return nil
}

func getUserInput(prompt string) string{
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func UploadFileToCloud(localFilePath string) error{
	ctx := context.Background()
	
	
	remoteFilePath := getUserInput("Enter the desired remote file path: ")
	providerStr := getUserInput("Enter the cloud provider (AWS or GCP): ")

	var client CloudStorageClient
	var err error

	switch strings.ToUpper(providerStr) {
	case "AWS":
		default_or_specified := getUserInput("1) Load Default Config or 2) Specify the config details")
		if default_or_specified == "1" {
			bucket := getUserInput("Enter S3 Bucket Name:")
			client,err=NewAWSStorageClientLoadDefault(ctx,bucket)
			if err!=nil{
				return fmt.Errorf("failed to create AWS client : %v",err)
			}
		} else{
			accessKeyID := getUserInput("Enter AWS Access Key ID: ")
			secretAccessKey := getUserInput("Enter AWS Secret Access Key: ")
			region := getUserInput("Enter AWS Region: ")
			bucket := getUserInput("Enter S3 Bucket Name: ")

			client, err = NewAWSStorageClient(ctx, accessKeyID, secretAccessKey, region, bucket)
			if err != nil {
				return fmt.Errorf("failed to create AWS client: %v", err)
			}
		}
/*
	case "GCP":
		bucket := getUserInput("Enter GCS Bucket Name: ")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", getUserInput("Enter path to Google Cloud service account key file: "))

		client, err = NewGCPStorageClient(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to create GCP client: %v", err)
		}
*/
	default:
		return fmt.Errorf("unsupported cloud provider: %s", providerStr)
	}

	err = client.UploadFile(ctx, localFilePath, remoteFilePath)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	fmt.Println("File uploaded successfully!")
	return nil
}
