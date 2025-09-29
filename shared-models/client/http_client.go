package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"shared-models/common"
)

// HTTPClient 封装的 HTTP 客户端
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
}

// NewHTTPClient 创建新的 HTTP 客户端
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		headers: make(map[string]string),
	}
}

// SetHeader 设置请求头
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// Post 发送 POST 请求
func (c *HTTPClient) Post(ctx context.Context, endpoint string, body interface{}) (*common.BaseResponse, error) {
	return c.request(ctx, http.MethodPost, endpoint, body)
}

// Get 发送 GET 请求
func (c *HTTPClient) Get(ctx context.Context, endpoint string) (*common.BaseResponse, error) {
	return c.request(ctx, http.MethodGet, endpoint, nil)
}

// request 统一请求方法
func (c *HTTPClient) request(ctx context.Context, method, endpoint string, body interface{}) (*common.BaseResponse, error) {
	url := c.baseURL + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置默认头部
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义头部
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP 错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var response common.BaseResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &response, nil
}
