package oos

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

// CopyFile is multipart copy object
//
// srcBucketName    source bucket name
// srcObjectKey    source object name
// destObjectKey    target object name in the form of bucketname.objectkey
// partSize    the part size in byte.
// options    object's contraints. Check out function InitiateMultipartUpload.
//
// error    it's nil if the operation succeeds, otherwise it's an error object.
//
//	func (bucket Bucket) CopyFile(srcBucketName, srcObjectKey, destObjectKey string, partSize int64, options ...Option) error {
//		destBucketName := bucket.BucketName
//		if partSize < MinPartSize || partSize > MaxPartSize {
//			return errors.New("oos: part size invalid range (1024KB, 5GB]")
//		}
//
//		routines := getRoutines(options)
//
//		return bucket.copyFile(srcBucketName, srcObjectKey, destBucketName, destObjectKey,
//			partSize, options, routines)
//	}
func (bucket Object) CopyObjectAsMultipart(coypSrcList []SrcCopyPartObject, destBucketName, destObjectKey string, options ...Option) error {

	if len(coypSrcList) == 0 {
		return errors.New("the parameter is invalid: coypSrcList size is 0")
	}

	for _, value := range coypSrcList {
		fmt.Println(value)
		if value.BucketName == "" {
			return errors.New("the parameter is invalid: BucketName is empty")
		}
		if value.ObjectName == "" {
			return errors.New("the parameter is invalid: ObjectName is empty")
		}
		if value.PartNumber < 1 {
			return errors.New("the parameter is invalid: PartNumber is error")
		}
	}

	if destBucketName == "" {
		return errors.New("the parameter is invalid: destBucketName is empty")
	}

	if destObjectKey == "" {
		return errors.New("the parameter is invalid: destObjectKey is empty")
	}

	routines := getRoutines(options)

	return bucket.copyObjectAsPartToMutliPart(coypSrcList, destBucketName, destObjectKey,
		options, routines)
}

// ----- Concurrently copy without checkpoint ---------

// copyWorkerArg defines the copy worker arguments
type copyWorkerArg struct {
	bucket  *Object
	imur    InitiateMultipartUploadResult
	options []Option
	hook    copyPartHook
}

// copyPartHook is the hook for testing purpose
type copyPartHook func(part copyPart) error

var copyPartHooker copyPartHook = defaultCopyPartHook

func defaultCopyPartHook(part copyPart) error {
	return nil
}

// copyWorker copies worker
func copyWorker(id int, arg copyWorkerArg, jobs <-chan SrcCopyPartObject, results chan<- UploadPart, failed chan<- error, die <-chan bool) {
	for chunk := range jobs {

		part, err := arg.bucket.UploadPartCopy(arg.imur, chunk.BucketName, chunk.ObjectName, 0, 0, chunk.PartNumber, arg.options...)
		if err != nil {
			failed <- err
			break
		}
		select {
		case <-die:
			return
		default:
		}
		results <- part
	}
}

// copyScheduler
func copyScheduler(jobs chan SrcCopyPartObject, parts []SrcCopyPartObject) {
	for _, part := range parts {
		jobs <- part
	}
	close(jobs)
}

// copyPart structure
type copyPart struct {
	Number int   // Part number (from 1 to 10,000)
	Start  int64 // The start index in the source file.
	End    int64 // The end index in the source file
}

// getCopyParts calculates copy parts
func getCopyParts(objectSize, partSize int64) []copyPart {
	parts := []copyPart{}
	part := copyPart{}
	i := 0
	for offset := int64(0); offset < objectSize; offset += partSize {
		part.Number = i + 1
		part.Start = offset
		part.End = GetPartEnd(offset, objectSize, partSize)
		parts = append(parts, part)
		i++
	}
	return parts
}

// getSrcObjectBytes gets the source file size
func getSrcObjectBytes(parts []copyPart) int64 {
	var ob int64
	for _, part := range parts {
		ob += (part.End - part.Start + 1)
	}
	return ob
}

// copyFile is a concurrently copy without checkpoint
func (bucket Object) copyObjectAsPartToMutliPart(copySrcList []SrcCopyPartObject, destBucketName, destObjectKey string,
	options []Option, routines int) error {
	descBucket, err := bucket.Bucket.Bucket(destBucketName)

	listener := getProgressListener(options)

	payerOptions := []Option{}
	payer := getPayer(options)
	if payer != "" {
		payerOptions = append(payerOptions, RequestPayer(PayerType(payer)))
	}

	// Initialize the multipart upload
	imur, err := descBucket.InitiateMultipartUpload(destObjectKey, options...)
	if err != nil {
		return err
	}

	jobs := make(chan SrcCopyPartObject, len(copySrcList))
	results := make(chan UploadPart, len(copySrcList))
	failed := make(chan error)
	die := make(chan bool)

	event := newProgressEvent(TransferStartedEvent, 0, 0)
	publishProgress(listener, event)

	// Start to copy workers
	arg := copyWorkerArg{descBucket, imur, payerOptions, copyPartHooker}
	for w := 1; w <= routines; w++ {
		go copyWorker(w, arg, jobs, results, failed, die)
	}

	// Start the scheduler
	go copyScheduler(jobs, copySrcList)

	// Wait for the parts finished.
	completed := 0
	ups := make([]UploadPart, len(copySrcList))
	for completed < len(copySrcList) {
		select {
		case part := <-results:
			completed++
			ups[part.PartNumber-1] = part
			event = newProgressEvent(TransferDataEvent, 0, 0)
			publishProgress(listener, event)
		case err := <-failed:
			close(die)
			descBucket.AbortMultipartUpload(imur, payerOptions...)
			event = newProgressEvent(TransferFailedEvent, 0, 0)
			publishProgress(listener, event)
			return err
		}

		if completed >= len(copySrcList) {
			break
		}
	}

	event = newProgressEvent(TransferCompletedEvent, 0, 0)
	publishProgress(listener, event)

	// Complete the multipart upload
	_, err = descBucket.CompleteMultipartUpload(imur, ups, payerOptions...)
	if err != nil {
		bucket.AbortMultipartUpload(imur, payerOptions...)
		return err
	}
	return nil
}

