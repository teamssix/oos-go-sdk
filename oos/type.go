package oos

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"time"
)

// ListBucketsResult defines the result object from ListBuckets request
type ListBucketsResult struct {
	XMLName xml.Name           `xml:"ListAllMyBucketsResult"`
	Owner   Owner              `xml:"Owner"`          // The owner information
	Buckets []BucketProperties `xml:"Buckets>Bucket"` // The bucket list
}

// BucketProperties defines bucket properties
type BucketProperties struct {
	XMLName      xml.Name  `xml:"Bucket"`
	Name         string    `xml:"Name"`         // Bucket name
	CreationDate time.Time `xml:"CreationDate"` // Bucket create time
}

type GetRegionsResult struct {
	XMLName         xml.Name `xml:"BucketRegions"`
	MetadataRegions []string `xml:"MetadataRegions>Region"`
	DataRegions     []string `xml:"DataRegions>Region"`
}

type GetBucketLocation struct {
	XMLName          xml.Name         `xml:"BucketConfiguration"`
	MetaLocation     string           `xml:"MetadataLocationConstraint>Location"`
	DataLocationType DataLocationType `xml:"DataLocationConstraint>Type"`
	DataLocationList []string         `xml:"DataLocationConstraint>LocationList>Location"`
	ScheduleStrategy ScheduleStrategy `xml:"DataLocationConstraint>ScheduleStrategy"`
}

// GetBucketACLResult defines GetBucketACL request's result
type GetBucketACLResult struct {
	XMLName   xml.Name   `xml:"AccessControlPolicy"`
	Owner     Owner      `xml:"Owner"`                   // Bucket owner
	GrantList []AclGrant `xml:"AccessControlList>Grant"` // Bucket ACL
}

type AclGrant struct {
	XMLName    xml.Name `xml:"Grant"`
	GranteeURI string   `xml:"Grantee>URI"` // Grantee URI
	Permission string   `xml:"Permission"`  // Bucket许可信息 READ 只读 ; FULL_CONTROL 公有 ；空 私有
}

// LifecycleConfiguration is the Bucket Lifecycle configuration
type LifecycleConfiguration struct {
	XMLName xml.Name        `xml:"LifecycleConfiguration"`
	Rules   []LifecycleRule `xml:"Rule"`
}

// LifecycleRule defines Lifecycle rules
type LifecycleRule struct {
	XMLName    xml.Name             `xml:"Rule"`
	ID         string               `xml:"ID,omitempty"`         // The rule ID
	Prefix     string               `xml:"Prefix"`               // The object key prefix
	Status     string               `xml:"Status"`               // The rule status (enabled or not)
	Expiration *LifecycleExpiration `xml:"Expiration,omitempty"` // The expiration property
	Transition *LifecycleTransition `xml:"Transition,omitempty"` // The Transition property
}

// LifecycleExpiration defines the rule's expiration property
type LifecycleExpiration struct {
	XMLName xml.Name `xml:"Expiration"`
	Days    int      `xml:"Days,omitempty"` // Relative expiration time: The expiration time in days after the last modified time
	Date    string   `xml:"Date,omitempty"` // Absolute expiration time: The expiration time in date.
}

type LifecycleTransition struct {
	XMLName      xml.Name `xml:"Transition"`
	Days         int      `xml:"Days,omitempty"` // Relative expiration time: The expiration time in days after the last modified time
	Date         string   `xml:"Date,omitempty"` // Absolute expiration time: The expiration time in date.
	StorageClass string   `xml:"StorageClass"`
}

type LifecycleXML struct {
	XMLName xml.Name        `xml:"LifecycleConfiguration"`
	Rules   []LifecycleRule `xml:"Rule"`
}

const lifecycleDateFormat = "2006-01-02T15:04:05.000Z"

// BuildLifecycleExpirRuleByDays builds a lifecycle rule with specified expiration days
func BuildLifecycleExpirRuleByDays(id, prefix string, status bool, days int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: &LifecycleExpiration{Days: days}}
}

// BuildLifecycleExpirRuleByDate builds a lifecycle rule with specified expiration time.
func BuildLifecycleExpirRuleByDate(id, prefix string, status bool, year, month, day int) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Expiration: &LifecycleExpiration{Date: date.Format(lifecycleDateFormat)}}
}

