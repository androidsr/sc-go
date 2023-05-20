package hcli

import (
	"bytes"
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	Json      ContentType = "application/json"
	Form      ContentType = "application/x-www-form-urlencoded"
	Multipart ContentType = "multipart/form-data"
	Xml       ContentType = "text/xml"
	Html      ContentType = "text/html"
)

var (
	c *resty.Client
)

type ContentType string

func New(headers, cookies map[string]string) *HClient {
	if c == nil {
		c = resty.New()
	}
	client := &HClient{c, headers, cookies}
	return client
}

type HClient struct {
	*resty.Client
	headers map[string]string
	cookies map[string]string
}

func (m *HClient) Get(dest interface{}, url string, params map[string]string) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	r.SetQueryParams(params)
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	resp, err := r.Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Post(dest interface{}, url string, contentType ContentType, body interface{}) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	r.SetHeader("Content-Type", string(contentType))
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	r.SetBody(body)
	resp, err := r.Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Put(dest interface{}, url string, contentType ContentType, body interface{}) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	r.SetHeader("Content-Type", string(contentType))
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	r.SetBody(body)
	resp, err := r.Put(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Delete(dest interface{}, url string, params map[string]string) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	r.SetQueryParams(params)
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	resp, err := r.Delete(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Patch(dest interface{}, url string, contentType ContentType, body interface{}) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	r.SetHeader("Content-Type", string(contentType))
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	r.SetBody(body)
	resp, err := r.Patch(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Head(dest interface{}, url string, body interface{}) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}

	resp, err := r.Head(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) Options(dest interface{}, url string, body interface{}) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	resp, err := r.Options(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}

func (m *HClient) UploadFile(dest interface{}, url string, fileName string, data []byte, params map[string]string) ([]byte, error) {
	r := m.R()
	if dest != nil {
		r.SetResult(dest)
	}
	r.SetFileReader("file", fileName, bytes.NewBuffer(data))
	r.SetFormData(params)
	for k, v := range m.headers {
		r.Header.Add(k, v)
	}
	for k, v := range m.cookies {
		r.SetCookie(&http.Cookie{Name: k, Value: v})
	}
	resp, err := r.Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), err
}
