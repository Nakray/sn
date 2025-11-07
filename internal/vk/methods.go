package vk

import (
	"encoding/json"
	"strconv"
)

type WallPost struct {
	ID       int64                  `json:"id"`
	OwnerID  int64                  `json:"owner_id"`
	FromID   int64                  `json:"from_id"`
	Date     int64                  `json:"date"`
	Text     string                 `json:"text"`
	Comments map[string]interface{} `json:"comments"`
	Likes    map[string]interface{} `json:"likes"`
}

func (c *Client) GetWallPosts(ownerID int64, count int) ([]map[string]interface{}, error) {
	params := map[string]string{
		"owner_id": strconv.FormatInt(ownerID, 10),
		"count":    strconv.Itoa(count),
		"filter":   "all",
	}

	resp, err := c.Call("wall.get", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) GetPhotos(ownerID int64, albumID string, count int) ([]map[string]interface{}, error) {
	params := map[string]string{
		"owner_id":  strconv.FormatInt(ownerID, 10),
		"album_id":  albumID,
		"count":     strconv.Itoa(count),
		"photo_sizes": "1",
	}

	resp, err := c.Call("photos.get", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) GetPhotoAlbums(ownerID int64) ([]map[string]interface{}, error) {
	params := map[string]string{
		"owner_id": strconv.FormatInt(ownerID, 10),
	}

	resp, err := c.Call("photos.getAlbums", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result.Items, nil
}

func (c *Client) GetFollowers(userID int64, count int) ([]int64, error) {
	params := map[string]string{
		"user_id": strconv.FormatInt(userID, 10),
		"count":   strconv.Itoa(count),
	}

	resp, err := c.Call("users.getFollowers", params)
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

func (c *Client) GetGroupMembers(groupID int64, count int) ([]int64, error) {
	params := map[string]string{
		"group_id": strconv.FormatInt(groupID, 10),
		"count":    strconv.Itoa(count),
	}

	resp, err := c.Call("groups.getMembers", params)
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

func (c *Client) GetLikes(ownerID int64, itemID int64, itemType string, count int) ([]int64, error) {
	params := map[string]string{
		"owner_id": strconv.FormatInt(ownerID, 10),
		"item_id":  strconv.FormatInt(itemID, 10),
		"type":     itemType,
		"count":    strconv.Itoa(count),
	}

	resp, err := c.Call("likes.getList", params)
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
