package vk

import (
	"fmt"
	"log"

	"github.com/Nakray/sn/internal/database"
)

type Collector struct {
	client *Client
	db     *database.DB
}

func NewCollector(client *Client, db *database.DB) *Collector {
	return &Collector{
		client: client,
		db:     db,
	}
}

func (col *Collector) CollectUser(userID int64) error {
	log.Printf("Collecting user %d\n", userID)

	// Get user info
	userInfo, err := col.client.GetUserInfo(userID)
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	// Save user object
	owner := database.Owner{
		Type: database.OwnerTypeUser,
		ID:   userID,
	}

	if err := col.db.WriteObject("vkontakte", owner, "user", nil, userInfo); err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Get friends
	friends, err := col.client.GetFriends(userID)
	if err != nil {
		log.Printf("Failed to get friends for user %d: %v\n", userID, err)
	} else {
		if err := col.db.WriteRelations("vkontakte", owner, database.RelationTypeFriend, nil, friends); err != nil {
			log.Printf("Failed to save friends: %v\n", err)
		}
	}

	// Get groups
	groups, err := col.client.GetGroups(userID)
	if err != nil {
		log.Printf("Failed to get groups for user %d: %v\n", userID, err)
	} else {
		if err := col.db.WriteRelations("vkontakte", owner, database.RelationTypeGroup, nil, groups); err != nil {
			log.Printf("Failed to save groups: %v\n", err)
		}
	}

	// Get wall posts
	posts, err := col.client.GetWallPosts(userID, 100)
	if err != nil {
		log.Printf("Failed to get wall posts for user %d: %v\n", userID, err)
	} else {
		var postIDs []int64
		for _, post := range posts {
			if id, ok := post["id"].(float64); ok {
				postIDs = append(postIDs, int64(id))
				// Save post object
				postOwner := database.Owner{Type: database.OwnerTypeUser, ID: userID}
				postDetails := map[string]interface{}{"id": int64(id)}
				col.db.WriteObject("vkontakte", postOwner, "post", postDetails, post)
			}
		}
		if len(postIDs) > 0 {
			col.db.WriteRelations("vkontakte", owner, database.RelationTypePost, nil, postIDs)
		}
	}

		// Collect likes for posts
	for _, postID := range postIDs {
		likes, err := col.client.GetLikes(userID, postID, "post", 1000)
		if err != nil {
			log.Printf("Failed to get likes for post %d: %v\n", postID, err)
		} else {
			likeDetails := map[string]interface{}{"post_id": postID}
			if err := col.db.WriteRelations("vkontakte", owner, database.RelationTypeLike, likeDetails, likes); err != nil {
				log.Printf("Failed to save likes for post %d: %v\n", postID, err)
			}
		}
	}

	// Get followers
	followers, err := col.client.GetFollowers(userID, 1000)
	if err != nil {
		log.Printf("Failed to get followers for user %d: %v\n", userID, err)
	} else {
		if err := col.db.WriteRelations("vkontakte", owner, database.RelationTypeFollower, nil, followers); err != nil {
			log.Printf("Failed to save followers: %v\n", err)
		}
	}

	// Get photos
	photos, err := col.client.GetPhotos(userID, "profile", 100)
	if err != nil {
		log.Printf("Failed to get photos for user %d: %v\n", userID, err)
	} else {
		var photoIDs []int64
		for _, photo := range photos {
			if id, ok := photo["id"].(float64); ok {
				photoIDs = append(photoIDs, int64(id))
				photoOwner := database.Owner{Type: database.OwnerTypeUser, ID: userID}
				photoDetails := map[string]interface{}{"id": int64(id)}
				col.db.WriteObject("vkontakte", photoOwner, "photo", photoDetails, photo)
			}
		}
		if len(photoIDs) > 0 {
			col.db.WriteRelations("vkontakte", owner, database.RelationTypePhoto, nil, photoIDs)
		
		
			// Collect likes for photos
	for _, photoID := range photoIDs {
		likes, err := col.client.GetLikes(userID, photoID, "photo", 1000)
		if err != nil {
			log.Printf("Failed to get likes for photo %d: %v\n", photoID, err)
		} else {
			likeDetails := map[string]interface{}{"photo_id": photoID}
			if err := col.db.WriteRelations("vkontakte", owner, database.RelationTypeLike, likeDetails, likes); err != nil {
				log.Printf("Failed to save likes for photo %d: %v\n", photoID, err)
			}
		}
	}}
	}

	return nil
}

func (col *Collector) CollectGroup(groupID int64) error {
	log.Printf("Collecting group %d\n", groupID)

	// Get group info
	groupInfo, err := col.client.GetGroupInfo(groupID)
	if err != nil {
		return fmt.Errorf("failed to get group info: %w", err)
	}

	// Save group object
	owner := database.Owner{
		Type: database.OwnerTypeGroup,
		ID:   groupID,
	}

	if err := col.db.WriteObject("vkontakte", owner, "group", nil, groupInfo); err != nil {
		return fmt.Errorf("failed to save group: %w", err)
	}

	return nil
}

func (col *Collector) CollectEntity(ownerType database.OwnerType, ownerID int64) error {
	switch ownerType {
	case database.OwnerTypeUser:
		return col.CollectUser(ownerID)
	case database.OwnerTypeGroup:
		return col.CollectGroup(ownerID)
	default:
		return fmt.Errorf("unknown owner type: %s", ownerType)
	}
}
