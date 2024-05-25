package shttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	FORM      = "application/x-www-form-urlencoded"
	JSON      = "application/json"
	XML       = "text/xml"
	HTML      = " text/html"
	MULTIPART = "multipart/form-data"
)

func Get(url string, contentType string, headers map[string]string) ([]byte, error) {
	return request(url, http.MethodGet, contentType, headers, nil)
}

func Post(url string, contentType string, headers map[string]string, body interface{}) ([]byte, error) {
	return request(url, http.MethodPost, contentType, headers, body)
}

func Put(url string, contentType string, headers map[string]string, body interface{}) ([]byte, error) {
	return request(url, http.MethodPut, contentType, headers, body)
}

func Delete(url string, contentType string, headers map[string]string, body interface{}) ([]byte, error) {
	return request(url, http.MethodDelete, contentType, headers, body)
}

func PostForm(url string, contentType string, headers map[string]string, body interface{}) ([]byte, error) {
	return request(url, http.MethodPost, contentType, headers, body)
}

func request(method, url, contentType string, headers map[string]string, body interface{}) ([]byte, error) {
	requestBody, err := formatRequestBody(body, contentType)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, err
	}

	headers["Content-Type"] = contentType
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	client := http.Client{Timeout: time.Second * 30}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return responseBody, nil
}

func formatRequestBody(body interface{}, contentType string) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	var requestBody io.Reader
	switch contentType {
	case "application/json":
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewBuffer(jsonBody)
	case "application/x-www-form-urlencoded":
		formData := url.Values{}
		if data, ok := body.(map[string]string); ok {
			for key, value := range data {
				formData.Set(key, value)
			}
		} else {
			return nil, fmt.Errorf("主体应该是map[string]string类型的表单数据")
		}
		requestBody = strings.NewReader(formData.Encode())
	case "text/xml", "application/xml":
		if str, ok := body.(string); ok {
			requestBody = strings.NewReader(str)
		} else {
			return nil, fmt.Errorf("对于XML数据，body应该是字符串类型")
		}
	default:
		return nil, fmt.Errorf("不支持的内容类型: %s", contentType)
	}
	return requestBody, nil
}
