package sample

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"oos-go-sdk/oos"
)

// PutObjectSample illustrates two methods for uploading a file: simple upload and multipart upload.
func PutObjectSample() {
	// Create bucketfunc PutObjectSample() {
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var val = "红豆生南国，春来发几枝。愿君多采撷，此物最相思。"

	// Case 1: Upload an object from a string
	objectName := "test.txt"
	err = bucket.PutObject(objectName, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	// Case 2: Upload an object whose value is a byte[]
	err = bucket.PutObject("test_byte", bytes.NewReader([]byte(val)))
	if err != nil {
		HandleError(err)
	}

	// Case 3: Upload the local file with file handle, user should open the file at first.
	fd, err := os.Open(localFile)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()

	err = bucket.PutObject("test_file_fd", fd)
	if err != nil {
		HandleError(err)
	}

	// Case 4: Upload an object with local file name, user need not open the file.
	err = bucket.PutObjectFromFile("test_file_path", localFile)
	if err != nil {
		HandleError(err)
	}

	// Case 5: Upload an object with specified properties, PutObject/PutObjectFromFile/UploadFile also support this feature.
	options := []oos.Option{
		oos.Expires(futureDate),
		oos.Meta("mykey", "myval"),
		oos.StorageClass("STANDARD"),
		oos.Connection("close"),
		oos.ContentType("jpeg"),
		oos.ObjectDataLocation(oos.BuildObjectLocalDataLocation(false)),
		oos.ObjectDataLocation(oos.BuildObjectSpecifiedDataLocation("ChengDu", false)),
	}

	err = bucket.PutObject("test_meta", strings.NewReader(val), options...)
	if err != nil {
		HandleError(err)
	}

	isExist, err := bucket.IsObjectExist(objectName)
	fmt.Printf("isExist:%v, err:%v\n", isExist, err)
	props, err := bucket.HeadObject(objectName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Object Meta:", props)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("PutObjectSample completed")
}
