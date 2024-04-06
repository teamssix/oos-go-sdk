package oos

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Conn defines oos Conn
type Conn struct {
	config *Config
	url    *urlMaker
	client *http.Client
}

//var signKeyList = []string{"acl", "uploads", "location", "cors", "logging", "website", "referer", "lifecycle", "delete", "append", "tagging", "objectMeta", "uploadId", "partNumber", "security-token", "position", "img", "style", "styleName", "replication", "replicationProgress", "replicationLocation", "cname", "bucketInfo", "comp", "qos", "live", "status", "vod", "startTime", "endTime", "symlink", "x-oos-process", "response-content-type", "response-content-language", "response-expires", "response-cache-control", "response-content-disposition", "response-content-encoding", "udf", "udfName", "udfImage", "udfId", "udfImageDesc", "udfApplication", "comp", "udfApplicationLog", "restore", "callback", "callback-var"}

var signKeyList = []string{"acl", "torrent", "logging", "location", "policy", "requestPayment", "versioning",
	"versions", "versionId", "notification", "uploadId", "uploads", "partNumber", "website",
	"delete", "lifecycle", "tagging", "cors", "restore", "response-cache-control", "response-content-disposition",
	"response-content-type", "response-content-language", "response-content-encoding", "response-expires"}

// init initializes Conn
func (conn *Conn) init(config *Config, urlMaker *urlMaker) error {
	// New transport
	transport := newTransport(conn, config)

	conn.config = config
	conn.url = urlMaker
	conn.client = &http.Client{Transport: transport}

	return nil
}

// Do sends request and returns the response
func (conn Conn) Do(method, bucketName, objectName string, params map[string]interface{}, headers map[string]string,
	data io.Reader, listener ProgressListener) (*Response, error) {
	urlParams := conn.getURLParams(params)
	subResource := conn.getSubResource(params)
	uri := conn.url.getURL(bucketName, objectName, urlParams)
	canonResource := ""
	if conn.config.IsV4Sign {
		canonResource = conn.url.getResourceV4(bucketName, objectName, subResource)
	} else {
		canonResource = conn.url.getResource(bucketName, objectName, subResource)
	}

	return conn.doRequest(method, uri, canonResource, headers, params, data, listener)
}

// DoURL sends the request with signed URL and returns the response result.
func (conn Conn) DoURL(method HTTPMethod, signedURL string, headers map[string]string,
	data io.Reader, initCRC uint64, listener ProgressListener) (*Response, error) {
	// Get URI from signedURL
	uri, err := url.ParseRequestURI(signedURL)
	if err != nil {
		return nil, err
	}

	m := strings.ToUpper(string(method))
	req := &http.Request{
		Method:     m,
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}

	tracker := &readerTracker{completedBytes: 0}
	fd := conn.handleBody(req, data, listener, tracker, true)
	if fd != nil {
		defer func() {
			fd.Close()
			os.Remove(fd.Name())
		}()
	}

	req.Header.Set(HTTPHeaderHost, conn.config.Endpoint)
	req.Header.Set(HTTPHeaderUserAgent, conn.config.UserAgent)

	for k, v := range headers {
		req.Header.Set(k, v)
		//req.Header[k] = []string{v}
	}

	// Transfer started
	event := newProgressEvent(TransferStartedEvent, 0, req.ContentLength)
	publishProgress(listener, event)

	resp, err := conn.client.Do(req)
	if err != nil {
		// Transfer failed
		event = newProgressEvent(TransferFailedEvent, tracker.completedBytes, req.ContentLength)
		publishProgress(listener, event)
		return nil, err
	}

	// Transfer completed
	event = newProgressEvent(TransferCompletedEvent, tracker.completedBytes, req.ContentLength)
	publishProgress(listener, event)

	return conn.handleResponse(resp)
}

func (conn Conn) getURLParams(params map[string]interface{}) string {
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
		}
	}

	return buf.String()
}

