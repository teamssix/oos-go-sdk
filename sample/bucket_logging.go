package sample

import (
	"fmt"
)

// BucketLoggingSample shows how to set, get and delete the bucket logging configuration
func BucketLoggingSample() {
	// New client
	client := NewClient()

	// Create target bucket to store the logging files.
	var targetBucketName = bucketName

	// Case 1: Set the logging for the object prefixed with "prefix-1" and save their access logs to the target bucket
	err := client.SetBucketLogging(bucketName, targetBucketName, "prefix-1", true)
	if err != nil {
		HandleError(err)
	}

	// Get the bucket's logging configuration
	logInfo, err := client.GetBucketLogging(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("Bucket Logging: %s \r\n", logInfo.LoggingEnabled.TargetBucket)

	fmt.Println("BucketLoggingSample completed")
}
