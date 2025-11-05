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
