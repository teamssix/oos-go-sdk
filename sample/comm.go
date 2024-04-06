package sample

import (
	"fmt"
	"os"
	"strings"
	"time"

	"oos-go-sdk/oos"
)

var (
	pastDate   = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	futureDate = time.Date(2049, time.January, 10, 23, 0, 0, 0, time.UTC)
)

// HandleError is the error handling method in the sample code
func HandleError(err error) {
	fmt.Println("occurred error:", err)
	os.Exit(-1)
}

// create client
func NewClientWithTimeOut(connectTimeoutSec, readWriteTimeout int64) *oos.Client {
	clientOptionV4 := oos.V4Signature(true)
	isEnableSha256 := oos.EnableSha256ForPayload(false)
	timeOut := oos.Timeout(connectTimeoutSec, readWriteTimeout)
	client, err := oos.New(endpoint, accessKey, secretKey, clientOptionV4, isEnableSha256, timeOut)
	if err != nil {
		HandleError(err)
	}
	return client
}

// create client
func NewClient() *oos.Client {
	clientOptionV4 := oos.V4Signature(true)
	isEnableSha256 := oos.EnableSha256ForPayload(false)
	client, err := oos.New(endpoint, accessKey, secretKey, clientOptionV4, isEnableSha256)
	if err != nil {
		HandleError(err)
	}
	return client
}

// create IAM client
func NewIAMClient() *oos.Client {

	clientOptionV4 := oos.V4Signature(true)
	isEnableSha256 := oos.EnableSha256ForPayload(true)
	client, err := oos.New(iamEndpoint, accessKey, secretKey, clientOptionV4, isEnableSha256)
	if err != nil {
		HandleError(err)
	}
	return client
}

// GetTestBucket creates the test bucket
func GetTestBucket(bucketName string) (*oos.Object, error) {
	// New client
	client := NewClient()

	// Get bucket
	bucket, err := client.Bucket(bucketName)

	if err != nil {
		return nil, err
	}

	return bucket, nil
}

// DeleteTestBucketAndObject deletes the test bucket and its objects
func DeleteTestBucketAndObject(bucketName string) error {
	// New client
	client := NewClient()

	// Get bucket
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}

	// Delete objects
	lor, err := bucket.ListObjects()
	if err != nil {
		return err
	}

	for _, object := range lor.Objects {
		err = bucket.DeleteObject(object.Key)
		if err != nil {
			return err
		}
	}

	// Delete bucket
	err = client.DeleteBucket(bucketName)
	if err != nil {
		return err
	}

	return nil
}

// Object defines pair of key and value
type Object struct {
	Key   string
	Value string
}

// CreateObjects creates some objects
func CreateObjects(bucket *oos.Object, objects []Object) error {
	for _, object := range objects {
		err := bucket.PutObject(object.Key, strings.NewReader(object.Value))
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteObjects deletes some objects.
func DeleteObjects(bucket *oos.Object, objects []Object) error {
	for _, object := range objects {
		err := bucket.DeleteObject(object.Key)
		if err != nil {
			return err
		}
	}
	return nil
}
