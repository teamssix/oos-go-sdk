package oos

import "os"

// ACLType bucket/object ACL
type ACLType string

const (
	// ACLPrivate definition : private read and write
	ACLPrivate ACLType = "private"

	// ACLPublicRead definition : public read and private write
	ACLPublicRead ACLType = "public-read"

	// ACLPublicReadWrite definition : public read and public write
	ACLPublicReadWrite ACLType = "public-read-write"
)

type ScheduleStrategy string
type DataLocationType string

const (
	ScheduleStrategyAllow      ScheduleStrategy = "Allowed"
	ScheduleStrategyNotAllowed ScheduleStrategy = "NotAllowed"
	DataLocationTypeLocal      DataLocationType = "Local"
	DataLocationTypeSpecified  DataLocationType = "Specified"
)

// MetadataDirectiveType specifying whether use the metadata of source object when copying object.
type MetadataDirectiveType string

const (
	// MetaCopy the target object's metadata is copied from the source one
	MetaCopy MetadataDirectiveType = "COPY"

	// MetaReplace the target object's metadata is created as part of the copy request (not same as the source one)
	MetaReplace MetadataDirectiveType = "REPLACE"
)

// StorageClassType bucket storage type
type StorageClassType string

const (
	StorageClassStandard          StorageClassType = "STANDARD"
	StorageClassStandardIA        StorageClassType = "STANDARD_IA"
	StorageClassReducedRedundancy StorageClassType = "REDUCED_REDUNDANCY"
)

type BucketMode string

const (
	BucketModeCompliance BucketMode = "COMPLIANCE"
)

// PayerType the type of request payer
type PayerType string

const (
	// Requester the requester who send the request
	Requester PayerType = "requester"
)

// HTTPMethod HTTP request method
type HTTPMethod string

const (
	// HTTPGet HTTP GET
	HTTPGet HTTPMethod = "GET"

	// HTTPPut HTTP PUT
	HTTPPut HTTPMethod = "PUT"

	// HTTPHead HTTP HEAD
	HTTPHead HTTPMethod = "HEAD"

	// HTTPPost HTTP POST
	HTTPPost HTTPMethod = "POST"

	// HTTPDelete HTTP DELETE
	HTTPDelete HTTPMethod = "DELETE"
)

// HTTP headers
const (
	HTTPHeaderAcceptEncoding     = "Accept-Encoding"
	HTTPHeaderAuthorization      = "Authorization"
	HTTPHeaderCacheControl       = "Cache-Control"
	HTTPHeaderContentDisposition = "Content-Disposition"
	HTTPHeaderContentEncoding    = "Content-Encoding"
	HTTPHeaderContentLength      = "Content-Length"
	HTTPHeaderContentMD5         = "Content-MD5"
	HTTPHeaderContentType        = "Content-Type"
	HTTPHeaderContentLanguage    = "Content-Language"
	HTTPHeaderDate               = "Date"
	HTTPHeaderEtag               = "ETag"
	HTTPHeaderExpires            = "Expires"
	HTTPHeaderHost               = "Host"
	HTTPHeaderLastModified       = "Last-Modified"
	HTTPHeaderRange              = "Range"
	HTTPHeaderLocation           = "Location"
	HTTPHeaderOrigin             = "Origin"
	HTTPHeaderServer             = "Server"
	HTTPHeaderUserAgent          = "User-Agent"
	HTTPHeaderIfModifiedSince    = "If-Modified-Since"
	HTTPHeaderIfUnmodifiedSince  = "If-Unmodified-Since"
	HTTPHeaderIfMatch            = "If-Match"
	HTTPHeaderIfNoneMatch        = "If-None-Match"
	HTTPHeaderConnection         = "Connection"

	HTTPHeaderoosACL                         = "x-amz-acl"
	HTTPHeaderoosMetaPrefix                  = "x-amz-meta-"
	HTTPHeaderoosObjectACL                   = "x-amz-object-acl"
	HTTPHeaderoosExpiration                  = "x-amz-expiration"
	HTTPHeaderoosWebsiteRedirectLoca         = "x-amz-website-redirect-location"
	HTTPHeaderoosSecurityToken               = "x-amz-security-token"
	HTTPHeaderoosCopySource                  = "x-amz-copy-source"
	HTTPHeaderoosCopySourceRange             = "x-amz-copy-source-range"
	HTTPHeaderoosCopySourceIfMatch           = "x-amz-copy-source-if-match"
	HTTPHeaderoosCopySourceIfNoneMatch       = "x-amz-copy-source-if-none-match"
	HTTPHeaderoosCopySourceIfModifiedSince   = "x-amz-copy-source-if-modified-Since"
	HTTPHeaderoosCopySourceIfUnmodifiedSince = "x-amz-copy-source-if-unmodified-Since"
	HTTPHeaderoosMetadataDirective           = "x-amz-metadata-directive"
	HTTPHeaderoosRequestID                   = "x-amz-request-id"
	HTTPHeaderoosStorageClass                = "x-amz-storage-class"
	HTTPHeaderoosRequester                   = "x-amz-request-payer"
	HTTPHeaderoosContentSHA256               = "x-amz-content-sha256"
	HTTPHeaderXamzDate                       = "x-amz-date"
	HTTPHeaderXamzLimit                      = "x-amz-limit"
	HTTPHeaderXctyunDataLocation             = "x-ctyun-data-location"
)

