package sample

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"oos-go-sdk/oos"
)

// GetObjectSample shows the streaming download, range download and resumable download.
func GetObjectSample() {
	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Upload the object
	err = bucket.PutObjectFromFile(objectKey, localFile, oos.ObjectDataLocation(oos.BuildObjectSpecifiedDataLocation("ChengDu", true)))
	if err != nil {
		HandleError(err)
	}

	//check location
	result, _ := bucket.DoGetObject(&oos.GetObjectRequest{ObjectKey: objectKey}, []oos.Option{})
	fmt.Println(result.Response.Headers.Get("x-ctyun-metadata-location"))
	fmt.Println(result.Response.Headers.Get("x-ctyun-data-location"))

	// Case 1: Download the object into ReadCloser(). The body needs to be closed
	var body io.ReadCloser
	body, err = bucket.GetObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	data, err := ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("read data len:%d\n", len(data))

	// Case 2: Download the object to local file with file name specified
	err = bucket.GetObjectToFile(objectKey, "mynewfile-2.jpg")
	if err != nil {
		HandleError(err)
	}

	// Case 3: Get the object with contraints. When contraints are met, download the file. Otherwise return precondition error
	// last modified time constraint is met, download the file
	body, err = bucket.GetObject(objectKey, oos.IfModifiedSince(pastDate))
	if err != nil {
		HandleError(err)
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("read data len:%d\n", len(data))

	// Last modified time contraint is not met, do not download the file
	_, err = bucket.GetObject(objectKey, oos.IfUnmodifiedSince(pastDate))
	if err == nil {
		HandleError(err)
	}

	meta, err := bucket.HeadObject(objectKey)
	if err != nil {
		HandleError(err)
	}
	etag := meta.Get(oos.HTTPHeaderEtag)
	etag = strings.Trim(etag, "\"")
	// Check the content, etag contraint is met, download the file
	body, err = bucket.GetObject(objectKey, oos.IfMatch(etag))
	if err != nil {
		HandleError(err)
	}
	data, err = ioutil.ReadAll(body)
	body.Close()
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("read data len:%d\n", len(data))

	// Check the content, etag contraint is not met, do not download the file
	body, err = bucket.GetObject(objectKey, oos.IfNoneMatch(etag))
	if err == nil {
		HandleError(err)
	}
	fmt.Print(body)
	// Delete the object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("GetObjectSample completed")
}
