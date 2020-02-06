package gimage

import (
	"errors"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	uuid "github.com/satori/go.uuid"
)

var (
	// ErrAWSKeyRequired is returned when new s3uploader instance is created with a blank aws key
	ErrAWSKeyRequired = errors.New("AWS key is required to create an uploader")
	// ErrAWSSecretRequired is returned when new s3uploader instance is created with a blank aws key
	ErrAWSSecretRequired = errors.New("AWS secret is required to create an uploader")
	// ErrAWSBucketRequired is returned when new s3uploader instance is created with a blank aws key
	ErrAWSBucketRequired = errors.New("S3 bucket name is required to create an uploader")
)

// S3Config is the configuration struct for instanciating a new S3Uploader
type S3Config struct {
	AWSKeyID  string
	AWSSecret string
	AWSRegion string
	S3Bucket  string
}

// UploadRequest is the request structure for the uploading
type UploadRequest struct {
	Bucket        string    // Bucket is the s3 bucket where the file needs to be uploaded
	ContentType   string    // ContentType of the file to be uploaded. This is passed to s3 object
	UploadPath    string    // UploadPath is the s3'sd logical folder structure
	FileExtension string    // FileExtension of the file being uploaded
	Payload       io.Reader // Payload is the actual file.
}

// GetUploadKey generates the path to which the image would be uploaded win S3.
func (ur UploadRequest) GetUploadKey(filename string) string {
	return ur.UploadPath + filename + "." + ur.FileExtension
}

// S3Uploader encapsulates uploading to S3 bucket.
type S3Uploader struct {
	s3Uploader *s3manager.Uploader
}

// Upload uploads the content to the specified bucket.
func (s3u *S3Uploader) Upload(req UploadRequest) (string, error) {
	filename := s3u.generateUniqueName()
	uploadParams := &s3manager.UploadInput{
		Bucket:      aws.String(req.Bucket),
		Key:         aws.String(req.GetUploadKey(filename)),
		Body:        req.Payload,
		ContentType: aws.String(req.ContentType),
	}

	out, err := s3u.s3Uploader.Upload(uploadParams)
	if err != nil {
		return "", err
	}

	return out.Location, nil
}

// generateUniqueName generates a new unique string for the filename.
func (s3u *S3Uploader) generateUniqueName() string {
	uid := uuid.NewV4()
	idString := uid.String()

	idString = strings.Join(strings.Split(idString, "-"), "")
	return strings.ToUpper(idString)
}

// NewS3Uploader returns a new S3Uploader.
func NewS3Uploader(config S3Config) (*S3Uploader, error) {
	if config.AWSKeyID == "" {
		return nil, ErrAWSKeyRequired
	}

	if config.AWSSecret == "" {
		return nil, ErrAWSSecretRequired
	}

	if config.S3Bucket == "" {
		return nil, ErrAWSBucketRequired
	}

	if config.AWSRegion == "" {
		config.AWSRegion = s3.BucketLocationConstraintApSouth1
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.AWSRegion),
		Credentials: credentials.NewStaticCredentials(config.AWSKeyID, config.AWSSecret, ""),
	})

	if err != nil {
		return nil, err
	}

	s3Service := s3.New(awsSession)
	uploader := s3manager.NewUploaderWithClient(s3Service)

	return &S3Uploader{s3Uploader: uploader}, nil
}
