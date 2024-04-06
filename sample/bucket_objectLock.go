package sample

import (
	"fmt"

	"oos-go-sdk/oos"
)

// "6.2.0/oos"
func BucketObjectLockSample() {
	client := NewClient()
	// put
	bucketObjectLock := oos.BuildBucketObjectLockByDays(string(oos.BucketModeCompliance), 10000, false)
	err := client.SetBucketObjectLock(bucketName, bucketObjectLock)
	if err != nil {
		HandleError(err)
	}

	bucketObjectLock = oos.BuildBucketObjectLockByYears(string(oos.BucketModeCompliance), 10, false)
	err = client.SetBucketObjectLock(bucketName, bucketObjectLock)
	if err != nil {
		HandleError(err)
	}

	//get
	out, err := client.GetBucketObjectLock(bucketName)
	if err != nil {
		HandleError(err)
	}
	if 0 != out.DefaultRetention.Days {
		fmt.Println(out.DefaultRetention.Days)
	}
	if 0 != out.DefaultRetention.Years {
		fmt.Println(out.DefaultRetention.Years)
	}

	//delete
	err = client.DeleteBucketObjectLock(bucketName)
	if err != nil {
		HandleError(err)
	}
}