func (conn Conn) getSubResource(params map[string]interface{}) string {
	// Sort
	keys := make([]string, 0, len(params))
	for k := range params {
		if conn.isParamSign(k) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Serialize
	var buf bytes.Buffer
	for _, k := range keys {
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		if params[k] != nil {
			buf.WriteString("=" + params[k].(string))
		}
	}

	return buf.String()
}

func (conn Conn) isParamSign(paramKey string) bool {
	for _, k := range signKeyList {
		if paramKey == k {
			return true
		}
	}
	return false
}

func (conn Conn) doRequest(method string, uri *url.URL, canonicalizedResource string, headers map[string]string,
	params map[string]interface{}, data io.Reader, listener ProgressListener) (*Response, error) {
	method = strings.ToUpper(method)
	req := &http.Request{
		Method:     method,
		URL:        uri,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       uri.Host,
	}

	tracker := &readerTracker{completedBytes: 0}
	fd := conn.handleBody(req, data, listener, tracker, false)
	if fd != nil {
		defer func() {
			fd.Close()
			os.Remove(fd.Name())
		}()
	}

	date := ""
	if conn.config.IsV4Sign {
		date = time.Now().UTC().Format("20060102T150405Z")
		//date = "20220914T075027Z"
		req.Header.Set(HTTPHeaderXamzDate, date)
	} else {
		date = time.Now().UTC().Format(http.TimeFormat)
		req.Header.Set(HTTPHeaderDate, date)
	}

	req.Header.Set(HTTPHeaderHost, conn.config.Endpoint)
	req.Header.Set(HTTPHeaderUserAgent, conn.config.UserAgent)
	if conn.config.SecurityToken != "" {
		req.Header.Set(HTTPHeaderoosSecurityToken, conn.config.SecurityToken)
	}

	for k, v := range headers {
		req.Header[k] = []string{v}
	}

	if conn.config.IsV4Sign {
		conn.signHeaderV4(req, date, canonicalizedResource, params)
	} else {
		conn.signHeader(req, canonicalizedResource)
	}

	// Transfer started
	event := newProgressEvent(TransferStartedEvent, 0, req.ContentLength)
	publishProgress(listener, event)

	// for k, v := range req.Header {
	// 	fmt.Println(k + ":" + v[0])
	// }

	resp, err := conn.client.Do(req)
	if err != nil {
		// Transfer failed
		event = newProgressEvent(TransferFailedEvent, tracker.completedBytes, req.ContentLength)
		publishProgress(listener, event)
		return nil, err
	}

	// Transfer completed
	event = newProgressEvent(TransferCompletedEvent, tracker.completedBytes, req.ContentLength)
	publishProgress(listener, event)

	return conn.handleResponse(resp)
}

func (conn Conn) signURL(method HTTPMethod, bucketName, objectName string, expiredInSec int64, params map[string]interface{}, headers map[string]string) string {
	if conn.config.SecurityToken != "" {
		params[HTTPParamSecurityToken] = conn.config.SecurityToken
	}
	subResource := conn.getSubResource(params)

	canonResource := ""
	if conn.config.IsV4Sign {
		canonResource = conn.url.getResourceV4(bucketName, objectName, subResource)
	} else {
		canonResource = conn.url.getResource(bucketName, objectName, subResource)
	}

	m := strings.ToUpper(string(method))
	req := &http.Request{
		Method: m,
		Header: make(http.Header),
	}

	date := ""
	if conn.config.IsV4Sign {
		date = time.Now().UTC().Format("20060102T150405Z")
	} else {
		expiration := time.Now().UTC().Unix() + expiredInSec
		date = strconv.FormatInt(expiration, 10)
		req.Header.Set(HTTPHeaderDate, date)
	}

	req.Header.Set(HTTPHeaderHost, conn.config.Endpoint)
	req.Header.Set(HTTPHeaderUserAgent, conn.config.UserAgent)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	_, SignedHeaders := conn.getCanonicalHeadersV4(req)
	scope, _, _, _ := conn.getScopeV4(req)

	if conn.config.IsV4Sign {
		params[HTTPParamXAmzAlgorithm] = "AWS4-HMAC-SHA256"
		params[HTTPParamXAmzExpires] = strconv.FormatInt(expiredInSec, 10)
		params[HTTPParamXAmzCredential] = conn.config.AccessKeyID + "/" + scope
		params[HTTPParamXAmzSignedHeaders] = SignedHeaders
		params[HTTPParamXAmzDate] = date
	} else {
		params[HTTPParamExpires] = date
		params[HTTPParamAWSAccessKeyID] = conn.config.AccessKeyID
	}

	signedStr := ""
	if conn.config.IsV4Sign {
		signedStr, _, _ = conn.getSignedStrV4(req, date, canonResource, params, true)
		params[HTTPParamXAmzSignature] = signedStr
	} else {
		signedStr = conn.getSignedStr(req, canonResource)
		params[HTTPParamSignature] = signedStr
	}

	urlParams := conn.getURLParams(params)
	return conn.url.getSignURL(bucketName, objectName, urlParams)
}

// handleBody handles request body
func (conn Conn) handleBody(req *http.Request, body io.Reader,
	listener ProgressListener, tracker *readerTracker, isSignUrl bool) *os.File {
	var file *os.File
	reader := body

	// Length
	switch v := body.(type) {
	case *bytes.Buffer:
		req.ContentLength = int64(v.Len())
	case *bytes.Reader:
		req.ContentLength = int64(v.Len())
	case *strings.Reader:
		req.ContentLength = int64(v.Len())
	case *os.File:
		req.ContentLength = tryGetFileSize(v)
	case *io.LimitedReader:
		req.ContentLength = int64(v.N)
	}
	req.Header.Set(HTTPHeaderContentLength, strconv.FormatInt(req.ContentLength, 10))

	// MD5
	if body != nil && req.Header.Get(HTTPHeaderContentMD5) == "" {
		md5 := ""
		reader, md5, file, _ = calcMD5(body, req.ContentLength, conn.config.MD5Threshold)
		req.Header[HTTPHeaderContentMD5] = []string{md5}
	}
	body = reader
	if isSignUrl == false && conn.config.IsV4Sign == true {
		// v4 SHA258
		if req.Header.Get(HTTPHeaderoosContentSHA256) == "" {
			if conn.config.IsEnableSHA256 {
				sha256 := ""
				if body != nil {
					reader, sha256, file, _ = calcSha256Hash(body, req.ContentLength, conn.config.SHA256Threshold)
				} else {
					sha256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
				}
				req.Header.Set(HTTPHeaderoosContentSHA256, sha256)
			} else {
				req.Header.Set(HTTPHeaderoosContentSHA256, "UNSIGNED-PAYLOAD")
			}
		}
	}

	// HTTP body
	rc, ok := reader.(io.ReadCloser)
	if !ok && reader != nil {
		rc = ioutil.NopCloser(reader)
	}
	req.Body = rc

	return file
}

func tryGetFileSize(f *os.File) int64 {
	fInfo, _ := f.Stat()
	return fInfo.Size()
}

// handleResponse handles response
func (conn Conn) handleResponse(resp *http.Response) (*Response, error) {

	statusCode := resp.StatusCode
	if statusCode >= 400 && statusCode <= 505 {
		// 4xx and 5xx indicate that the operation has error occurred
		var respBody []byte
		respBody, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}

		if len(respBody) == 0 {
			// No error in response body
			err = fmt.Errorf("oos: service returned empty response body, status = %s, RequestId = %s", resp.Status, resp.Header.Get(HTTPHeaderoosRequestID))
		} else {
			// Response contains storage service error object, unmarshal
			srvErr, errIn := serviceErrFromXML(respBody, resp.StatusCode,
				resp.Header.Get(HTTPHeaderoosRequestID))
			if errIn != nil { // error unmarshaling the error response
				err = fmt.Errorf("oos: service returned invalid response body, status = %s, RequestId = %s", resp.Status, resp.Header.Get(HTTPHeaderoosRequestID))
			} else {
				err = srvErr
			}
		}

		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       ioutil.NopCloser(bytes.NewReader(respBody)), // restore the body
		}, err
	} else if statusCode >= 300 && statusCode <= 307 {
		// oos use 3xx, but response has no body
		err := fmt.Errorf("oos: service returned %d,%s", resp.StatusCode, resp.Status)
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, err
	}

	// 2xx, successful
	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       resp.Body,
	}, nil
}

