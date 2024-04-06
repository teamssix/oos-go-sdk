package oos

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Object implements the operations of object.
type Object struct {
	Bucket     Client
	BucketName string
}

// PutObject creates a new object and it will overwrite the original one if it exists already.
//
// objectKey    the object key in UTF-8 encoding. The length must be between 1 and 1023, and cannot start with "/" or "\".
// reader    io.Reader instance for reading the data for uploading
// options    the options for uploading the object. The valid options here are CacheControl, ContentDisposition, ContentEncoding
//
//	Expires, ServerSideEncryption, ObjectACL and Meta.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) PutObject(objectKey string, reader io.Reader, options ...Option) error {

	if objectKey == "" {
		return errors.New("the parameter is invalid: objectKey is empty")
	}

	var ContentLength int64
	switch v := reader.(type) {
	case *bytes.Buffer:
		ContentLength = int64(v.Len())
	case *bytes.Reader:
		ContentLength = int64(v.Len())
	case *strings.Reader:
		ContentLength = int64(v.Len())
	case *os.File:
		ContentLength = tryGetFileSize(v)
	case *io.LimitedReader:
		ContentLength = int64(v.N)
	}

	if ContentLength == 0 {
		return errors.New("the parameter is invalid: Object's content is empty")
	}

	request := &PutObjectRequest{
		ObjectKey: objectKey,
		Reader:    reader,
	}
	resp, err := bucket.DoPutObject(request, options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// PutObjectFromFile creates a new object from the local file.
//
// objectKey    object key.
// filePath    the local file path to upload.
// options    the options for uploading the object. Refer to the parameter options in PutObject for more details.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) PutObjectFromFile(objectKey, filePath string, options ...Option) error {

	if objectKey == "" {
		return errors.New("the parameter is invalid: ObjectKey is empty")
	}

	if filePath == "" {
		return errors.New("the parameter is invalid: filePath  is empty")
	}

	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	opts := addContentType(options, filePath, objectKey)

	request := &PutObjectRequest{
		ObjectKey: objectKey,
		Reader:    fd,
	}
	resp, err := bucket.DoPutObject(request, opts)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// DoPutObject does the actual upload work.
//
// request    the request instance for uploading an object.
// options    the options for uploading an object.
//
// Response    the response from oos.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DoPutObject(request *PutObjectRequest, options []Option) (*Response, error) {
	isOptSet, _, _ := isOptionSet(options, HTTPHeaderContentType)
	if !isOptSet {
		options = addContentType(options, request.ObjectKey)
	}

	listener := getProgressListener(options)

	params := map[string]interface{}{}
	resp, err := bucket.do("PUT", request.ObjectKey, params, options, request.Reader, listener)
	if err != nil {
		return nil, err
	}

	err = checkRespCode(resp.StatusCode, []int{http.StatusOK})

	return resp, err
}

// GetObject downloads the object.
//
// objectKey    the object key.
// options    the options for downloading the object. The valid values are: Range, IfModifiedSince, IfUnmodifiedSince, IfMatch,
//
//	IfNoneMatch, AcceptEncoding.
//
// io.ReadCloser    reader instance for reading data from response. It must be called close() after the usage and only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) GetObject(objectKey string, options ...Option) (io.ReadCloser, error) {

	if objectKey == "" {
		return nil, errors.New("the parameter is invalid: ObjectKey is empty")
	}

	result, err := bucket.DoGetObject(&GetObjectRequest{objectKey}, options)
	if err != nil {
		return nil, err
	}

	return result.Response, nil
}

