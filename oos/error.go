package oos

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
)

// ServiceError contains fields of the error response from oos Service REST API.
type ServiceError struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`      // The error code returned from oos to the caller
	Message    string   `xml:"Message"`   // The detail error message from oos
	RequestID  string   `xml:"RequestId"` // The UUID used to uniquely identify the request
	HostID     string   `xml:"HostId"`    // The oos server cluster's Id
	Resource   string   `xml:"Resource"`
	RawMessage string   // The raw messages from oos
	StatusCode int      // HTTP status code
}

// Error implements interface error
func (e ServiceError) Error() string {
	return fmt.Sprintf("oos: service returned error: StatusCode=%d, ErrorCode=%s, ErrorMessage=%s, RequestId=%s , Resource=%s",
		e.StatusCode, e.Code, e.Message, e.RequestID, e.Resource)
}

// UnexpectedStatusCodeError is returned when a storage service responds with neither an error
// nor with an HTTP status code indicating success.
type UnexpectedStatusCodeError struct {
	allowed []int // The expected HTTP stats code returned from oos
	got     int   // The actual HTTP status code from oos
}

// Error implements interface error
func (e UnexpectedStatusCodeError) Error() string {
	s := func(i int) string { return fmt.Sprintf("%d %s", i, http.StatusText(i)) }

	got := s(e.got)
	expected := []string{}
	for _, v := range e.allowed {
		expected = append(expected, s(v))
	}
	return fmt.Sprintf("oos: status code from service response is %s; was expecting %s",
		got, strings.Join(expected, " or "))
}

// Got is the actual status code returned by oos.
func (e UnexpectedStatusCodeError) Got() int {
	return e.got
}

// checkRespCode returns UnexpectedStatusError if the given response code is not
// one of the allowed status codes; otherwise nil.
func checkRespCode(respCode int, allowed []int) error {
	for _, v := range allowed {
		if respCode == v {
			return nil
		}
	}
	return UnexpectedStatusCodeError{allowed, respCode}
}