func calcMD5(body io.Reader, contentLen, md5Threshold int64) (reader io.Reader, b64 string, tempFile *os.File, err error) {
	if contentLen == 0 || contentLen > md5Threshold {
		// Huge body, use temporary file
		tempFile, err = ioutil.TempFile(os.TempDir(), TempFilePrefix)
		if tempFile != nil {
			io.Copy(tempFile, body)
			tempFile.Seek(0, os.SEEK_SET)
			md5 := md5.New()
			io.Copy(md5, tempFile)
			sum := md5.Sum(nil)
			b64 = base64.StdEncoding.EncodeToString(sum[:])
			tempFile.Seek(0, os.SEEK_SET)
			reader = tempFile
		}
	} else {
		// Small body, use memory
		buf, _ := ioutil.ReadAll(body)
		sum := md5.Sum(buf)
		b64 = base64.StdEncoding.EncodeToString(sum[:])
		reader = bytes.NewReader(buf)
	}
	return
}
func calcSha256Hash(body io.Reader, contentLen, sha256Threshold int64) (reader io.Reader, b64 string, tempFile *os.File, err error) {
	if contentLen == 0 || contentLen > sha256Threshold {
		// Huge body, use temporary file
		tempFile, err = ioutil.TempFile(os.TempDir(), TempFilePrefix)
		if tempFile != nil {
			io.Copy(tempFile, body)
			tempFile.Seek(0, os.SEEK_SET)

			hash := sha256.New()
			io.Copy(hash, tempFile)
			hashResult := hash.Sum(nil)
			b64 += hex.EncodeToString(hashResult)

			tempFile.Seek(0, os.SEEK_SET)
			reader = tempFile
		}
	} else {
		// Small body, use memory
		buf, _ := ioutil.ReadAll(body)
		hash := sha256.New()
		hash.Write(buf)
		sum := hash.Sum(nil)
		b64 = hex.EncodeToString(sum)
		reader = bytes.NewReader(buf)
	}
	return
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}

