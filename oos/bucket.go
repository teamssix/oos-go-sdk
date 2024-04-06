// Package oos implements functions for access oos service.
// It has two main struct Client and Bucket.
package oos

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Client SDK's entry point. It's for bucket related options such as create/delete/set bucket (such as set/get ACL/lifecycle/referer/logging/website).
// Object related operations are done by Bucket class.
// Users use oos.New to create Client instance.
type (
	// Client oos client
	Client struct {
		Config *Config // oos client configuration
		Conn   *Conn   // Send HTTP request
	}

	// ClientOption client option such as UseCname, Timeout, SecurityToken.
	ClientOption func(*Client)
)

// New creates a new client.
//
// endpoint    the oos datacenter endpoint such as https://oos.ctyun.cn.
// accessKeyId    access key Id.
// accessKeySecret    access key secret.
//
// Client    creates the new client instance, the returned value is valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func New(endpoint, accessKeyID, accessKeySecret string, options ...ClientOption) (*Client, error) {
	// Configuration
	config := getDefaultoosConfig()

	// URL parse
	url := &urlMaker{}
	url.Init(endpoint, config.IsCname)

	config.Endpoint = url.NetLoc
	config.AccessKeyID = accessKeyID
	config.AccessKeySecret = accessKeySecret

	// HTTP connect
	conn := &Conn{config: config, url: url}

	// oos client
	client := &Client{
		config,
		conn,
	}

	// Client options parse
	for _, option := range options {
		option(client)
	}

	// Create HTTP connection
	err := conn.init(config, url)

	return client, err
}

// Bucket gets the bucket instance.
//
// bucketName    the bucket name.
// Bucket    the bucket object, when error is nil.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) Bucket(bucketName string) (*Object, error) {
	if bucketName == "" {
		return nil, errors.New("the parameter is invalid: bucket's name is empty")
	}
	return &Object{
		client,
		bucketName,
	}, nil
}

// oos 6.0版本 API 支持
// CreateBucket creates a bucket.
//
// conf          CreateBucketConfiguration
// bucketName    the bucket name, it's globably unique and immutable. The bucket name can only consist of lowercase letters, numbers and dash ('-').
//
//	It must start with lowercase letter or number and the length can only be between 3 and 255.
//
// options    options for creating the bucket, with optional ACL. The ACL could be ACLPrivate, ACLPublicRead, and ACLPublicReadWrite. By default it's ACLPrivate.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) CreateBucket(bucketName string, conf interface{}, options ...Option) error {
	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}
	headers := make(map[string]string)
	handleOptions(headers, options)

	buffer := new(bytes.Buffer)

	var bs []byte
	var err error
	bs, err = xml.Marshal(conf)

	if err != nil {
		return err
	}
	fmt.Println(string(bs))
	buffer.Write(bs)

	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

func (client Client) HeadBucket(bucketName string) (bool, error) {
	if bucketName == "" {
		return false, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	resp, err := client.do("HEAD", bucketName, params, nil, nil)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()
	err = checkRespCode(resp.StatusCode, []int{http.StatusOK})
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteBucket deletes the bucket. Only empty bucket can be deleted (no object and parts).
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) DeleteBucket(bucketName string) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusNoContent})
}

// ListBuckets lists buckets of the current account under the given endpoint, with optional filters.
//
// options    specifies the filters such as Prefix, Marker and MaxKeys. Prefix is the bucket name's prefix filter.
//
//	And marker makes sure the returned buckets' name are greater than it in lexicographic order.
//	Maxkeys limits the max keys to return, and by default it's 100 and up to 1000.
//	For the common usage scenario, please check out list_bucket.go in the sample.
//
// ListBucketsResponse    the response object if error is nil.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) ListBuckets(options ...Option) (ListBucketsResult, error) {
	var out ListBucketsResult

	params, err := getRawParams(options)
	if err != nil {
		return out, err
	}

	resp, err := client.do("GET", "", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// SetBucketACL sets bucket's ACL.
//
// bucketName    the bucket name
// bucketAcl    the bucket ACL: ACLPrivate, ACLPublicRead and ACLPublicReadWrite.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) SetBucketACL(bucketName string, bucketACL ACLType) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	headers := map[string]string{HTTPHeaderoosACL: string(bucketACL)}
	params := map[string]interface{}{}
	resp, err := client.do("PUT", bucketName, params, headers, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// GetBucketACL gets the bucket ACL.
//
// bucketName    the bucket name.
//
// GetBucketAclResponse    the result object, and it's only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (client Client) GetBucketACL(bucketName string) (GetBucketACLResult, error) {

	var out GetBucketACLResult
	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["acl"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// SetBucketLifecycle sets the bucket's lifecycle.
//
// bucketName    the bucket name.
// rules    the lifecycle rules. There're two kind of rules: absolute time expiration and relative time expiration in days and day/month/year respectively.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) SetBucketLifecycle(bucketName string, rules []LifecycleRule) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	var lxml LifecycleXML
	lxml.Rules = rules
	bs, err := xml.Marshal(lxml)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	//contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	params["lifecycle"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// DeleteBucketLifecycle deletes the bucket's lifecycle.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) DeleteBucketLifecycle(bucketName string) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["lifecycle"] = nil
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusNoContent})
}

