// main of samples

package main

import (
	"fmt"
	"github.com/teamssix/oos-go-sdk/sample"
)

func main() {

	/*************** bucket test *******************/
	sample.CreateBucketSample()
	sample.GetBucketLocation()
	sample.BucketACLSample()
	sample.DeleteBucketSample()
	sample.BucketPolicySample()
	sample.BucketWebSiteSample()
	sample.BucketLoggingSample()
	sample.BucketLifecycleSample()
	sample.BucketCorsSample()
	sample.BucketObjectLockSample()

	/*************** AccessKey test *******************/
	sample.AccessKeySample() // 6版本 只支持 https类型的endpoint 只支持V4签名

	/*************** object test ***************/
	sample.PutObjectSample()
	sample.DeleteObjectSample()
	sample.ListObjectsSample()
	sample.GetObjectSample()
	sample.CopyObjectSample()
	sample.SignURLSample()

	/*************** object multipart test ***************/
	sample.StepMultipartSample()
	sample.PutObjectMultipartSample()
	sample.GetObjectMultipartSample()
	sample.CopyPartMultipartSample()

	/*************** service test ***************/
	sample.ListBucketsSample()
	sample.GetRegionSample()

	fmt.Println("All samples completed")
}