func serviceErrFromXML(body []byte, statusCode int, requestID string) (ServiceError, error) {
	var storageErr ServiceError

	if err := xml.Unmarshal(body, &storageErr); err != nil {
		return storageErr, err
	}

	storageErr.StatusCode = statusCode
	storageErr.RequestID = requestID
	storageErr.RawMessage = string(body)
	return storageErr, nil
}

func xmlUnmarshal(body io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

func jsonUnmarshal(body io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// timeoutConn handles HTTP timeout
type timeoutConn struct {
	conn        net.Conn
	timeout     time.Duration
	longTimeout time.Duration
}

func newTimeoutConn(conn net.Conn, timeout time.Duration, longTimeout time.Duration) *timeoutConn {
	conn.SetReadDeadline(time.Now().Add(longTimeout))
	return &timeoutConn{
		conn:        conn,
		timeout:     timeout,
		longTimeout: longTimeout,
	}
}

func (c *timeoutConn) Read(b []byte) (n int, err error) {
	c.SetReadDeadline(time.Now().Add(c.timeout))
	n, err = c.conn.Read(b)
	c.SetReadDeadline(time.Now().Add(c.longTimeout))
	return n, err
}

func (c *timeoutConn) Write(b []byte) (n int, err error) {
	c.SetWriteDeadline(time.Now().Add(c.timeout))
	n, err = c.conn.Write(b)
	c.SetReadDeadline(time.Now().Add(c.longTimeout))
	return n, err
}

func (c *timeoutConn) Close() error {
	return c.conn.Close()
}

func (c *timeoutConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *timeoutConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *timeoutConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *timeoutConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *timeoutConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// UrlMaker builds URL and resource
const (
	urlTypeCname = 1
	urlTypeIP    = 2
	urlTypeOOS   = 3
)

type urlMaker struct {
	Scheme string // HTTP or HTTPS
	NetLoc string // Host or IP
	Type   int    // 1 CNAME, 2 IP, 3 OOS
}

// Init parses endpoint
func (um *urlMaker) Init(endpoint string, isCname bool) {
	if strings.HasPrefix(endpoint, "http://") {
		um.Scheme = "http"
		um.NetLoc = endpoint[len("http://"):]
	} else if strings.HasPrefix(endpoint, "https://") {
		um.Scheme = "https"
		um.NetLoc = endpoint[len("https://"):]
	} else {
		um.Scheme = "http"
		um.NetLoc = endpoint
	}

	host, _, err := net.SplitHostPort(um.NetLoc)
	if err != nil {
		host = um.NetLoc
		if host[0] == '[' && host[len(host)-1] == ']' {
			host = host[1 : len(host)-1]
		}
	}

	ip := net.ParseIP(host)
	if ip != nil {
		um.Type = urlTypeIP
	} else if isCname {
		um.Type = urlTypeCname
	} else {
		um.Type = urlTypeOOS
	}
}

// getURL gets URL
func (um urlMaker) getURL(bucket, object, params string) *url.URL {
	host, path := um.buildURL(bucket, object)
	addr := ""
	if params == "" {
		addr = fmt.Sprintf("%s://%s%s", um.Scheme, host, path)
	} else {
		addr = fmt.Sprintf("%s://%s%s?%s", um.Scheme, host, path, params)
	}
	uri, _ := url.ParseRequestURI(addr)
	return uri
}

// getSignURL gets sign URL
func (um urlMaker) getSignURL(bucket, object, params string) string {
	host, path := um.buildURL(bucket, object)
	return fmt.Sprintf("%s://%s%s?%s", um.Scheme, host, path, params)
}
func (um urlMaker) shouldEncode(c byte, isObject bool) bool {
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	switch c {
	case '/':
		if isObject {
			return false
		} else {
			return true
		}
	case '-', '_', '.', '~':
		return false
	}
	return true
}
func (um urlMaker) UriEncode(s string, isObject bool) string {
	spaceCount, hexCount := 0, 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if um.shouldEncode(c, isObject) {
			if c == ' ' {
				spaceCount++
			} else {
				hexCount++
			}
		}
	}

	if spaceCount == 0 && hexCount == 0 {
		return s
	}

	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case c == ' ':
			t[j] = '%'
			t[j+1] = '2'
			t[j+2] = '0'
			j += 3
		case c == '/' && isObject == false:
			t[j] = '%'
			t[j+1] = '2'
			t[j+2] = 'F'
			j += 3
		case um.shouldEncode(c, isObject):
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		default:
			t[j] = s[i]
			j++
		}
	}
	return string(t)
}

