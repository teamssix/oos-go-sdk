package sample

import (
	"fmt"
	"oos-go-sdk/oos"
)

// BucketWebSiteSample shows how to get and set the bucket WebSite
func BucketWebSiteSample() {
	var err error

	// New client
	client := NewClient()

	// case 1 website all request to
	config := oos.WebsiteConfiguration{}
	config.WebsiteAllRequestTo = &oos.WebsiteAllRequestToXML{HostName: "www.example.com", Protocol: "http"}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	// case 2 website all request to no protocol
	config = oos.WebsiteConfiguration{}
	config.WebsiteAllRequestTo = &oos.WebsiteAllRequestToXML{HostName: "www.example.com"}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	//case 3 all exist
	config = oos.WebsiteConfiguration{}
	config.IndexDocument.Suffix = "index.html"
	config.ErrorDocument.Key = "error.html"
	condition := oos.Condition{HttpErrorCodeReturnedEquals: "403", KeyPrefixEquals: "docs"}
	redirect := oos.Redirect{HostName: "www.example.com", Protocol: "http", ReplaceKeyPrefixWith: "documents/"}
	rule := oos.RoutingRule{Condition: &condition, Redirect: &redirect}
	config.RoutingRules = []oos.RoutingRule{rule}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	//case 4 no condition
	config = oos.WebsiteConfiguration{}
	config.IndexDocument.Suffix = "index.html"
	config.ErrorDocument.Key = "error.html"
	redirect = oos.Redirect{HostName: "www.example.com", Protocol: "http", ReplaceKeyWith: "errorpage.html"}
	rule = oos.RoutingRule{Condition: nil, Redirect: &redirect}
	config.RoutingRules = []oos.RoutingRule{rule}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	//case 5 all with out redirect > protocol
	config = oos.WebsiteConfiguration{}
	config.IndexDocument.Suffix = "index.html"
	config.ErrorDocument.Key = "error.html"
	condition = oos.Condition{HttpErrorCodeReturnedEquals: "404", KeyPrefixEquals: "docs"}
	redirect = oos.Redirect{HostName: "www.example.com", ReplaceKeyWith: "errorpage.html"}
	rule = oos.RoutingRule{Condition: &condition, Redirect: &redirect}
	config.RoutingRules = []oos.RoutingRule{rule}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	//case 6 without condition > HttpErrorCodeReturnedEquals
	config = oos.WebsiteConfiguration{}
	config.IndexDocument.Suffix = "index.html"
	config.ErrorDocument.Key = "error.html"
	condition = oos.Condition{KeyPrefixEquals: "docs"}
	redirect = oos.Redirect{HostName: "www.example.com", Protocol: "http", ReplaceKeyPrefixWith: "documents/"}
	rule = oos.RoutingRule{Condition: &condition, Redirect: &redirect}
	config.RoutingRules = []oos.RoutingRule{rule}
	err = client.SetBucketWebsite(bucketName, config)
	if err != nil {
		HandleError(err)
	}

	webSite, err := client.GetBucketWebsite(bucketName)
	if err != nil {
		HandleError(err)
	}
	fmt.Println("website info: \n IndexDocument:" + webSite.IndexDocument.Suffix + "\n ErrorDocument:" + webSite.ErrorDocument.Key)
	for _, s := range *webSite.RoutingRules {
		fmt.Println(s.Redirect)
		fmt.Println(s.Condition)
	}

	err = client.DeleteBucketWebsite(bucketName)
	if err != nil {
		HandleError(err)
	}

}
