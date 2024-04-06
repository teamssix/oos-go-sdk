package sample

import (
	"fmt"
	"io/ioutil"

	"oos-go-sdk/oos"
)

func GetObjectMultipartSample() {

	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// multipart upload with 100K as part size. By default 1 coroutine is used and no checkpoint is used.
	err = bucket.UploadFile(objectKeyMultipart, localFileMultipart, 5*1024*1024)
	if err != nil {
		HandleError(err)
	}

	body, err := bucket.GetObject(objectKeyMultipart)
	if err != nil {
		HandleError(err)
	}
	data, err := ioutil.ReadAll(body)
	defer body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("read data len:%d\n", len(data))

	// Case 1: Big file's multipart download, concurrent and resumable download is supported.
	// multipart download with part size 100KB. By default single coroutine is used and no checkpoint
	err = bucket.DownloadFile(objectKeyMultipart, "mynewfile-multipart.file", 5*1024*1024)
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines are used
	err = bucket.DownloadFile(objectKeyMultipart, "mynewfile-multipart.file", 5*1024*1024, oos.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}
}