// buildURL builds URL
func (um urlMaker) buildURL(bucket, object string) (string, string) {
	var host = ""
	var path = ""

	object = um.UriEncode(object, true)

	if um.Type == urlTypeCname {
		host = um.NetLoc
		path = "/" + object
	} else if um.Type == urlTypeIP {
		if bucket == "" {
			host = um.NetLoc
			path = "/"
		} else {
			host = um.NetLoc
			path = fmt.Sprintf("/%s/%s", bucket, object)
		}
	} else {
		if bucket == "" {
			host = um.NetLoc
			path = "/"
		} else {
			host = um.NetLoc
			path = "/"
			if bucket != "" {
				path += bucket
			}
			if object != "" {
				path += "/" + object
			}
		}
	}

	return host, path
}

// getResource gets canonicalized resource
func (um urlMaker) getResource(bucketName, objectName, subResource string) string {
	resource := ""
	if bucketName != "" {
		resource += "/" + bucketName
	}
	if objectName != "" {
		objectName = um.UriEncode(objectName, true)
		resource += "/" + objectName
	}
	if subResource != "" {
		resource += "?" + subResource
	}
	if resource == "" {
		resource += "/"
	}
	return resource
}

func (um urlMaker) getResourceV4(bucketName, objectName, subResource string) string {
	resource := ""
	if bucketName != "" {
		resource += "/" + bucketName
	}
	if objectName != "" {
		objectName = um.UriEncode(objectName, true)
		resource += "/" + objectName
	}
	if resource == "" {
		resource += "/"
	}
	return resource
}
