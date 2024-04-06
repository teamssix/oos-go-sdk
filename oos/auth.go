package oos

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// headerSorter defines the key-value structure for storing the sorted data in signHeader.
type headerSorter struct {
	Keys []string
	Vals []string
}

func (conn Conn) getCanonicalQueryStringV4(params map[string]interface{}) string {
	// Sort
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Serialize
	var buf bytes.Buffer
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(conn.url.UriEncode(k, false))
		if params[k] != nil {
			buf.WriteString("=" + conn.url.UriEncode(params[k].(string), false))
		} else {
			buf.WriteString("=")
		}
	}

	return buf.String()
}

func (conn Conn) getCanonicalHeadersV4(req *http.Request) (string, string) {

	temp := make(map[string]string)
	for k, v := range req.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-amz-") {
			temp[strings.ToLower(k)] = strings.Trim(v[0], " ")
		}
		if strings.ToLower(k) == "host" {
			temp[strings.ToLower(k)] = strings.Trim(v[0], " ")
		}
		if strings.ToLower(k) == "content-type" {
			temp[strings.ToLower(k)] = strings.Trim(v[0], " ")
		}
	}
	hs := newHeaderSorter(temp)

	// Sort the temp by the ascending order
	hs.Sort()

	// Get the canonicalizedoosHeaders
	canonicalizedoosHeaders := ""
	SignedHeaders := ""
	for i := range hs.Keys {
		canonicalizedoosHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
		SignedHeaders += hs.Keys[i] + ";"
	}
	SignedHeaders = strings.TrimRight(SignedHeaders, ";")
	return canonicalizedoosHeaders, SignedHeaders
}

func (conn Conn) getHexEncodePayLoadV4(req *http.Request) string {
	return req.Header.Get(HTTPHeaderoosContentSHA256)
}

func (conn Conn) getScopeV4(req *http.Request) (string, string, string, string) {
	date := time.Now().UTC().Format("20060102")
	service := ""
	region := ""

	//req.URL.Host
	tempArray := strings.Split(conn.url.NetLoc, ".")
	for _, v := range tempArray {
		if strings.HasPrefix(v, "oos-") {
			tempChilds := strings.Split(v, "-")
			if len(tempChilds) >= 2 {
				region = tempChilds[1]
			}
			if len(tempChilds) >= 3 {
				if tempChilds[2] == "iam" {
					service = "sts"
				} else if tempChilds[2] == "cloudtrail" {
					service = "cloudtrail"
				}
			} else {
				service = "s3"
			}
			break
		}
	}
	return date + "/" + region + "/" + service + "/" + "aws4_request", date, region, service
}

func (conn Conn) hmacSha256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// signHeaderV4 signs the header in V4 format and sets it as the authorization header.
func (conn Conn) signHeaderV4(req *http.Request, dateISO string, canonicalizedResource string, params map[string]interface{}) {
	signature, scope, SignedHeaders := conn.getSignedStrV4(req, dateISO, canonicalizedResource, params, false)
	authorizationStr := "AWS4-HMAC-SHA256"
	authorizationStr += " Credential=" + conn.config.AccessKeyID + "/" + scope + ","
	authorizationStr += " SignedHeaders=" + SignedHeaders + ","
	authorizationStr += " Signature=" + signature //hex.EncodeToString(signResultStr)

	req.Header.Set(HTTPHeaderAuthorization, authorizationStr)
}

