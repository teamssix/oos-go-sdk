package sample

import (
	"strconv"

	"oos-go-sdk/oos"
)

func CopyPartMultipartSample() {

	// Create a bucket client
	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	descObjectKey := "descobject"

	var partSize int64 = 5 * 1024 * 1024
	chunks, err := oos.SplitFileByPartSize(localFileMultipart, partSize)
	if err != nil {
		HandleError(err)
	}

	srcCopyPartObjectList := make([]oos.SrcCopyPartObject, 0)
	for index, chunk := range chunks {
		data := readFile(localFileMultipart, chunk.Offset, chunk.Size)
		objectName := "copy-multi-test" + strconv.Itoa(index)
		err := bucket.PutObject(objectName, data)
		if err != nil {
			HandleError(err)
		}
		temp := oos.SrcCopyPartObject{BucketName: bucketName, ObjectName: objectName, PartNumber: index + 1}
		srcCopyPartObjectList = append(srcCopyPartObjectList, temp)
	}

	err = bucket.CopyObjectAsMultipart(srcCopyPartObjectList, bucketName, descObjectKey)
	if err != nil {
		HandleError(err)
	}

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}
}
