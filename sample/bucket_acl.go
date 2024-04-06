package sample

import (
	"fmt"
	"strconv"

	"oos-go-sdk/oos"
)

// BucketACLSample shows how to get and set the bucket ACL
func BucketACLSample() {
	// New client
	client := NewClient()

	//set bucket ACL pubilcRead
	err := client.SetBucketACL(bucketName, oos.ACLPublicRead)
	if err != nil {
		HandleError(err)
	}

	// Get bucket ACL
	gbar, err := client.GetBucketACL(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("ID:" + gbar.Owner.ID)
	fmt.Println("DisplayName:" + gbar.Owner.DisplayName)
	for i, grant := range gbar.GrantList {
		fmt.Println("Index:" + strconv.Itoa(i))
		fmt.Println("GranteeURI:" + grant.GranteeURI)
		fmt.Println("Permission:" + grant.Permission)
	}

	fmt.Println("BucketACLSample completed")
}