func (conn Conn) getSignedStrV4(req *http.Request, dateISO string, canonicalizedResource string, params map[string]interface{}, isSignUrl bool) (string, string, string) {
	/** 1 canonical request **/
	canonicalRequest := ""
	/*** 1.1 HTTP Verb ***/
	canonicalRequest += req.Method + "\n"

	/*** 1.2 canonical URI ***/
	canonicalURI := ""
	uris := strings.Split(canonicalizedResource, "?")
	if len(uris) > 0 {
		canonicalURI = uris[0]
	}
	canonicalRequest += canonicalURI + "\n"

	/*** 1.3 canonical Query String ***/
	canonicalRequest += conn.getCanonicalQueryStringV4(params) + "\n"

	/*** 1.4 canonical Headers ***/
	canonicalHeaders, SignedHeaders := conn.getCanonicalHeadersV4(req)
	canonicalRequest += canonicalHeaders + "\n"

	/*** 1.5 signed Headers ***/
	canonicalRequest += SignedHeaders + "\n"

	// 1.6 hash payload
	hashPayload := ""
	if isSignUrl {
		hashPayload = "UNSIGNED-PAYLOAD"
	} else {
		hashPayload = conn.getHexEncodePayLoadV4(req)
	}
	canonicalRequest += hashPayload
	// fmt.Println("canonicalRequest:" + "\n" + canonicalRequest)

	/** 2 make StringToSign **/
	stringToSign := ""

	/*** 2.1 AWS4-HMAC-SHA256 ***/
	stringToSign += "AWS4-HMAC-SHA256" + "\n"

	/*** 2.2 timestamp ***/
	stringToSign += dateISO + "\n"

	/*** 2.3 scope ***/
	scope, date, region, service := conn.getScopeV4(req)
	stringToSign += scope + "\n"

	/*** 2.4 Hex (SHA256Hash(canonical request)) ***/
	hash := sha256.New()
	hash.Write([]byte(canonicalRequest))
	hashResult := hash.Sum(nil)
	stringToSign += hex.EncodeToString(hashResult)
	// fmt.Println("stringToSign:" + "\n" + stringToSign)

	/** 3 make signature **/
	/*** 3.1 DateKey ***/
	dateKey := conn.hmacSha256([]byte("AWS4"+conn.config.AccessKeySecret), []byte(date))
	//fmt.Println(len(dateKey))
	//fmt.Println(dateKey)

	/*** 3.2 DateRegionKey ***/
	dateRegionKey := conn.hmacSha256(dateKey, []byte(region))
	//fmt.Println(len(dateRegionKey))
	//fmt.Println(dateRegionKey)

	/*** 3.3 DateRegionServiceKey ***/
	dateRegionServiceKey := conn.hmacSha256(dateRegionKey, []byte(service))
	//fmt.Println(len(dateRegionServiceKey))
	//fmt.Println(dateRegionServiceKey)

	/*** 3.4 SigningKey ***/
	signingKey := conn.hmacSha256(dateRegionServiceKey, []byte("aws4_request"))
	//fmt.Println(len(signingKey))
	//fmt.Println(signingKey)

	// sign
	signResultStr := conn.hmacSha256(signingKey, []byte(stringToSign))

	return hex.EncodeToString(signResultStr), scope, SignedHeaders
}

// signHeader signs the header and sets it as the authorization header.
func (conn Conn) signHeader(req *http.Request, canonicalizedResource string) {
	// fmt.Println(conn.getScopeV4(req))
	// Get the final authorization string
	authorizationStr := "AWS " + conn.config.AccessKeyID + ":" + conn.getSignedStr(req, canonicalizedResource)

	// Give the parameter "Authorization" value
	req.Header.Set(HTTPHeaderAuthorization, authorizationStr)
}

func (conn Conn) getSignedStr(req *http.Request, canonicalizedResource string) string {
	// Find out the "x-oos-"'s address in header of the request
	temp := make(map[string]string)

	for k, v := range req.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-amz-") {
			temp[strings.ToLower(k)] = v[0]
		}
	}
	hs := newHeaderSorter(temp)

	// Sort the temp by the ascending order
	hs.Sort()

	// Get the canonicalizedoosHeaders
	canonicalizedoosHeaders := ""
	for i := range hs.Keys {
		canonicalizedoosHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	// Give other parameters values
	// when sign URL, date is expires
	date := req.Header.Get(HTTPHeaderDate)
	contentType := req.Header.Get(HTTPHeaderContentType)
	//contentMd5 := req.Header.Get(HTTPHeaderContentMD5)
	contentMd5Value, _ := req.Header[HTTPHeaderContentMD5]
	contentMd5 := ""
	if contentMd5Value != nil {
		contentMd5 = contentMd5Value[0]
	}
	signStr := req.Method + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedoosHeaders + canonicalizedResource
	// fmt.Println("signStr:" + signStr)
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(conn.config.AccessKeySecret))
	io.WriteString(h, signStr)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signedStr
}

// newHeaderSorter is an additional function for function SignHeader.
func newHeaderSorter(m map[string]string) *headerSorter {
	hs := &headerSorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

// Sort is an additional function for function SignHeader.
func (hs *headerSorter) Sort() {
	sort.Sort(hs)
}

// Len is an additional function for function SignHeader.
func (hs *headerSorter) Len() int {
	return len(hs.Vals)
}

// Less is an additional function for function SignHeader.
func (hs *headerSorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

// Swap is an additional function for function SignHeader.
func (hs *headerSorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}
