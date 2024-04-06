package sample

import (
	"bytes"
	"fmt"
	"os"

	"oos-go-sdk/oos"
)

func readFile(path string, offset int64, size int64) *bytes.Buffer {
	fd, err := os.Open(path)
	if err != nil {
		HandleError(err)
	}
	defer fd.Close()
	fd.Seek(offset, os.SEEK_SET)
	data := make([]byte, size)
	len, _ := fd.Read(data)
	fmt.Print("read :")
	fmt.Println(len)
	return bytes.NewBuffer(data)
}

func StepMultipartSample() {

	bucket, err := GetTestBucket(bucketName)
	if err != nil {
		HandleError(err)
	}

	imur, err := bucket.InitiateMultipartUpload(objectKeyMultipart, oos.ObjectDataLocation(oos.BuildObjectSpecifiedDataLocation("ChengDu", true)))
	if err != nil {
		HandleError(err)
	}
	err = bucket.AbortMultipartUpload(imur)
	if err != nil {
		HandleError(err)
	}

	imur, err = bucket.InitiateMultipartUpload(objectKeyMultipart, oos.StorageClass(oos.StorageClassStandard))
	if err != nil {
		HandleError(err)
	}

	var partSize int64 = 5 * 1024 * 1024
	chunks, err := oos.SplitFileByPartSize(localFileMultipart, partSize)
	if err != nil {
		HandleError(err)
	}
	uploadParts := make([]oos.UploadPart, 0)
	for index, chunk := range chunks {
		data := readFile(localFileMultipart, chunk.Offset, chunk.Size)
		uploadPart, err := bucket.UploadPart(imur, data, chunk.Size, index+1)
		if err != nil {
			HandleError(err)
		}
		uploadParts = append(uploadParts, uploadPart)

	}

	lmurs, err := bucket.ListMultipartUploads()
	if err != nil {
		HandleError(err)
	}

	// list part test
	upload := lmurs.Uploads[0]
	imurTemp := oos.InitiateMultipartUploadResult{Bucket: bucket.BucketName,
		Key: upload.Key, UploadID: upload.UploadID}
	lpr, err := bucket.ListUploadedParts(imurTemp)
	if err != nil {
		HandleError(err)
	}
	fmt.Printf("list part result %v", lpr)

	/*
		for _, upload := range lmurs.Uploads {
			var imur = oos.InitiateMultipartUploadResult{Bucket: bucket.BucketName,
				Key: upload.Key, UploadID: upload.UploadID}
			err = bucket.AbortMultipartUpload(imur)
			if err != nil {
				HandleError(err)
			}
		}*/

	compleRet, err := bucket.CompleteMultipartUpload(imur, uploadParts)
	if err != nil {
		HandleError(err)
	}
	fmt.Println(compleRet)

	// Delete object and bucket
	err = DeleteTestBucketAndObject(bucketName)
	if err != nil {
		HandleError(err)
	}

}
