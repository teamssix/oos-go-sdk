package oos

import (
	"bytes"
	"errors"
	"fmt"
)

// CreateAccessKey	 Create a pair of regular AccessKey and SecretKey.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) CreateAccessKey(userName string) (CreateAccessKeyResponse, error) {
	var out CreateAccessKeyResponse

	body := fmt.Sprintf("%s=%s&%s=%s", ACCESS_KEY_ACTION, CREATE_ACCESS_KEY, VERSION, VERSION_IAM)
	if userName != "" {
		body = body + "&" + USER_NAME + "=" + userName
	}
	buffer := new(bytes.Buffer)
	buffer.Write([]byte(body))

	resp, err := client.do("POST", "", nil, nil, buffer)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// DeleteAccessKey	 Delete a pair of regular AccessKey and SecretKey.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) DeleteAccessKey(accessKeyId, userName string) (DeleteAccessKeyResponse, error) {
	var out DeleteAccessKeyResponse

	if accessKeyId == "" {
		return out, errors.New("the parameter is invalid: bucket's name is empty")
	}

	body := fmt.Sprintf("%s=%s&%s=%s&%s=%s", ACCESS_KEY_ACTION, DELETE_ACCESS_KEY, ACCESS_KEY_ID,
		accessKeyId, VERSION, VERSION_IAM)
	if userName != "" {
		body = body + "&" + USER_NAME + "=" + userName
	}
	buffer := new(bytes.Buffer)
	buffer.Write([]byte(body))

	resp, err := client.do("POST", "", nil, nil, buffer)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

func (client Client) GetAccessKeyLastUsed(accessKeyId string) (GetAccessKeyLastUsedResponse, error) {
	var out GetAccessKeyLastUsedResponse

	body := fmt.Sprintf("%s=%s&%s=%s", ACCESS_KEY_ACTION, GET_ACCESS_KEY_LAST_USED,
		ACCESS_KEY_ID, accessKeyId)

	buffer := new(bytes.Buffer)
	buffer.Write([]byte(body))

	resp, err := client.do("POST", "", nil, nil, buffer)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// ListAccessKey	 List all  pair of regular AccessKey and SecretKey.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) ListAccessKey(maxCount int, Marker, userName string) (ListAccessKeysResponse, error) {
	var out ListAccessKeysResponse

	if maxCount < 0 {
		maxCount = 100
	}
	var body string
	if Marker != "" {
		body = fmt.Sprintf("%s=%s&%s=%d&%s=%s", ACCESS_KEY_ACTION, LIST_ACCESS_KEY,
			ACCESS_KEY_MAXITEM, maxCount, ACCESS_KEY_MARKER, Marker)
	} else {
		body = fmt.Sprintf("%s=%s&%s=%d", ACCESS_KEY_ACTION, LIST_ACCESS_KEY,
			ACCESS_KEY_MAXITEM, maxCount)
	}

	if userName != "" {
		body = body + "&" + USER_NAME + "=" + userName
	}

	body = body + "&" + VERSION + "=" + VERSION_IAM

	buffer := new(bytes.Buffer)
	buffer.Write([]byte(body))

	resp, err := client.do("POST", "", nil, nil, buffer)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	err = xmlUnmarshal(resp.Body, &out)
	return out, err
}

// UpdateAccessKey	 Update  regular AccessKey 's Status.
//
// bucketName    the bucket name.
//
// error    it's nil if no error, otherwise it's an error object.
func (client Client) UpdateAccessKey(accessKeyId string, bActive bool) error {

	if accessKeyId == "" {
		return errors.New("the parameter is invalid: bucket's name is empty")
	}

	var sStatus string
	if bActive {
		sStatus = ACCESS_KEY_ACTIVE
	} else {
		sStatus = ACCESS_KEY_INACTIVE
	}
	body := fmt.Sprintf("%s=%s&%s=%s&%s=%s&%s=%s", ACCESS_KEY_ACTION, UPDATE_ACCESS_KEY,
		ACCESS_KEY_ID, accessKeyId, ACCESS_KEY_STATUS, sStatus, VERSION, VERSION_IAM)
	buffer := new(bytes.Buffer)
	buffer.Write([]byte(body))

	resp, err := client.do("POST", "", nil, nil, buffer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return err
}
