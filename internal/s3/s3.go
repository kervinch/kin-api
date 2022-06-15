package s3

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/kervinch/internal/data"
)

const (
	URL              = "https://kin-public.s3.ap-southeast-1.amazonaws.com/"
	REGION           = "ap-southeast-1"
	BANNER           = "banners/"
	BRAND            = "brands/"
	BLOG             = "blogs/"
	BLOG_CATEGORY    = "blog_categories/"
	INBOX            = "inbox/"
	PRODUCT          = "products/"
	PRODUCT_CATEGORY = "product_categories/"
	PRODUCT_REFUND   = "product_refunds/"
	STOREFRONT       = "storefronts/"
	VOUCHER          = "vouchers/"
)

type S3 struct {
	bucketName string
}

func New(bucketName string) S3 {
	return S3{
		bucketName: bucketName,
	}
}

func (s S3) ListBuckets() error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION)},
	)
	if err != nil {
		return err
	}

	// Create an S3 session.
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		return err
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	return nil
}

func (s S3) Upload(file multipart.File, key, filename, contentType string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION)},
	)
	if err != nil {
		return "", err
	}

	// Create an S3 session.
	svc := s3.New(sess)

	// Check if file is image
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filetype := http.DetectContentType(buff)
	switch filetype {
	case "image/jpeg", "image/jpg":
	case "image/png":
	case "image/gif":
		return "", data.ErrImageFormat
	case "application/pdf":
		return "", data.ErrImageFormat
	default:
		return "", data.ErrImageFormat
	}

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:        file,
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key + filename),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", err
	}

	return URL + key + filename, nil
}

func (s S3) UploadVideo(file multipart.File, key, filename, contentType string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(REGION)},
	)
	if err != nil {
		return "", err
	}

	// Create an S3 session.
	svc := s3.New(sess)

	// Check if file is image
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	filetype := http.DetectContentType(buff)
	switch filetype {
	case "video/mp4":
	case "video/x-flv":
	case "application/x-mpegURL":
	case "video/MP2T":
	case "video/3gpp":
	case "video/quicktime":
	case "video/x-msvideo":
	case "video/x-ms-wmv":
	default:
		return "", data.ErrVideoFormat
	}

	// Upload the file to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Body:        file,
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key + filename),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", err
	}

	return URL + key + filename, nil
}