// GetObjectToFile downloads the data to a local file.
//
// objectKey    the object key to download.
// filePath    the local file to store the object data.
// options    the options for downloading the object. Refer to the parameter options in method GetObject for more details.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) GetObjectToFile(objectKey, filePath string, options ...Option) error {

	if objectKey == "" {
		return errors.New("the parameter is invalid: ObjectKey is empty")
	}

	if filePath == "" {
		return errors.New("the parameter is invalid: filePath is empty")
	}

	tempFilePath := filePath + TempFileSuffix

	// Calls the API to actually download the object. Returns the result instance.
	result, err := bucket.DoGetObject(&GetObjectRequest{objectKey}, options)
	if err != nil {
		return err
	}
	defer result.Response.Close()

	// If the local file does not exist, create a new one. If it exists, overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, FilePermMode)
	if err != nil {
		return err
	}

	// Copy the data to the local file path.
	_, err = io.Copy(fd, result.Response.Body)
	fd.Close()
	if err != nil {
		return err
	}

	return os.Rename(tempFilePath, filePath)
}

// DoGetObject is the actual API that gets the object. It's the internal function called by other public APIs.
//
// request    the request to download the object.
// options    the options for downloading the file. Checks out the parameter options in method GetObject.
//
// GetObjectResult    the result instance of getting the object.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DoGetObject(request *GetObjectRequest, options []Option) (*GetObjectResult, error) {
	params, _ := getRawParams(options)
	resp, err := bucket.do("GET", request.ObjectKey, params, options, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &GetObjectResult{
		Response: resp,
	}

	// Progress
	listener := getProgressListener(options)

	contentLen, _ := strconv.ParseInt(resp.Headers.Get(HTTPHeaderContentLength), 10, 64)
	resp.Body = TeeReader(resp.Body, nil, contentLen, listener, nil)

	return result, nil
}

// CopyObject copies the object inside the bucket.
//
// srcObjectKey    the source object to copy.
// destObjectKey    the target object to copy.
// options    options for copying an object. You can specify the conditions of copy. The valid conditions are CopySourceIfMatch,
//
//	CopySourceIfNoneMatch, CopySourceIfModifiedSince, CopySourceIfUnmodifiedSince, MetadataDirective.
//	Also you can specify the target object's attributes, such as CacheControl, ContentDisposition, ContentEncoding, Expires,
//	, ObjectACL, Meta. s
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) CopyObject(srcObjectKey, destObjectKey string, options ...Option) (CopyObjectResult, error) {
	var out CopyObjectResult

	if destObjectKey == "" {
		return out, errors.New("the parameter is invalid: srcObjectKey is empty")
	}

	if srcObjectKey == "" {
		return out, errors.New("the parameter is invalid: destObjectKey is empty")
	}

	options = append(options, CopySource(bucket.BucketName, url.QueryEscape(srcObjectKey)))
	params := map[string]interface{}{}
	resp, err := bucket.do("PUT", destObjectKey, params, options, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// CopyObjectTo copies the object to another bucket.
//
// srcObjectKey    source object key. The source bucket is Bucket.BucketName .
// destBucketName    target bucket name.
// destObjectKey    target object name.
// options    copy options, check out parameter options in function CopyObject for more details.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) CopyObjectTo(destBucketName, destObjectKey, srcObjectKey string, options ...Option) (CopyObjectResult, error) {
	return bucket.copy(srcObjectKey, destBucketName, destObjectKey, options...)
}

// CopyObjectFrom copies the object to another bucket.
//
// srcBucketName    source bucket name.
// srcObjectKey    source object name.
// destObjectKey    target object name. The target bucket name is Bucket.BucketName.
// options    copy options. Check out parameter options in function CopyObject.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) CopyObjectFrom(srcBucketName, srcObjectKey, destObjectKey string, options ...Option) (CopyObjectResult, error) {
	destBucketName := bucket.BucketName
	var out CopyObjectResult
	srcBucket, err := bucket.Bucket.Bucket(srcBucketName)
	if err != nil {
		return out, err
	}

	return srcBucket.copy(srcObjectKey, destBucketName, destObjectKey, options...)
}

