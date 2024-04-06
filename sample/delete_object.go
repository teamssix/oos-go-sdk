package sample

import (
	"fmt"
	"strings"

	"oos-go-sdk/oos"
)

// DeleteObjectSample shows how to delete single file or multiple files
func DeleteObjectSample() {
	// Create a bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	var val = "千山鸟飞绝，万径人踪灭。孤舟蓑笠翁，独钓寒江雪。"

	err = bucket.PutObject(objectKey, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObject(objectKey+"one", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	err = bucket.PutObject(objectKey+"two", strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	delRes, err := bucket.DeleteObjects([]string{objectKey + "one", objectKey + "two"}, oos.DeleteObjectsQuiet(false))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("Sample Del Res:", delRes)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("DeleteObjectSample completed")
}
