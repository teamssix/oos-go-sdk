package sample

import (
	"fmt"
)

func AccessKeySample() {
	// New client
	client := NewIAMClient()

	getAKlastUsedRet, err := client.GetAccessKeyLastUsed(accessKey)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("get access key last used %v\n", getAKlastUsedRet)

	// Create Accesskey
	createRet, err := client.CreateAccessKey(userName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("access key info %v\n", createRet)

	// List Accesskey
	listRet, err := client.ListAccessKey(5, "", userName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("access key %v\n", listRet)

	//This parameter is available only for keys created after IAM goes online
	if nil != listRet.ListAccessKeysResult.MemberList[0].CreateDate {
		fmt.Println("create time ->", listRet.ListAccessKeysResult.MemberList[0].CreateDate.String())
	}

	// Delete AccessKey
	AccessKeyId := createRet.CreateAccessKeyResult.AcessKey.AccessKeyId
	deleteRet, err := client.DeleteAccessKey(AccessKeyId, userName)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("access key %v\n", deleteRet)
}