func (bucket Object) copy(srcObjectKey, destBucketName, destObjectKey string, options ...Option) (CopyObjectResult, error) {
	var out CopyObjectResult
	options = append(options, CopySource(bucket.BucketName, url.QueryEscape(srcObjectKey)))
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return out, err
	}
	params := map[string]interface{}{}
	resp, err := bucket.Bucket.Conn.Do("PUT", destBucketName, destObjectKey, params, headers, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// DeleteObject deletes the object.
//
// objectKey    the object key to delete.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DeleteObject(objectKey string) error {

	if objectKey == "" {
		return errors.New("the parameter is invalid: ObjectKey is empty")
	}

	params := map[string]interface{}{}
	resp, err := bucket.do("DELETE", objectKey, params, nil, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusNoContent})
}

// DeleteObjects deletes multiple objects.
//
// objectKeys    the object keys to delete.
// options    the options for deleting objects.
//
//	Supported option is DeleteObjectsQuiet which means it will not return error even deletion failed (not recommended). By default it's not used.
//
// DeleteObjectsResult    the result object.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DeleteObjects(objectKeys []string, options ...Option) (DeleteObjectsResult, error) {
	out := DeleteObjectsResult{}
	dxml := deleteXML{}
	for _, key := range objectKeys {
		dxml.Objects = append(dxml.Objects, DeleteObject{Key: key})
	}
	isQuiet, _ := findOption(options, deleteObjectsQuiet, false)
	dxml.Quiet = isQuiet.(bool)

	bs, err := xml.Marshal(dxml)
	if err != nil {
		return out, err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	sum := md5.Sum(bs)
	b64 := base64.StdEncoding.EncodeToString(sum[:])
	options = append(options, ContentMD5(b64))

	params := map[string]interface{}{}
	params["delete"] = nil

	resp, err := bucket.do("POST", "", params, options, buffer, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	if !dxml.Quiet {
		err = xmlUnmarshal(resp.Body, &out)
	}
	return out, err
}

// IsObjectExist checks if the object exists.
//
// bool    flag of object's existence (true:exists; false:non-exist) when error is nil.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) IsObjectExist(objectKey string) (bool, error) {

	if objectKey == "" {
		return false, errors.New("the parameter is invalid: ObjectKey is empty")
	}

	_, err := bucket.GetObjectMeta(objectKey)
	if err == nil {
		return true, nil
	}

	switch err.(type) {
	case ServiceError:
		if err.(ServiceError).StatusCode == 404 && err.(ServiceError).Code == "NoSuchKey" {
			return false, nil
		}
	}

	return false, err
}

// ListObjects lists the objects under the current bucket.
//
// options    it contains all the filters for listing objects.
//
//	It could specify a prefix filter on object keys,  the max keys count to return and the object key marker and the delimiter for grouping object names.
//	The key marker means the returned objects' key must be greater than it in lexicographic order.
//
//	For example, if the bucket has 8 objects, my-object-1, my-object-11, my-object-2, my-object-21,
//	my-object-22, my-object-3, my-object-31, my-object-32. If the prefix is my-object-2 (no other filters), then it returns
//	my-object-2, my-object-21, my-object-22 three objects. If the marker is my-object-22 (no other filters), then it returns
//	my-object-3, my-object-31, my-object-32 three objects. If the max keys is 5, then it returns 5 objects.
//	The three filters could be used together to achieve filter and paging functionality.
//	If the prefix is the folder name, then it could list all files under this folder (including the files under its subfolders).
//	But if the delimiter is specified with '/', then it only returns that folder's files (no subfolder's files). The direct subfolders are in the commonPrefixes properties.
//	For example, if the bucket has three objects fun/test.jpg, fun/movie/001.avi, fun/movie/007.avi. And if the prefix is "fun/", then it returns all three objects.
//	But if the delimiter is '/', then only "fun/test.jpg" is returned as files and fun/movie/ is returned as common prefix.
//
//	For common usage scenario, check out sample/list_object.go.
//
// ListObjectsResponse    the return value after operation succeeds (only valid when error is nil).
func (bucket Object) ListObjects(options ...Option) (ListObjectsResult, error) {
	var out ListObjectsResult

	options = append(options, EncodingType(HTTPParamEncodingType))
	params, err := getRawParams(options)
	if err != nil {
		return out, err
	}

	resp, err := bucket.do("GET", "", params, options, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	if err != nil {
		return out, err
	}

	err = decodeListObjectsResult(&out)
	return out, err
}

// SetObjectMeta sets the metadata of the Object.
//
// objectKey    object
// options    options for setting the metadata. The valid options are CacheControl, ContentDisposition, ContentEncoding, Expires,
//
//	ServerSideEncryption, and custom metadata.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) SetObjectMeta(objectKey string, options ...Option) error {
	options = append(options, MetadataDirective(MetaReplace))
	_, err := bucket.CopyObject(objectKey, objectKey, options...)
	return err
}

