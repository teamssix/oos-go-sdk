package sample

import (
	"fmt"
	"oos-go-sdk/oos"
	"time"
)

func BucketCorsSample() {
	var err error

	// New client
	client := NewClient()

	//1 put cors with all
	rule := oos.CORSRule{}
	rule.AllowedOrigin = []string{"http://www.example1.com", "http://ctyun.cn"}
	rule.AllowedMethod = []string{"PUT", "POST", "DELETE"}
	rule.AllowedHeader = []string{"*", "x-amz-*"}
	rule.ExposeHeader = []string{"JavaScript XMLHttpRequest"}
	rule.MaxAgeSeconds = 1000
	corsRules := []oos.CORSRule{rule}

	err = client.SetBucketCors(bucketName, corsRules)
	if err != nil {
		HandleError(err)
	}

	//2 Creating cors does not contain AllowedHeader
	rule1 := oos.CORSRule{}
	rule1.AllowedOrigin = []string{"http://www.example1.com", "http://ctyun.cn"}
	rule1.AllowedMethod = []string{"PUT", "POST", "DELETE"}
	rule1.ExposeHeader = []string{"JavaScript XMLHttpRequest"}
	rule1.MaxAgeSeconds = 1000
	corsRules1 := []oos.CORSRule{rule1}

	err = client.SetBucketCors(bucketName, corsRules1)
	if err != nil {
		HandleError(err)
	}

	//3 Creating cors does not contain ExposeHeader
	rule2 := oos.CORSRule{}
	rule2.AllowedOrigin = []string{"http://www.example1.com", "http://ctyun.cn"}
	rule2.AllowedMethod = []string{"PUT", "POST", "DELETE"}
	rule2.AllowedHeader = []string{"*", "x-amz-*"}
	rule2.MaxAgeSeconds = 1000

	corsRules2 := []oos.CORSRule{rule2}

	err = client.SetBucketCors(bucketName, corsRules2)
	if err != nil {
		HandleError(err)
	}

	// 4 get cors  注意：因为oosbucket有缓存，所以设置了cors后立即进行get 有404 的可能。 缓存在至少5分钟后才全部刷新。
	time.Sleep(time.Duration(5) * time.Minute)
	cors, err := client.GetBucketCors(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println(cors)

	//5 delete cors  同get， 删除后再get 是有可能再get到的。
	err = client.DeleteBucketCors(bucketName)
	if err != nil {
		HandleError(err)
	}

	fmt.Println("BucketCorsSample completed")
}
