package sample

import (
	"fmt"

	"oos-go-sdk/oos"
)

// ListObjectsSample shows the file list, including default and specified parameters.
func ListObjectsSample() {
	var myObjects = []Object{
		{"my-object-11", "test_1"},
		{"my-object-12", "test_2"},
		{"my-object-13", "test_3"},
		{"my-object-14", "test_4"},
		{"my-object-21", "test_5"},
		{"my-object-22", "test_6"},
		{"my-object-23", "test_7"},
		{"my-object-24", "test_8"}}

	// Create bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Create objects
	err = CreateObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Use default parameters*/
	lor, err := bucket.ListObjects()
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects:", getObjectsFormResponse(lor))

	// Case 2: List object with paging , with prefix and max keys; return 3 items each time.
	pre := oos.Prefix("my-object-2")
	marker := oos.Marker("")
	for {
		lor, err := bucket.ListObjects(oos.MaxKeys(3), marker, pre)
		if err != nil {
			HandleError(err)
		}
		pre = oos.Prefix(lor.Prefix)
		marker = oos.Marker(lor.NextMarker)
		fmt.Println("my objects prefix&page :", getObjectsFormResponse(lor))
		if !lor.IsTruncated {
			break
		}
	}

	err = DeleteObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Combine the prefix and delimiter for grouping. ListObjectsResponse.Objects is the objects returned.
	// ListObjectsResponse.CommonPrefixes is the common prefixes returned.
	myObjects = []Object{
		{"fun/test.txt", "111"},
		{"fun/test.jpg", "111"},
		{"fun/movie/001.avi", "111"},
		{"fun/movie/007.avi", "111"},
		{"fun/music/001.mp3", "111"},
		{"fun/music/001.mp3", "111"}}

	// Create object
	err = CreateObjects(bucket, myObjects)
	if err != nil {
		HandleError(err)
	}

	lor, err = bucket.ListObjects(oos.Prefix("fun/"), oos.Delimiter("/"))
	if err != nil {
		HandleError(err)
	}
	fmt.Println("my objects prefix :", getObjectsFormResponse(lor),
		"common prefixes:", lor.CommonPrefixes)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("ListObjectsSample completed")
}

func getObjectsFormResponse(lor oos.ListObjectsResult) string {
	var output string
	for _, object := range lor.Objects {
		output += object.Key + "  "
	}
	return output
}
