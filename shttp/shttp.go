package shttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

var (
	ProxyUrl string
	server   = make(map[string]string, 0)
)

const (
	FORM      = "application/x-www-form-urlencoded"
	JSON      = "application/json"
	XML       = "text/xml"
	HTML      = " text/html"
	MULTIPART = "multipart/form-data"
)

func Get[T any](url string) (T, error) {
	return request[T](url, http.MethodGet, JSON, nil)
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

// 从代理服务获取地址
func GetServer(name string) (string, error) {
	if ProxyUrl == "" {
		return "", errors.New("未配置代理服务器地址")
	}
	if len(server) == 0 {
		getServerInfo()
	}
	ip := server[name]
	if ip == "" {
		getServerInfo()
	}
	ip = server[name]
	if ip == "" {
		return "", errors.New("未获取到代理服务地址")
	}
	return ip, nil
}

func getServerInfo() {
	s, err := Get[map[string]string](ProxyUrl + "/address")
	if err != nil {
		log.Printf("获取server代理地址失败")
		return
	}
	server = s
}
