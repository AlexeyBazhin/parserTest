package awsS3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	EuCentral1 = "eu-central-1"
	USEast1    = "us-east-1"
)

type S3 struct {
	s3Client *s3.S3
	sess     *session.Session
	Endpoint string
}

func New(endpoint, id, secret string) (*S3, error) {
	awsSession, err := session.NewSession(
		aws.NewConfig().
			WithCredentials(
				credentials.NewStaticCredentials(
					id,
					secret,
					id)).
			WithRegion(USEast1).
			WithEndpoint(endpoint).
			WithDisableSSL(*aws.Bool(false)).
			WithS3ForcePathStyle(*aws.Bool(true)),
	)
	if err != nil {
		return nil, err
	}
	s3Client := s3.New(awsSession, aws.NewConfig().WithEndpoint(endpoint))
	return &S3{
		s3Client: s3Client,
		sess:     awsSession,
		Endpoint: endpoint,
	}, nil
}

func (s3Wrap *S3) ListBucketsWithLocation() error {
	listBucketsOutput, err := s3Wrap.s3Client.ListBuckets(nil)
	if err != nil {
		return err
	}
	if len(listBucketsOutput.Buckets) == 0 {
		log.Println("No buckets found")
	}
	for _, bucket := range listBucketsOutput.Buckets {
		log.Printf("* %s created on %s\n",
			aws.StringValue(bucket.Name), aws.TimeValue(bucket.CreationDate))
		locationOutput, err := s3Wrap.s3Client.GetBucketLocation(&s3.GetBucketLocationInput{
			Bucket: bucket.Name,
		})
		if err == nil {
			log.Printf("Location: %s\n", *locationOutput.LocationConstraint)
		} else {
			log.Fatal("Location not defined")
		}
	}
	return nil
}

func (s3Wrap *S3) CreateBucket(bucketName string) error {
	_, err := s3Wrap.s3Client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		// CreateBucketConfiguration: &s3.CreateBucketConfiguration{
		// 	LocationConstraint: aws.String(USEast1),
		// },
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println("Bucket name already in use!")
				panic(err)
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println("Bucket exists and is owned by you!")
			default:
				panic(err)
			}
		}
	}
	if err := s3Wrap.s3Client.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	}); err != nil {
		return err
	}
	return nil
}

func (s3Wrap *S3) PushJSON(bucket string, filename string, goodMap map[string]any) error {
	data, err := json.MarshalIndent(goodMap, "", " ")
	if err != nil {
		log.Fatalf("Cannot marshal good to JSON: %q", err)
		return err
	}
	reader := bytes.NewReader(data)
	if _, err := s3manager.NewUploader(s3Wrap.sess).
		Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(filename),
			Body:   reader,
		}); err != nil {
		log.Fatalf("Cannot upload file to s3: %q", err)
		return err
	}
	return nil
}

func (s3Wrap *S3) ListObjects(bucket string) {
	resp, err := s3Wrap.s3Client.ListObjectsV2(
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
		})
	if err != nil {
		log.Fatalf("Cannot get list of objets")
	}
	// reso, err := s3.s3Client.ListObjectsPages()
	for _, item := range resp.Contents {
		fmt.Println("Name:         ", *item.Key)
		fmt.Println("Last modified:", *item.LastModified)
		fmt.Println("Size:         ", *item.Size)
		fmt.Println("Storage class:", *item.StorageClass)
		fmt.Println("")
	}
}

func (s3Wrap *S3) DownloadFile(bucket string, item string) ([]byte, error) {
	file := aws.NewWriteAtBuffer([]byte{})
	if _, err := s3manager.NewDownloader(s3Wrap.sess).
		Download(file,
			&s3.GetObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(item),
			}); err != nil {
		log.Fatal(err)
		return nil, err
	}
	return file.Bytes(), nil
}