// HeadObject gets the object's detailed metadata
//
// objectKey    object key.
// options    the constraints of the object. Only when the object meets the requirements this method will return the metadata. Otherwise returns error. Valid options are IfModifiedSince, IfUnmodifiedSince,
//
//	IfMatch, IfNoneMatch.
//
// http.Header    object meta when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) HeadObject(objectKey string, options ...Option) (http.Header, error) {

	if objectKey == "" {
		return nil, errors.New("the parameter is invalid: ObjectKey is empty")
	}

	params := map[string]interface{}{}
	resp, err := bucket.do("HEAD", objectKey, params, options, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Headers, nil
}

// GetObjectMeta gets object metadata.
//
// GetObjectMeta is more lightweight than HeadObject as it only returns basic metadata including ETag
// size, LastModified. The size information is in the HTTP header Content-Length.
//
// objectKey    object key
//
// http.Header    the object's metadata, valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) GetObjectMeta(objectKey string, options ...Option) (http.Header, error) {
	params := map[string]interface{}{}
	params["objectMeta"] = nil
	//resp, err := bucket.do("GET", objectKey, "?objectMeta", "", nil, nil, nil)
	resp, err := bucket.do("GET", objectKey, params, options, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return resp.Headers, nil
}

// SignURL signs the URL. Users could access the object directly with this URL without getting the AK.
//
// objectKey    the target object to sign.
// signURLConfig    the configuration for the signed URL
//
// string    returns the signed URL, when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) SignURL(objectKey string, method HTTPMethod, expiredInSec int64, options ...Option) (string, error) {
	if expiredInSec < 0 {
		return "", fmt.Errorf("invalid expires: %d, expires must bigger than 0", expiredInSec)
	}

	params, err := getRawParams(options)
	if err != nil {
		return "", err
	}

	headers := make(map[string]string)
	err = handleOptions(headers, options)
	if err != nil {
		return "", err
	}

	return bucket.Bucket.Conn.signURL(method, bucket.BucketName, objectKey, expiredInSec, params, headers), nil
}

// PutObjectWithURL uploads an object with the URL. If the object exists, it will be overwritten.
// PutObjectWithURL It will not generate minetype according to the key name.
//
// signedURL    signed URL.
// reader    io.Reader the read instance for reading the data for the upload.
// options    the options for uploading the data. The valid options are CacheControl, ContentDisposition, ContentEncoding,
//
//	Expires, ServerSideEncryption, ObjectACL and custom metadata.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) PutObjectWithURL(signedURL string, reader io.Reader, options ...Option) error {
	resp, err := bucket.DoPutObjectWithURL(signedURL, reader, options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// PutObjectFromFileWithURL uploads an object from a local file with the signed URL.
// PutObjectFromFileWithURL It does not generate mimetype according to object key's name or the local file name.
//
// signedURL    the signed URL.
// filePath    local file path, such as dirfile.txt, for uploading.
// options    options for uploading, same as the options in PutObject function.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) PutObjectFromFileWithURL(signedURL, filePath string, options ...Option) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fd.Close()

	resp, err := bucket.DoPutObjectWithURL(signedURL, fd, options)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

// DoPutObjectWithURL is the actual API that does the upload with URL work(internal for SDK)
//
// signedURL    the signed URL.
// reader    io.Reader the read instance for getting the data to upload.
// options    options for uploading.
//
// Response    the response object which contains the HTTP response.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DoPutObjectWithURL(signedURL string, reader io.Reader, options []Option) (*Response, error) {
	listener := getProgressListener(options)

	params := map[string]interface{}{}
	resp, err := bucket.doURL("PUT", signedURL, params, options, reader, listener)
	if err != nil {
		return nil, err
	}

	err = checkRespCode(resp.StatusCode, []int{http.StatusOK})

	return resp, err
}