// HTTP Param
const (
	HTTPParamExpires           = "Expires"
	HTTPParamAccessKeyID       = "oosAccessKeyId"
	HTTPParamAWSAccessKeyID    = "AWSAccessKeyId"
	HTTPParamSignature         = "Signature"
	HTTPParamSecurityToken     = "security-token"
	HTTPParamXAmzExpires       = "X-Amz-Expires"
	HTTPParamXAmzAlgorithm     = "X-Amz-Algorithm"
	HTTPParamXAmzCredential    = "X-Amz-Credential"
	HTTPParamXAmzDate          = "X-Amz-Date"
	HTTPParamXAmzSignedHeaders = "X-Amz-SignedHeaders"
	HTTPParamXAmzSignature     = "X-Amz-Signature"
	HTTPParamEncodingType      = "URL"
)

const (
	ACCESS_KEY_ACTION        = "Action"
	CREATE_ACCESS_KEY        = "CreateAccessKey"
	DELETE_ACCESS_KEY        = "DeleteAccessKey"
	UPDATE_ACCESS_KEY        = "UpdateAccessKey"
	LIST_ACCESS_KEY          = "ListAccessKey"
	GET_ACCESS_KEY_LAST_USED = "GetAccessKeyLastUsed"
	ACCESS_KEY_ID            = "AccessKeyId"
	ACCESS_KEY_STATUS        = "Status"
	ACCESS_KEY_ISPRIMARY     = "IsPrimary"
	ACCESS_KEY_MAXITEM       = "MaxItems"
	ACCESS_KEY_MARKER        = "Marker"
	ACCESS_KEY_ACTIVE        = "Active"
	ACCESS_KEY_INACTIVE      = "Inactive"
	ACCESS_KEY_TRUE          = "true"
	ACCESS_KEY_FALSE         = "false"
	VERSION                  = "Version"
	VERSION_IAM              = "2010-05-08"
	USER_NAME                = "UserName"
)

// Other constants
const (
	MaxPartSize = 5 * 1024 * 1024 * 1024 // Max part size, 5GB
	MinPartSize = 100 * 1024             // Min part size, 100KB

	FilePermMode = os.FileMode(0664) // Default file permission

	TempFilePrefix = "oos-go-temp-" // Temp file prefix
	TempFileSuffix = ".temp"        // Temp file suffix

	CheckpointFileSuffix = ".cp" // Checkpoint file suffix

	Version = "1.9.2" // Go SDK version
)
