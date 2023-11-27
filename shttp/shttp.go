package shttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

const (
	FORM      = "application/x-www-form-urlencoded"
	JSON      = "application/json"
	XML       = "text/xml"
	HTML      = " text/html"
	MULTIPART = "multipart/form-data"
)

func Get[T any](url string, payload []byte) (T, error) {
	return request[T](url, http.MethodGet, JSON, payload)
}

func Post[T any](url string, payload []byte) (T, error) {
	return request[T](url, http.MethodPost, JSON, payload)
}

func Put[T any](url string, payload []byte) (T, error) {
	return request[T](url, http.MethodPut, JSON, payload)
}

func Delete[T any](url string, payload []byte) (T, error) {
	return request[T](url, http.MethodDelete, JSON, payload)
}

func PostForm[T any](url string, payload []byte) (T, error) {
	return request[T](url, http.MethodPost, JSON, payload)
}

func request[T any](url string, method string, contentType string, payload []byte) (T, error) {
	client := &http.Client{}
	var response *http.Response
	var err error
	var result T

	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(payload))
	if err != nil {
		return result, err
	}
	response, err = client.Do(req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}
