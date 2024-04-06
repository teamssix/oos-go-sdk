package sample

import (
	"fmt"
	"io/ioutil"
	"strings"

	"oos-go-sdk/oos"
)

// SignURLSample signs URL sample
func SignURLSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	objectName := "test中文.txt"
	var val = "红豆生南国，春来发几枝。愿君多采撷，此物最相思。"
	err = bucket.PutObject(objectName, strings.NewReader(val))
	if err != nil {
		HandleError(err)
	}

	options := []oos.Option{
		oos.LimitForParam("rate=1000"),
	}
	signedURL, err := bucket.SignURL(objectName, oos.HTTPGet, 3600, options...)
	if err != nil {
		HandleError(err)
	}
	fmt.Println(signedURL)

	body, err := bucket.GetObjectWithURL(signedURL)
	if err != nil {
		HandleError(err)
	}
	// Read content
	data, err := ioutil.ReadAll(body)
	if err != nil {
		HandleError(err)
	}
	body.Close()
	fmt.Println(string(data))

	err = bucket.GetObjectToFileWithURL(signedURL, "mynewfile-1.jpg")
	if err != nil {
		HandleError(err)
	}

	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("SignURLSample completed")
}