func BuildLifecycleTransitionRuleByDays(id, prefix string, status bool, days int, storageClass string) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Transition: &LifecycleTransition{Days: days, StorageClass: storageClass}}
}

// BuildLifecycleExpirRuleByDate builds a lifecycle rule with specified expiration time.
func BuildLifecycleTransitionRuleByDate(id, prefix string, status bool, year, month, day int, storageClass string) LifecycleRule {
	var statusStr = "Enabled"
	if !status {
		statusStr = "Disabled"
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return LifecycleRule{ID: id, Prefix: prefix, Status: statusStr,
		Transition: &LifecycleTransition{Date: date.Format(lifecycleDateFormat), StorageClass: storageClass}}
}

// GetBucketLifecycleResult defines GetBucketLifecycle's result object
type GetBucketLifecycleResult LifecycleConfiguration

// RefererXML defines Referer configuration
type RefererXML struct {
	XMLName           xml.Name `xml:"RefererConfiguration"`
	AllowEmptyReferer bool     `xml:"AllowEmptyReferer"`   // Allow empty referrer
	RefererList       []string `xml:"RefererList>Referer"` // Referer whitelist
}

// GetBucketRefererResult defines result object for GetBucketReferer request
type GetBucketRefererResult RefererXML

// LoggingXML defines logging configuration
type LoggingXML struct {
	XMLName        xml.Name       `xml:"BucketLoggingStatus"`
	LoggingEnabled LoggingEnabled `xml:"LoggingEnabled"` // The logging configuration information
}

type loggingXMLEmpty struct {
	XMLName xml.Name `xml:"BucketLoggingStatus"`
}

// LoggingEnabled defines the logging configuration information
type LoggingEnabled struct {
	XMLName      xml.Name `xml:"LoggingEnabled"`
	TargetBucket string   `xml:"TargetBucket"` // The bucket name for storing the log files
	TargetPrefix string   `xml:"TargetPrefix"` // The log file prefix
}

// GetBucketLoggingResult defines the result from GetBucketLogging request
type GetBucketLoggingResult LoggingXML

type WebsiteConfiguration struct {
	IndexDocument       IndexDocument
	ErrorDocument       ErrorDocument
	WebsiteAllRequestTo *WebsiteAllRequestToXML
	RoutingRules        []RoutingRule
}

func buildBucketWebsiteXml(configruation WebsiteConfiguration) WebsiteXML {
	xml := WebsiteXML{}
	if configruation.IndexDocument.Suffix != "" {
		xml.IndexDocument = &configruation.IndexDocument
	}
	if configruation.ErrorDocument.Key != "" {
		xml.ErrorDocument = &configruation.ErrorDocument
	}
	if nil != configruation.WebsiteAllRequestTo {
		xml.WebsiteAllRequestTo = configruation.WebsiteAllRequestTo
	}
	if nil != configruation.RoutingRules || len(configruation.RoutingRules) > 0 {
		xml.RoutingRules = &configruation.RoutingRules
	}
	return xml
}

// WebsiteXML defines Website configuration
type WebsiteXML struct {
	XMLName             xml.Name                `xml:"WebsiteConfiguration"`
	IndexDocument       *IndexDocument          `xml:"IndexDocument,omitempty"` // The index page
	ErrorDocument       *ErrorDocument          `xml:"ErrorDocument,omitempty"` // The error page
	WebsiteAllRequestTo *WebsiteAllRequestToXML `xml:"RedirectAllRequestsTo,omitempty"`
	RoutingRules        *[]RoutingRule          `xml:"RoutingRules>RoutingRule,omitempty"`
}

type IndexDocument struct {
	XMLName xml.Name `xml:"IndexDocument"`
	Suffix  string   `xml:"Suffix"`
}

type ErrorDocument struct {
	XMLName xml.Name `xml:"ErrorDocument"`
	Key     string   `xml:"Key"`
}

type WebsiteAllRequestToXML struct {
	XMLName  xml.Name `xml:"RedirectAllRequestsTo"`
	HostName string   `xml:"HostName,omitempty"`
	Protocol string   `xml:"Protocol,omitempty"`
}

type RoutingRule struct {
	XMLName   xml.Name   `xml:"RoutingRule"`
	Condition *Condition `xml:"Condition,omitempty"`
	Redirect  *Redirect  `xml:"Redirect,omitempty"`
}

type Condition struct {
	XMLName                     xml.Name `xml:"Condition"`
	HttpErrorCodeReturnedEquals string   `xml:"HttpErrorCodeReturnedEquals,omitempty"`
	KeyPrefixEquals             string   `xml:"KeyPrefixEquals,omitempty"`
}

type Redirect struct {
	XMLName              xml.Name `xml:"Redirect"`
	HostName             string   `xml:"HostName,omitempty"`
	Protocol             string   `xml:"Protocol,omitempty"`
	ReplaceKeyPrefixWith string   `xml:"ReplaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith       string   `xml:"ReplaceKeyWith,omitempty"`
}

// GetBucketWebsiteResult defines the result from GetBucketWebsite request.
type GetBucketWebsiteResult WebsiteXML

// CreateAccessKeyResponse
type CreateAccessKeyResponse struct {
	XMLName               xml.Name              `xml:"CreateAccessKeyResponse"`
	CreateAccessKeyResult CreateAccessKeyResult `xml:"CreateAccessKeyResult"`
	ResponseMetadata      ResponseMetadata      `xml:"ResponseMetadata"`
}

// CreateAccessKeyResult>
type CreateAccessKeyResult struct {
	XMLName  xml.Name     `xml:"CreateAccessKeyResult"`
	AcessKey AcessKeyInfo `xml:"AccessKey"`
}

// AccessKey
type AcessKeyInfo struct {
	XMLName         xml.Name   `xml:"AccessKey"`
	UserName        string     `xml:"UserName"`
	AccessKeyId     string     `xml:"AccessKeyId"`
	Status          string     `xml:"Status"`
	SecretAccessKey string     `xml:"SecretAccessKey"`
	CreateDate      *time.Time `xml:"CreateDate,omitempty"`
}

// ResponseMetadata>
type ResponseMetadata struct {
	XMLName   xml.Name `xml:"ResponseMetadata"`
	RequestId string   `xml:"RequestId"`
}

// DeleteAccessKeyResponse
type DeleteAccessKeyResponse struct {
	XMLName          xml.Name         `xml:"DeleteAccessKeyResponse"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// UpdateAccessKeyResponse
type UpdateAccessKeyResponse struct {
	XMLName          xml.Name         `xml:"UpdateAccessKeyResponse"`
	ResponseMetadata ResponseMetadata `xml:"ResponseMetadata"`
}

// ListAccessKeysResponse
type ListAccessKeysResponse struct {
	XMLName              xml.Name             `xml:"ListAccessKeysResponse"`
	ListAccessKeysResult ListAccessKeysResult `xml:"ListAccessKeysResult"`
	ResponseMetadata     ResponseMetadata     `xml:"ResponseMetadata"`
}

type GetAccessKeyLastUsedResponse struct {
	XMLName                    xml.Name                   `xml:"GetAccessKeyLastUsedResponse"`
	GetAccessKeyLastUsedResult GetAccessKeyLastUsedResult `xml:"GetAccessKeyLastUsedResult"`
	ResponseMetadata           ResponseMetadata           `xml:"ResponseMetadata"`
}

type GetAccessKeyLastUsedResult struct {
	UserName     string     `xml:"UserName"`
	LastUsedDate *time.Time `xml:"AccessKeyLastUsed>LastUsedDate"`
	ServiceName  string     `xml:"AccessKeyLastUsed>ServiceName"`
}

// ListAccessKeysResult
type ListAccessKeysResult struct {
	XMLName     xml.Name `xml:"ListAccessKeysResult"`
	UserName    string   `xml:"UserName"`
	MemberList  []member `xml:"AccessKeyMetadata>member"`
	IsTruncated string   `xml:"IsTruncated"`
	Marker      string   `xml:"Marker"`
}

// AccessKeyMetadataMember
type member struct {
	XMLName     xml.Name   `xml:"member"`
	UserName    string     `xml:"UserName"`
	AccessKeyId string     `xml:"AccessKeyId"`
	Status      string     `xml:"Status"`
	IsPrimary   string     `xml:"IsPrimary"`
	CreateDate  *time.Time `xml:"CreateDate,omitempty"`
}

type deleteXML struct {
	XMLName xml.Name       `xml:"Delete"`
	Objects []DeleteObject `xml:"Object"` // Objects to delete
	Quiet   bool           `xml:"Quiet"`  // Flag of quiet mode.
}

// DeleteObject defines the struct for deleting object
type DeleteObject struct {
	XMLName xml.Name `xml:"Object"`
	Key     string   `xml:"Key"` // Object name
}

// CORSXML defines CORS configuration
type CORSXML struct {
	XMLName   xml.Name   `xml:"CORSConfiguration"`
	CORSRules []CORSRule `xml:"CORSRule"` // CORS rules
}

// CORSRule defines CORS rules
type CORSRule struct {
	XMLName       xml.Name `xml:"CORSRule"`
	AllowedOrigin []string `xml:"AllowedOrigin"` // Allowed origins. By default it's wildcard '*'
	AllowedMethod []string `xml:"AllowedMethod"` // Allowed methods
	AllowedHeader []string `xml:"AllowedHeader"` // Allowed headers
	ExposeHeader  []string `xml:"ExposeHeader"`  // Allowed response headers
	MaxAgeSeconds int      `xml:"MaxAgeSeconds"` // Max cache ages in seconds
}

// GetBucketCORSResult defines the result from GetBucketCORS request.
type GetBucketCORSResult CORSXML

// ListObjectsResult defines the result from ListObjects request
type ListObjectsResult struct {
	XMLName        xml.Name           `xml:"ListBucketResult"`
	Prefix         string             `xml:"Prefix"`                // The object prefix
	Marker         string             `xml:"Marker"`                // The marker filter.
	MaxKeys        int                `xml:"MaxKeys"`               // Max keys to return
	Delimiter      string             `xml:"Delimiter"`             // The delimiter for grouping objects' name
	IsTruncated    bool               `xml:"IsTruncated"`           // Flag indicates if all results are returned (when it's false)
	NextMarker     string             `xml:"NextMarker"`            // The start point of the next query
	Objects        []ObjectProperties `xml:"Contents"`              // Object list
	CommonPrefixes []string           `xml:"CommonPrefixes>Prefix"` // You can think of commonprefixes as "folders" whose names end with the delimiter
	EncodingType   string             `xml:"EncodingType"`          // encoding object key
}

// ObjectProperties defines Objecct properties
type ObjectProperties struct {
	XMLName      xml.Name  `xml:"Contents"`
	Key          string    `xml:"Key"`          // Object key
	Size         int64     `xml:"Size"`         // Object size
	ETag         string    `xml:"ETag"`         // Object ETag
	Owner        Owner     `xml:"Owner"`        // Object owner information
	LastModified time.Time `xml:"LastModified"` // Object last modified time
	StorageClass string    `xml:"StorageClass"` // Object storage class (Standard, IA, Archive)
}

// Owner defines Bucket/Object's owner
type Owner struct {
	XMLName     xml.Name `xml:"Owner"`
	ID          string   `xml:"ID"`          // Owner ID
	DisplayName string   `xml:"DisplayName"` // Owner's display name
}

// CopyObjectResult defines result object of CopyObject
type CopyObjectResult struct {
	XMLName      xml.Name  `xml:"CopyObjectResult"`
	LastModified time.Time `xml:"LastModified"` // New object's last modified time.
	ETag         string    `xml:"ETag"`         // New object's ETag
}

// GetObjectACLResult defines result of GetObjectACL request
type GetObjectACLResult GetBucketACLResult

// DeleteObjectsResult defines result of DeleteObjects request
type DeleteObjectsResult struct {
	XMLName        xml.Name `xml:"DeleteResult"`
	DeletedObjects []string `xml:"Deleted>Key"` // Deleted object list
}

// InitiateMultipartUploadResult defines result of InitiateMultipartUpload request
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`   // Bucket name
	Key      string   `xml:"Key"`      // Object name to upload
	UploadID string   `xml:"UploadId"` // Generated UploadId
}

// UploadPart defines the upload/copy part
type UploadPart struct {
	XMLName    xml.Name `xml:"Part"`
	PartNumber int      `xml:"PartNumber"` // Part number
	ETag       string   `xml:"ETag"`       // ETag value of the part's data
}

type uploadParts []UploadPart

func (slice uploadParts) Len() int {
	return len(slice)
}

func (slice uploadParts) Less(i, j int) bool {
	return slice[i].PartNumber < slice[j].PartNumber
}

func (slice uploadParts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

// UploadPartCopyResult defines result object of multipart copy request.
type UploadPartCopyResult struct {
	XMLName      xml.Name  `xml:"CopyPartResult"`
	LastModified time.Time `xml:"LastModified"` // Last modified time
	ETag         string    `xml:"ETag"`         // ETag
}

type completeMultipartUploadXML struct {
	XMLName xml.Name     `xml:"CompleteMultipartUpload"`
	Part    []UploadPart `xml:"Part"`
}

// CompleteMultipartUploadResult defines result object of CompleteMultipartUploadRequest
type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	Location string   `xml:"Location"` // Object URL
	Bucket   string   `xml:"Bucket"`   // Bucket name
	ETag     string   `xml:"ETag"`     // Object ETag
	Key      string   `xml:"Key"`      // Object name
}

// ListUploadedPartsResult defines result object of ListUploadedParts
type ListUploadedPartsResult struct {
	XMLName              xml.Name       `xml:"ListPartsResult"`
	Bucket               string         `xml:"Bucket"`               // Bucket name
	Key                  string         `xml:"Key"`                  // Object name
	UploadID             string         `xml:"UploadId"`             // Upload ID
	NextPartNumberMarker string         `xml:"NextPartNumberMarker"` // Next part number
	MaxParts             int            `xml:"MaxParts"`             // Max parts count
	IsTruncated          bool           `xml:"IsTruncated"`          // Flag indicates all entries returned.false: all entries returned.
	Initiator            Initiator      `xml:"Initiator"`            // Initiator
	Owner                Owner          `xml:"Owner"`                // Owner
	StorageClass         string         `xml:"StorageClass"`         // StorageClass
	UploadedParts        []UploadedPart `xml:"Part"`                 // Uploaded parts
	EncodingType         string         `xml:"EncodingType"`         // encoding object key
}

// UploadedPart defines uploaded part
type UploadedPart struct {
	XMLName      xml.Name  `xml:"Part"`
	PartNumber   int       `xml:"PartNumber"`   // Part number
	LastModified time.Time `xml:"LastModified"` // Last modified time
	ETag         string    `xml:"ETag"`         // ETag cache
	Size         int       `xml:"Size"`         // Part size
}

type SrcCopyPartObject struct {
	BucketName string
	ObjectName string
	PartNumber int
}

type Initiator struct {
	XMLName     xml.Name `xml:"Initiator"`
	ID          string   `xml:"ID"`
	DisplayName string   `xml:"DisplayName"`
}

// ListMultipartUploadResult defines result object of ListMultipartUpload
type ListMultipartUploadResult struct {
	XMLName            xml.Name            `xml:"ListMultipartUploadsResult"`
	Bucket             string              `xml:"Bucket"`                // Bucket name
	Delimiter          string              `xml:"Delimiter"`             // Delimiter for grouping object.
	Prefix             string              `xml:"Prefix"`                // Object prefix
	KeyMarker          string              `xml:"KeyMarker"`             // Object key marker
	UploadIDMarker     string              `xml:"UploadIdMarker"`        // UploadId marker
	NextKeyMarker      string              `xml:"NextKeyMarker"`         // Next key marker, if not all entries returned.
	NextUploadIDMarker string              `xml:"NextUploadIdMarker"`    // Next uploadId marker, if not all entries returned.
	MaxUploads         int                 `xml:"MaxUploads"`            // Max uploads to return
	IsTruncated        bool                `xml:"IsTruncated"`           // Flag indicates all entries are returned.
	Uploads            []UncompletedUpload `xml:"Upload"`                // Ongoing uploads (not completed, not aborted)
	CommonPrefixes     []string            `xml:"CommonPrefixes>Prefix"` // Common prefixes list.
	EncodingType       string              `xml:"EncodingType"`          // encoding object key
}

// UncompletedUpload structure wraps an uncompleted upload task
type UncompletedUpload struct {
	XMLName      xml.Name  `xml:"Upload"`
	Key          string    `xml:"Key"`          // Object name
	UploadID     string    `xml:"UploadId"`     // The UploadId
	Initiator    Initiator `xml:"Initiator"`    // Initiator
	Owner        Owner     `xml:"Owner"`        // Owner
	StorageClass string    `xml:"StorageClass"` // StorageClass
	Initiated    time.Time `xml:"Initiated"`    // Initialization time in the format such as 2012-02-23T04:18:23.000Z
}

// ProcessObjectResult defines result object of ProcessObject
type ProcessObjectResult struct {
	Bucket   string `json:"bucket"`
	FileSize int    `json:"fileSize"`
	Object   string `json:"object"`
	Status   string `json:"status"`
}

// decodeDeleteObjectsResult decodes deleting objects result in URL encoding
func decodeDeleteObjectsResult(result *DeleteObjectsResult) error {
	var err error
	for i := 0; i < len(result.DeletedObjects); i++ {
		result.DeletedObjects[i], err = url.QueryUnescape(result.DeletedObjects[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// decodeListObjectsResult decodes list objects result in URL encoding
func decodeListObjectsResult(result *ListObjectsResult) error {
	var err error
	result.Prefix, err = url.QueryUnescape(result.Prefix)
	if err != nil {
		return err
	}
	result.Marker, err = url.QueryUnescape(result.Marker)
	if err != nil {
		return err
	}
	result.Delimiter, err = url.QueryUnescape(result.Delimiter)
	if err != nil {
		return err
	}
	result.NextMarker, err = url.QueryUnescape(result.NextMarker)
	if err != nil {
		return err
	}
	for i := 0; i < len(result.Objects); i++ {
		result.Objects[i].Key, err = url.QueryUnescape(result.Objects[i].Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(result.CommonPrefixes); i++ {
		result.CommonPrefixes[i], err = url.QueryUnescape(result.CommonPrefixes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// decodeListUploadedPartsResult decodes
func decodeListUploadedPartsResult(result *ListUploadedPartsResult) error {
	var err error
	result.Key, err = url.QueryUnescape(result.Key)
	if err != nil {
		return err
	}
	return nil
}

// decodeListMultipartUploadResult decodes list multipart upload result in URL encoding
func decodeListMultipartUploadResult(result *ListMultipartUploadResult) error {
	var err error
	result.Prefix, err = url.QueryUnescape(result.Prefix)
	if err != nil {
		return err
	}
	result.Delimiter, err = url.QueryUnescape(result.Delimiter)
	if err != nil {
		return err
	}
	result.KeyMarker, err = url.QueryUnescape(result.KeyMarker)
	if err != nil {
		return err
	}
	result.NextKeyMarker, err = url.QueryUnescape(result.NextKeyMarker)
	if err != nil {
		return err
	}
	for i := 0; i < len(result.Uploads); i++ {
		result.Uploads[i].Key, err = url.QueryUnescape(result.Uploads[i].Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(result.CommonPrefixes); i++ {
		result.CommonPrefixes[i], err = url.QueryUnescape(result.CommonPrefixes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// createBucketConfiguration defines the configuration for creating a bucket.
type createBucketConfiguration struct {
	XMLName      xml.Name     `xml:"CreateBucketConfiguration"`
	MetaLocation string       `xml:"MetadataLocationConstraint>Location"`
	DataLocation dataLocation `xml:"DataLocationConstraint"`
}

// createBucketConfiguration defines the configuration for creating a bucket.
type createBucketConfigurationLocal struct {
	XMLName      xml.Name `xml:"CreateBucketConfiguration"`
	MetaLocation string   `xml:"MetadataLocationConstraint>Location"`
	DataType     string   `xml:"DataLocationConstraint>Type"`
}

// dataLocation
type dataLocation struct {
	XMLName          xml.Name `xml:"DataLocationConstraint"`
	Type             string   `xml:"Type"`
	LocationList     []string `xml:"LocationList>Location"`
	ScheduleStrategy string   `xml:"ScheduleStrategy"`
}

// meta location valid ranges
var metaLocationRange = []string{"ChengDu", "FuZhou", "GuiYang", "HangZhou", "LaSa",
	"LanZhou", "QingDao", "ShenYang", "ShenZhen", "WuHan", "WuHu", "WuLuMuQi", "ZhengZhou",
	"SH2", "SuZhou"}

// date location valid ranges
var dataLocationRange = []string{"ChengDu", "GuiYang", "LaSa", "LanZhou", "QingDao",
	"SH2", "ShenYang", "ShenZhen", "SuZhou", "WuHan", "WuHu", "WuLuMuQi", "ZhengZhou"}

func BuildCreateBucketConfigLocal(metaLocation string) (createBucketConfigurationLocal, error) {
	if result := IsInRange(metaLocation, metaLocationRange); !result {
		return createBucketConfigurationLocal{}, fmt.Errorf("Meta location is invalid, value must in %v", metaLocationRange)
	}

	return createBucketConfigurationLocal{
		MetaLocation: metaLocation,
		DataType:     "Local",
	}, nil
}

// build createBucketConfiguration Specified
func BuildCreateBucketConfigSpecified(metaLocation string, locationList []string, AllowedSchedule bool) (createBucketConfiguration, error) {
	if result := IsInRange(metaLocation, metaLocationRange); !result {
		return createBucketConfiguration{}, fmt.Errorf("meta location is invalid, value must in %v ", metaLocationRange)
	}

	if len(locationList) == 0 {
		return createBucketConfiguration{}, fmt.Errorf("data location list is not empty ")
	}

	for _, lo := range locationList {
		if result := IsInRange(lo, dataLocationRange); !result {
			return createBucketConfiguration{}, fmt.Errorf("data location is invalid, value must in %v ", dataLocationRange)
		}
	}

	scheduleStrategy := "NotAllowed"
	if AllowedSchedule {
		scheduleStrategy = "Allowed"
	}

	return createBucketConfiguration{
		MetaLocation: metaLocation,
		DataLocation: dataLocation{
			Type:             "Specified",
			LocationList:     locationList,
			ScheduleStrategy: scheduleStrategy,
		},
	}, nil
}

func BuildObjectLocalDataLocation(allowedSchedule bool) string {
	var dataLocation = "type=Local,scheduleStrategy="
	if allowedSchedule {
		dataLocation += "Allowed"
	} else {
		dataLocation += "NotAllowed"
	}
	return dataLocation
}

/*
You must pass in the datalocation to be specified and whether scheduling is allowed.
Otherwise, the upload will fail.
*/
func BuildObjectSpecifiedDataLocation(location string, allowedSchedule bool) string {
	var dataLocation = "type=Specified,location=" + location + ",scheduleStrategy="
	if allowedSchedule {
		dataLocation += "Allowed"
	} else {
		dataLocation += "NotAllowed"
	}
	return dataLocation
}

// object lock
type BucketObjectLock struct {
	XMLName           xml.Name         `xml:"ObjectLockConfiguration"`
	ObjectLockEnabled string           `xml:"ObjectLockEnabled"`
	DefaultRetention  DefaultRetention `xml:"Rule>DefaultRetention"`
}

type DefaultRetention struct {
	XMLName xml.Name `xml:"DefaultRetention"`
	Mode    string   `xml:"Mode"`
	Days    int      `xml:"Days,omitempty"`
	Years   int      `xml:"Years,omitempty"`
}

func BuildBucketObjectLockByDays(mode string, days int, isEnabel bool) BucketObjectLock {
	defaultRetention := DefaultRetention{Mode: mode, Days: days}
	var lock BucketObjectLock
	lock.ObjectLockEnabled = "Enabled"
	if !isEnabel {
		lock.ObjectLockEnabled = "Disabled"
	}
	lock.DefaultRetention = defaultRetention
	return lock
}

func BuildBucketObjectLockByYears(mode string, years int, isEnabel bool) BucketObjectLock {
	defaultRetention := DefaultRetention{Mode: mode, Years: years}
	var lock BucketObjectLock
	lock.ObjectLockEnabled = "Enabled"
	if !isEnabel {
		lock.ObjectLockEnabled = "Disabled"
	}
	lock.DefaultRetention = defaultRetention
	return lock
}

// object lock end
