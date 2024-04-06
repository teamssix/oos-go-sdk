package sample

import (
	"fmt"
	"oos-go-sdk/oos"
)

func GetBucketLocation() {
	// New client
	client := NewClient()
	ret, err := client.GetBucketLocation(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println(ret.DataLocationType == oos.DataLocationTypeSpecified)
	fmt.Println(ret)
}
