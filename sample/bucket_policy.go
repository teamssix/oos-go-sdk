package sample

import (
	"fmt"
	"time"
)

// BucketPolicySample shows how to get and set the bucket Policy
func BucketPolicySample() {

	var err error
	// New client
	client := NewClient()

	text := fmt.Sprintf("{\"Version\": \"2012-10-17\",\"Id\": \"preventHotLinking\",\"Statement\": " +
		"[ { \"Sid\": \"1\",\"Effect\": \"Allow\",\"Principal\": { \"AWS\": \"*\"}," +
		"\"Action\": \"s3:*\",\"Resource\": \"arn:aws:s3:::" + bucketName +"\",}]}")

	err = client.SetBucketPolicy(bucketName, text)
	if err != nil {
		HandleError(err)
	}

	time.Sleep(time.Duration(5) * time.Second)

	text, err = client.GetBucketPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Printf("policy info %v\n", text)

	err = client.DeleteBucketPolicy(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("bucket website sample complete")
}