// GetObjectWithURL downloads the object and returns the reader instance,  with the signed URL.
//
// signedURL    the signed URL.
// options    options for downloading the object. Valid options are IfModifiedSince, IfUnmodifiedSince, IfMatch,
//
//	IfNoneMatch, AcceptEncoding.
//
// io.ReadCloser    the reader object for getting the data from response. It needs be closed after the usage. It's only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) GetObjectWithURL(signedURL string, options ...Option) (io.ReadCloser, error) {
	result, err := bucket.DoGetObjectWithURL(signedURL, options)
	if err != nil {
		return nil, err
	}
	return result.Response, nil
}

// GetObjectToFileWithURL downloads the object into a local file with the signed URL.
//
// signedURL    the signed URL
// filePath    the local file path to download to.
// options    the options for downloading object. Check out the parameter options in function GetObject for the reference.
//
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) GetObjectToFileWithURL(signedURL, filePath string, options ...Option) error {
	tempFilePath := filePath + TempFileSuffix

	// Get the object's content
	result, err := bucket.DoGetObjectWithURL(signedURL, options)
	if err != nil {
		return err
	}
	defer result.Response.Close()

	// If the file does not exist, create one. If exists, then overwrite it.
	fd, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, FilePermMode)
	if err != nil {
		return err
	}

	// Save the data to the file.
	_, err = io.Copy(fd, result.Response.Body)
	fd.Close()
	if err != nil {
		return err
	}

	return os.Rename(tempFilePath, filePath)
}

// DoGetObjectWithURL is the actual API that downloads the file with the signed URL.
//
// signedURL    the signed URL.
// options    the options for getting object. Check out parameter options in GetObject for the reference.
//
// GetObjectResult    the result object when the error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (bucket Object) DoGetObjectWithURL(signedURL string, options []Option) (*GetObjectResult, error) {
	params, _ := getRawParams(options)
	resp, err := bucket.doURL("GET", signedURL, params, options, nil, nil)
	if err != nil {
		return nil, err
	}

	result := &GetObjectResult{
		Response: resp,
	}

	// Progress
	listener := getProgressListener(options)

	contentLen, _ := strconv.ParseInt(resp.Headers.Get(HTTPHeaderContentLength), 10, 64)
	resp.Body = TeeReader(resp.Body, nil, contentLen, listener, nil)

	return result, nil
}

// Private
func (bucket Object) do(method, objectName string, params map[string]interface{}, options []Option,
	data io.Reader, listener ProgressListener) (*Response, error) {
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return nil, err
	}
	return bucket.Bucket.Conn.Do(method, bucket.BucketName, objectName,
		params, headers, data, listener)
}

func (bucket Object) doURL(method HTTPMethod, signedURL string, params map[string]interface{}, options []Option,
	data io.Reader, listener ProgressListener) (*Response, error) {
	headers := make(map[string]string)
	err := handleOptions(headers, options)
	if err != nil {
		return nil, err
	}
	return bucket.Bucket.Conn.DoURL(method, signedURL, headers, data, 0, listener)
}

func addContentType(options []Option, keys ...string) []Option {
	typ := TypeByExtension("")
	for _, key := range keys {
		typ = TypeByExtension(key)
		if typ != "" {
			break
		}
	}

	if typ == "" {
		typ = "application/octet-stream"
	}

	//typ = "text/plain; charset=UTF-8"

	slices := strings.Split(typ, ";")

	opts := []Option{ContentType(slices[0])}
	opts = append(opts, options...)

	return opts
}
