package http_util

import (
	"crypto/tls"
	"time"

	"github.com/go-resty/resty/v2"
)

var client *resty.Client

func init() {

	client = resty.New()

	client.SetTimeout(10 * time.Second)

	// 忽略自签证书
	client.SetTLSClientConfig(&tls.Config{
		InsecureSkipVerify: true,
	})

	// 自动重试
	client.SetRetryCount(3)
	client.SetRetryWaitTime(1 * time.Second)

}

// HttpGet GET 请求
func HttpGet(url string, query map[string]string) ([]byte, error) {

	req := client.R()

	if query != nil {
		req.SetQueryParams(query)
	}

	resp, err := req.Get(url)

	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

// HttpPost POST JSON
func HttpPost(url string, body interface{}) ([]byte, error) {

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		//SetHeaders()
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}

// HttpPostForm POST 表单
func HttpPostForm(url string, form map[string]string) ([]byte, error) {

	resp, err := client.R().
		SetFormData(form).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.Body(), nil
}
