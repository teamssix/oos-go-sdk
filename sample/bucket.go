package sample

import (
	"fmt"

	"oos-go-sdk/oos"
)

// CreateBucketSample shows how to create bucket
func CreateBucketSample() {
	var err error

	bucketNameTest := bucketName

	// New client
	client := NewClient()

	// Case 1: Create a bucket with Local location
	confLocal, err := oos.BuildCreateBucketConfigLocal("ChengDu")
	if err != nil {
		HandleError(err)
	}
	err = client.CreateBucket(bucketNameTest, confLocal)
	if err != nil {
		HandleError(err)
	}

	exist, errHead := client.HeadBucket(bucketNameTest)
	if errHead != nil {
		HandleError(errHead)
	} else if exist {
		fmt.Println("bucket " + bucketNameTest + " exist")
	}

	//Delete bucket
	err = client.DeleteBucket(bucketNameTest)
	if err != nil {
		HandleError(err)
	}

	// Case 2: Create the bucket with ACL And Specified
	bucketNameTest = bucketName
	locationist := make([]string, 0)
	locationist = append(locationist, "ChengDu")
	locationist = append(locationist, "GuiYang")
	conf, err := oos.BuildCreateBucketConfigSpecified("ChengDu", locationist, true)
	if err != nil {
		HandleError(err)
	}

	err = client.CreateBucket(bucketNameTest, conf, oos.ACL(oos.ACLPublicRead))
	if err != nil {
		HandleError(err)
	}

	// Delete bucket
	err = client.DeleteBucket(bucketNameTest)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("CreateBucketSample completed")
}

// ListBucketsSample shows the list bucket, including default and specified parameters.
func ListBucketsSample() {
	// New client
	client := NewClientWithTimeOut(10000, 10000)

	// Case 1: Use default parameter
	lbr, err := client.ListBuckets()
	if err != nil {
		HandleError(err)
	}

	for _, bucket := range lbr.Buckets {
		fmt.Printf("list bucket name :%s CreateData:%s Owner ID:%s\n", bucket.Name, bucket.CreationDate.String(),
			lbr.Owner.ID)
	}
}

func DeleteBucketSample() {
	client := NewClient()
	err := client.DeleteBucket(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("bucket %s deleted ", bucketName)
}
