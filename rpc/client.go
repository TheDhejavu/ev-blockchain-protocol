package rpc

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client struct {
	url string
}

type Response struct {
	Status string `json:"status"`
	Body   *Body  `json:"body"`
}

type Body struct {
	JsonRpc string `json:"jsonrpc"`
	Error   Error  `json:"error"`
	Result  Result `json:"result"`
}

type Result struct {
	Data interface{} `json:"data"`
}

type Error struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func NewClient(url string) *Client {
	return &Client{url}
}

func (c *Client) Do(method string, params interface{}) (Response, error) {
	request := c.NewRequest(method, params)
	jsonStr, err := json.Marshal(request)
	if err != nil {
		return Response{}, err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return Response{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	response := Response{
		resp.Status,
		&Body{},
	}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(response.Body)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (c *Client) NewRequest(method string, params interface{}) map[string]interface{} {
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "qwe",
		"params":  params,
		"method":  method,
	}
}
