package sample

import (
	"fmt"

	"oos-go-sdk/oos"
)

// CopyObjectSample shows the copy files usage
func CopyObjectSample() {

	// Create a bucket
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	// Create an object
	err = bucket.PutObjectFromFile(objectKey, localFile)
	if err != nil {
		HandleError(err)
	}

	// Case 1: Copy an existing object
	var descObjectKey = "descobject"
	//var descObjectKey = "film/descobject"
	_, err = bucket.CopyObject(objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Copy an existing object to another existing object
	_, err = bucket.CopyObjectFrom(bucketName, objectKey, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	err = bucket.DeleteObject(descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// Case 3: Copy file with constraints. When the constraints are met, the copy executes. otherwise the copy does not execute.
	// constraints are not met, copy does not execute
	_, err = bucket.CopyObjectTo(bucketName, descObjectKey, objectKey, oos.CopySourceIfModifiedSince(futureDate))
	fmt.Println("CopyObjectError:", err)

	// Constraints are met, the copy executes
	_, err = bucket.CopyObject(objectKey, descObjectKey, oos.CopySourceIfUnmodifiedSince(futureDate))
	if err != nil {
		HandleError(err)
	}

	// Case 4: Specify the properties when copying. The MetadataDirective needs to be MetaReplace
	options := []oos.Option{
		oos.Expires(futureDate),
		oos.Meta("myprop", "mypropval"),
		oos.MetadataDirective(oos.MetaReplace),
		oos.ObjectDataLocation(oos.BuildObjectSpecifiedDataLocation("ChengDu", true)),
		oos.StorageClass(oos.StorageClassStandard),
	}
	_, err = bucket.CopyObject(objectKey, descObjectKey, options...)
	if err != nil {
		HandleError(err)
	}

	meta, err := bucket.HeadObject(descObjectKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("meta:", meta)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CopyObjectSample completed")
}