// ----- Concurrently copy with checkpoint  -----

const copyCpMagic = "84F1F18C-FF1D-403B-A1D8-9DEB5F65910A"

type copyCheckpoint struct {
	Magic          string       // Magic
	MD5            string       // CP content MD5
	SrcBucketName  string       // Source bucket
	SrcObjectKey   string       // Source object
	DestBucketName string       // Target bucket
	DestObjectKey  string       // Target object
	CopyID         string       // Copy ID
	ObjStat        objectStat   // Object stat
	Parts          []copyPart   // Copy parts
	CopyParts      []UploadPart // The uploaded parts
	PartStat       []bool       // The part status
}

// isValid checks if the data is valid which means CP is valid and object is not updated.
func (cp copyCheckpoint) isValid(meta http.Header) (bool, error) {
	// Compare CP's magic number and the MD5.
	cpb := cp
	cpb.MD5 = ""
	js, _ := json.Marshal(cpb)
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])

	if cp.Magic != downloadCpMagic || b64 != cp.MD5 {
		return false, nil
	}

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return false, err
	}

	// Compare the object size and last modified time and etag.
	if cp.ObjStat.Size != objectSize ||
		cp.ObjStat.LastModified != meta.Get(HTTPHeaderLastModified) ||
		cp.ObjStat.Etag != meta.Get(HTTPHeaderEtag) {
		return false, nil
	}

	return true, nil
}

// load loads from the checkpoint file
func (cp *copyCheckpoint) load(filePath string) error {
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, cp)
	return err
}

// update updates the parts status
func (cp *copyCheckpoint) update(part UploadPart) {
	cp.CopyParts[part.PartNumber-1] = part
	cp.PartStat[part.PartNumber-1] = true
}

// dump dumps the CP to the file
func (cp *copyCheckpoint) dump(filePath string) error {
	bcp := *cp

	// Calculate MD5
	bcp.MD5 = ""
	js, err := json.Marshal(bcp)
	if err != nil {
		return err
	}
	sum := md5.Sum(js)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	bcp.MD5 = b64

	// Serialization
	js, err = json.Marshal(bcp)
	if err != nil {
		return err
	}

	// Dump
	return ioutil.WriteFile(filePath, js, FilePermMode)
}

// todoParts returns unfinished parts
func (cp copyCheckpoint) todoParts() []copyPart {
	dps := []copyPart{}
	for i, ps := range cp.PartStat {
		if !ps {
			dps = append(dps, cp.Parts[i])
		}
	}
	return dps
}

// getCompletedBytes returns finished bytes count
func (cp copyCheckpoint) getCompletedBytes() int64 {
	var completedBytes int64
	for i, part := range cp.Parts {
		if cp.PartStat[i] {
			completedBytes += (part.End - part.Start + 1)
		}
	}
	return completedBytes
}

// prepare initializes the multipart upload
func (cp *copyCheckpoint) prepare(meta http.Header, srcBucket *Object, srcObjectKey string, destBucket *Object, destObjectKey string,
	partSize int64, options []Option) error {
	// CP
	cp.Magic = copyCpMagic
	cp.SrcBucketName = srcBucket.BucketName
	cp.SrcObjectKey = srcObjectKey
	cp.DestBucketName = destBucket.BucketName
	cp.DestObjectKey = destObjectKey

	objectSize, err := strconv.ParseInt(meta.Get(HTTPHeaderContentLength), 10, 0)
	if err != nil {
		return err
	}

	cp.ObjStat.Size = objectSize
	cp.ObjStat.LastModified = meta.Get(HTTPHeaderLastModified)
	cp.ObjStat.Etag = meta.Get(HTTPHeaderEtag)

	// Parts
	cp.Parts = getCopyParts(objectSize, partSize)
	cp.PartStat = make([]bool, len(cp.Parts))
	for i := range cp.PartStat {
		cp.PartStat[i] = false
	}
	cp.CopyParts = make([]UploadPart, len(cp.Parts))

	// Init copy
	imur, err := destBucket.InitiateMultipartUpload(destObjectKey, options...)
	if err != nil {
		return err
	}
	cp.CopyID = imur.UploadID

	return nil
}

func (cp *copyCheckpoint) complete(bucket *Object, parts []UploadPart, cpFilePath string, options []Option) error {
	imur := InitiateMultipartUploadResult{Bucket: cp.DestBucketName,
		Key: cp.DestObjectKey, UploadID: cp.CopyID}
	_, err := bucket.CompleteMultipartUpload(imur, parts, options...)
	if err != nil {
		return err
	}
	os.Remove(cpFilePath)
	return err
}
