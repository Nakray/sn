package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Account struct {
	ID                int64
	SocialNetworkType string
	Login             string
	Password          string
	Session           map[string]interface{}
	Proxy             string
	IsBlocked         bool
	Info              string
	UnavailableUntil  *time.Time
	GroupID           int
}

func (db *DB) GetAvailableAccount(socialNetworkType string, groupID int) (*Account, error) {
	query := `
		SELECT "ID", "SocialNetworkType", "Login", "Password", "Session", 
		       "Proxy", "IsBlocked", "Info", "UnavailableUntil", "GroupID"
		FROM public."Accounts"
		WHERE "SocialNetworkType" = $1 
		  AND "GroupID" = $2
		  AND "IsBlocked" = false
		  AND ("UnavailableUntil" IS NULL OR "UnavailableUntil" < NOW())
		ORDER BY RANDOM()
		LIMIT 1
	`

	var acc Account
	var sessionJSON []byte
	var unavailableUntil sql.NullTime

	err := db.conn.QueryRow(query, socialNetworkType, groupID).Scan(
		&acc.ID,
		&acc.SocialNetworkType,
		&acc.Login,
		&acc.Password,
		&sessionJSON,
		&acc.Proxy,
		&acc.IsBlocked,
		&acc.Info,
		&unavailableUntil,
		&acc.GroupID,
	)

	if err != nil {
		return nil, err
	}

	if len(sessionJSON) > 0 {
		json.Unmarshal(sessionJSON, &acc.Session)
	}

	if unavailableUntil.Valid {
		acc.UnavailableUntil = &unavailableUntil.Time
	}

	return &acc, nil
}

func (db *DB) UpdateAccountSession(accountID int64, session map[string]interface{}) error {
	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return err
	}

	query := `UPDATE public."Accounts" SET "Session" = $1, "IsChanged" = true WHERE "ID" = $2`
	_, err = db.conn.Exec(query, sessionJSON, accountID)
	return err
}

func (db *DB) MarkAccountBlocked(accountID int64, info string) error {
	query := `UPDATE public."Accounts" SET "IsBlocked" = true, "Info" = $1, "IsChanged" = true WHERE "ID" = $2`
	_, err := db.conn.Exec(query, info, accountID)
	return err
}

func (db *DB) SetAccountUnavailable(accountID int64, duration time.Duration) error {
	until := time.Now().Add(duration)
	query := `UPDATE public."Accounts" SET "UnavailableUntil" = $1, "IsChanged" = true WHERE "ID" = $2`
	_, err := db.conn.Exec(query, until, accountID)
	return err
}

func (db *DB) ListAccounts() ([]Account, error) {
	query := `
		SELECT "ID", "SocialNetworkType", "Login", "Password", "Session", 
		       "Proxy", "IsBlocked", "Info", "UnavailableUntil", "GroupID"
		FROM public."Accounts"
		ORDER BY "ID" DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var acc Account
		var sessionJSON []byte
		var unavailableUntil sql.NullTime

		err := rows.Scan(
			&acc.ID,
			&acc.SocialNetworkType,
			&acc.Login,
			&acc.Password,
			&sessionJSON,
			&acc.Proxy,
			&acc.IsBlocked,
			&acc.Info,
			&unavailableUntil,
			&acc.GroupID,
		)
		if err != nil {
			return nil, err
		}

		if len(sessionJSON) > 0 {
			json.Unmarshal(sessionJSON, &acc.Session)
		}

		if unavailableUntil.Valid {
			acc.UnavailableUntil = &unavailableUntil.Time
		}

		accounts = append(accounts, acc)
	}

	return accounts, rows.Err()
}

func (db *DB) CreateAccount(acc *Account) error {
	sessionJSON, _ := json.Marshal(acc.Session)

	query := `
		INSERT INTO public."Accounts" 
		("SocialNetworkType", "Login", "Password", "Session", "Proxy", "IsBlocked", "Info", "GroupID")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING "ID"
	`

	return db.conn.QueryRow(query,
		acc.SocialNetworkType,
		acc.Login,
		acc.Password,
		sessionJSON,
		acc.Proxy,
		acc.IsBlocked,
		acc.Info,
		acc.GroupID,
	).Scan(&acc.ID)
}

func (db *DB) DeleteAccount(accountID int64) error {
	query := `DELETE FROM public."Accounts" WHERE "ID" = $1`
	_, err := db.conn.Exec(query, accountID)
	return err
}
