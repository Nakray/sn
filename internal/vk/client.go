package vk

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	APIEndpoint = "https://api.vk.com/method/"
	APIVersion  = "5.131"
)

type Client struct {
	accessToken string
	version     string
	client      *http.Client
	lastRequest time.Time
}

type APIResponse struct {
	Response json.RawMessage `json:"response"`
	Error    *APIError       `json:"error"`
}

type APIError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func NewClient(accessToken string, proxyURL *string) (*Client, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if proxyURL != nil {
		proxy, err := url.Parse(*proxyURL)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}

	return &Client{
		accessToken: accessToken,
		version:     APIVersion,
		client:      client,
	}, nil
}

func (c *Client) Call(method string, params map[string]string) (json.RawMessage, error) {
	// Rate limiting: ~3 requests per second
	since := time.Since(c.lastRequest)
	if since < 350*time.Millisecond {
		time.Sleep(350*time.Millisecond - since)
	}
	c.lastRequest = time.Now()

	if params == nil {
		params = make(map[string]string)
	}
	params["access_token"] = c.accessToken
	params["v"] = c.version

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := c.client.PostForm(APIEndpoint+method, formData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, err
	}

	if apiResp.Error != nil {
		return nil, fmt.Errorf("VK API error %d: %s", apiResp.Error.ErrorCode, apiResp.Error.ErrorMsg)
	}

	return apiResp.Response, nil
}

func (c *Client) GetUserInfo(userID int64) (map[string]interface{}, error) {
	params := map[string]string{
		"user_ids": strconv.FormatInt(userID, 10),
		"fields":   "sex,bdate,city,country,photo_max,status,last_seen",
	}

	resp, err := c.Call("users.get", params)
	if err != nil {
		return nil, err
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(resp, &users); err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return users[0], nil
}

func (c *Client) GetGroupInfo(groupID int64) (map[string]interface{}, error) {
	params := map[string]string{
		"group_id": strconv.FormatInt(groupID, 10),
		"fields":   "description,members_count,city,country",
	}

	resp, err := c.Call("groups.getById", params)
	if err != nil {
		return nil, err
	}

	var groups []map[string]interface{}
	if err := json.Unmarshal(resp, &groups); err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("group not found")
	}

	return groups[0], nil
}

func (c *Client) GetFriends(userID int64) ([]int64, error) {
	params := map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	}

	resp, err := c.Call("friends.get", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []int64 `json:"items"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) GetGroups(userID int64) ([]int64, error) {
	params := map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
	}

	resp, err := c.Call("groups.get", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []int64 `json:"items"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}
