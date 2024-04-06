package sample

import "oos-go-sdk/oos"

func PutObjectMultipartSample() {

	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Big file's multipart upload. It supports concurrent upload with resumable upload.
	// multipart upload with 100K as part size. By default 1 coroutine is used and no checkpoint is used.
	err = bucket.UploadFile(objectKeyMultipart, localFileMultipart, 5*1024*1024)
	if err != nil {
		HandleError(err)
	}

	// Part size is 100K and 3 coroutines are used
	err = bucket.UploadFile(objectKeyMultipart, localFileMultipart, 5*1024*1024, oos.Routines(3))
	if err != nil {
		HandleError(err)
	}

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

}
