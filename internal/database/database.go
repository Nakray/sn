package database

import (
	"database/sql"
	"encoding/json"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

type OwnerType string

const (
	OwnerTypeUser  OwnerType = "user"
	OwnerTypeGroup OwnerType = "group"
)

type Owner struct {
	Type OwnerType
	ID   int64
}

type RelationType string

const (
	RelationTypeFriend        RelationType = "friend"
	RelationTypeFollower      RelationType = "follower"
	RelationTypeGroup         RelationType = "group"
	RelationTypePost          RelationType = "post"
	RelationTypePhoto         RelationType = "photo"
	RelationTypePostLike      RelationType = "post.like"
	RelationTypePhotoLike     RelationType = "photo.like"
	RelationTypePostComment   RelationType = "post.comment"
	RelationTypePhotoComment  RelationType = "photo.comment"
)

func New(connStr string) (*DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) WriteRelations(socialNetworkType string, owner Owner, relationType RelationType, details map[string]interface{}, ids []int64) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO public."Relations" 
		("Timestamp", "SocialNetworkType", "OwnerType", "OwnerID", "RelationType", "Details", "IDs")
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ("SocialNetworkType", "OwnerType", "OwnerID", "RelationType", "Details")
		DO UPDATE SET "Timestamp" = EXCLUDED."Timestamp", "IDs" = EXCLUDED."IDs"
	`

	_, err = db.conn.Exec(query, time.Now(), socialNetworkType, owner.Type, owner.ID, relationType, detailsJSON, ids)
	return err
}

func (db *DB) WriteObject(socialNetworkType string, owner Owner, objectType string, details map[string]interface{}, data map[string]interface{}) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO public."Objects_` + objectType + `" 
		("Timestamp", "SocialNetworkType", "OwnerType", "OwnerID", "Details", "Data")
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT ("SocialNetworkType", "OwnerType", "OwnerID", "Details")
		DO UPDATE SET "Timestamp" = EXCLUDED."Timestamp", "Data" = EXCLUDED."Data", "IsChanged" = true
	`

	_, err = db.conn.Exec(query, time.Now(), socialNetworkType, owner.Type, owner.ID, detailsJSON, dataJSON)
	return err
}