// GetBucketLifecycle gets the bucket's lifecycle settings.
//
// bucketName    the bucket name.
//
// GetBucketLifecycleResult    the result object upon successful request. It's only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (client Client) GetBucketLifecycle(bucketName string) (GetBucketLifecycleResult, error) {
	var out GetBucketLifecycleResult

	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["lifecycle"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// SetBucketPolicy	 sets the bucket's Policy.
//
// bucketName    the bucket name.
// error    it's nil if no error, otherwise it's an error object.
func (client Client) SetBucketPolicy(bucketName string, text string) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	buffer := new(bytes.Buffer)
	buffer.Write([]byte(text))

	md5byte := md5.Sum([]byte(text))
	md5str := base64.StdEncoding.EncodeToString(md5byte[:])

	headers := map[string]string{}
	headers[HTTPHeaderContentType] = "application/xml"
	headers[HTTPHeaderContentMD5] = md5str

	params := map[string]interface{}{}
	params["policy"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// DeleteBucketPolicy deletes the bucket's Policy.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) DeleteBucketPolicy(bucketName string) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["policy"] = nil
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// GetBucketPolicy gets the bucket's Policy settings.
//
// bucketName    the bucket name.
//
// string    the result object upon successful request. It's only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (client Client) GetBucketPolicy(bucketName string) (string, error) {
	var out string

	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["policy"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	out = string(data)
	return out, err
}

// SetBucketLogging sets the bucket logging settings.
//
// oos could automatically store the access log. Only the bucket owner could enable the logging.
// Once enabled, oos would save all the access log into hourly log files in a specified bucket.
//
// bucketName    bucket name to enable the log.
// targetBucket    the target bucket name to store the log files.
// targetPrefix    the log files' prefix.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) SetBucketLogging(bucketName, targetBucket, targetPrefix string,
	isEnable bool) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	var err error
	var bs []byte
	if isEnable {
		lxml := LoggingXML{}
		lxml.LoggingEnabled.TargetBucket = targetBucket
		lxml.LoggingEnabled.TargetPrefix = targetPrefix
		bs, err = xml.Marshal(lxml)
	} else {
		lxml := loggingXMLEmpty{}
		bs, err = xml.Marshal(lxml)
	}

	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	buffer.Write(bs)
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	params["logging"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// GetBucketLogging gets the bucket's logging settings
//
// bucketName    the bucket name
// GetBucketLoggingResponse    the result object upon successful request. It's only valid when error is nil.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) GetBucketLogging(bucketName string) (GetBucketLoggingResult, error) {
	var out GetBucketLoggingResult

	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["logging"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// SetBucketWebsite sets the bucket's static website's index and error page.
//
// oos supports static web site hosting for the bucket data. When the bucket is enabled with that, you can access the file in the bucket like the way to access a static website.
//
// bucketName    the bucket name to enable static web site.
// indexDocument    index page.
// errorDocument    error page.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) SetBucketWebsite(bucketName string, configuration WebsiteConfiguration) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	wxml := buildBucketWebsiteXml(configuration)

	bs, err := xml.Marshal(wxml)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	//contentType := http.DetectContentType(buffer.Bytes())
	headers := make(map[string]string)
	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	params["website"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// DeleteBucketWebsite deletes the bucket's static web site settings.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) DeleteBucketWebsite(bucketName string) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["website"] = nil
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

// GetBucketWebsite gets the bucket's default page (index page) and the error page.
//
// bucketName    the bucket name
//
// GetBucketWebsiteResponse    the result object upon successful request. It's only valid when error is nil.
// error    it's nil if no error, otherwise it's an error object.
func (client Client) GetBucketWebsite(bucketName string) (GetBucketWebsiteResult, error) {
	var out GetBucketWebsiteResult

	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["website"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

/*
<ObjectLockConfiguration>

	<ObjectLockEnabled>Enabled</ObjectLockEnabled>
	<Rule>
	    <DefaultRetention>
	        <Mode>COMPLIANCE</Mode>
	        <Days>days</Days>
	        <Years>years</Years>
	    </DefaultRetention>
	</Rule>

</ObjectLockConfiguration>
*/
func (client Client) SetBucketObjectLock(bucketName string, lock BucketObjectLock) error {
	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	var err error
	var bs []byte
	bs, err = xml.Marshal(lock)
	if err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	headers := map[string]string{}
	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	params["object-lock"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

func (client Client) GetBucketObjectLock(bucketName string) (BucketObjectLock, error) {
	var out BucketObjectLock

	if bucketName == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["object-lock"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

func (client Client) DeleteBucketObjectLock(bucketName string) error {
	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}
	params := map[string]interface{}{}
	params["object-lock"] = nil
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

func (client Client) SetBucketCors(bucketName string, rules []CORSRule) error {

	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	var corsXml CORSXML
	corsXml.CORSRules = rules
	bs, err := xml.Marshal(corsXml)
	if err != nil {
		return err
	}
	buffer := new(bytes.Buffer)
	buffer.Write(bs)

	// //contentType := http.DetectContentType(buffer.Bytes())
	headers := map[string]string{}
	headers[HTTPHeaderContentType] = "application/xml"

	params := map[string]interface{}{}
	params["cors"] = nil
	resp, err := client.do("PUT", bucketName, params, headers, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

func (client Client) GetBucketCors(bucketName string) ([]CORSRule, error) {
	var out CORSXML

	if bucketName == "" {
		return nil, errors.New("the parameter is invalid: bucket's name is empty")
	}

	params := map[string]interface{}{}
	params["cors"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out.CORSRules, err
}

func (client Client) DeleteBucketCors(bucketName string) error {
	if bucketName == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}
	params := map[string]interface{}{}
	params["cors"] = nil
	resp, err := client.do("DELETE", bucketName, params, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return checkRespCode(resp.StatusCode, []int{http.StatusOK})
}

func (client Client) GetRegions() (GetRegionsResult, error) {
	var out GetRegionsResult
	params := map[string]interface{}{}
	params["regions"] = nil
	resp, err := client.do("GET", "", params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

func (client Client) GetBucketLocation(bucketName string) (GetBucketLocation, error) {
	var out GetBucketLocation
	params := map[string]interface{}{}
	params["location"] = nil
	resp, err := client.do("GET", bucketName, params, nil, nil)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()

	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// UseCname sets the flag of using CName. By default it's false.
//
// isUseCname    true: the endpoint has the CName, false: the endpoint does not have cname. Default is false.
func UseCname(isUseCname bool) ClientOption {
	return func(client *Client) {
		client.Config.IsCname = isUseCname
		client.Conn.url.Init(client.Config.Endpoint, client.Config.IsCname)
	}
}

// Timeout sets the HTTP timeout in seconds.
//
// connectTimeoutSec    HTTP timeout in seconds. Default is 10 seconds. 0 means infinite (not recommended)
// readWriteTimeout    HTTP read or write's timeout in seconds. Default is 20 seconds. 0 means infinite.
func Timeout(connectTimeoutSec, readWriteTimeout int64) ClientOption {
	return func(client *Client) {
		client.Config.HTTPTimeout.ConnectTimeout =
			time.Second * time.Duration(connectTimeoutSec)
		client.Config.HTTPTimeout.ReadWriteTimeout =
			time.Second * time.Duration(readWriteTimeout)
		client.Config.HTTPTimeout.HeaderTimeout =
			time.Second * time.Duration(readWriteTimeout)
		client.Config.HTTPTimeout.IdleConnTimeout =
			time.Second * time.Duration(readWriteTimeout)
		client.Config.HTTPTimeout.LongTimeout =
			time.Second * time.Duration(readWriteTimeout*10)
	}
}

// SecurityToken sets the temporary user's SecurityToken.
//
// token    STS token
func SecurityToken(token string) ClientOption {
	return func(client *Client) {
		client.Config.SecurityToken = strings.TrimSpace(token)
	}
}

// MD5ThresholdCalcInMemory sets the memory usage threshold for computing the MD5, default is 16MB.
//
// threshold    the memory threshold in bytes. When the uploaded content is more than 16MB, the temp file is used for computing the MD5.
func MD5ThresholdCalcInMemory(threshold int64) ClientOption {
	return func(client *Client) {
		client.Config.MD5Threshold = threshold
	}
}

// EnableSha256ForPayload set whether to use sha256 to create a hash for the payload . default is true
//
// isEnable    Whether to use sha256 to create a hash for the payload. true:enbale ; false:diaable
func EnableSha256ForPayload(isEnable bool) ClientOption {
	return func(client *Client) {
		client.Config.IsEnableSHA256 = isEnable
	}
}

// SHA256ThresholdCalcInMemory sets the memory usage threshold for computing the MD5, default is 16MB.
//
// threshold    the memory threshold in bytes. When the uploaded content is more than 16MB, the temp file is used for computing the MD5.
func SHA256ThresholdCalcInMemory(threshold int64) ClientOption {
	return func(client *Client) {
		client.Config.SHA256Threshold = threshold
	}
}

// V4Signature set request signature . default is V4
//
// isV4    the signure is V4 or V2 . true:V4 ; false:V2
func V4Signature(isV4 bool) ClientOption {
	return func(client *Client) {
		client.Config.IsV4Sign = isV4
	}
}

// UserAgent specifies UserAgent. The default is oos-go-sdk-go/1.2.0 (windows/-/amd64;go1.5.2).
//
// userAgent    the user agent string.
func UserAgent(userAgent string) ClientOption {
	return func(client *Client) {
		client.Config.UserAgent = userAgent
	}
}

// Private
func (client Client) do(method, bucketName string, params map[string]interface{},
	headers map[string]string, data io.Reader) (*Response, error) {
	return client.Conn.Do(method, bucketName, "", params,
		headers, data, nil)
}
