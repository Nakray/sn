package database

import (
	"database/sql"
	"encoding/json"
	"time"
)

type MonitoringTask struct {
	ID                int64
	SocialNetworkType string
	OwnerType         OwnerType
	OwnerID           int64
	Period            int
	LastTimestamp     time.Time
	Filters           map[string]interface{}
	FilterLimits      map[string]interface{}
	AccountGroupID    int
}

func (db *DB) GetDueMonitoringTasks() ([]MonitoringTask, error) {
	query := `
		SELECT "ID", "SocialNetworkType", "OwnerType", "OwnerID", "Period", 
		       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID"
		FROM monitoring."Tasks"
		WHERE "LastTimestamp" + ("Period" * INTERVAL '1 minute') <= NOW()
		ORDER BY "LastTimestamp" ASC
		LIMIT 100
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []MonitoringTask
	for rows.Next() {
		var task MonitoringTask
		var filtersJSON, filterLimitsJSON []byte

		err := rows.Scan(
			&task.ID,
			&task.SocialNetworkType,
			&task.OwnerType,
			&task.OwnerID,
			&task.Period,
			&task.LastTimestamp,
			&filtersJSON,
			&filterLimitsJSON,
			&task.AccountGroupID,
		)
		if err != nil {
			return nil, err
		}

		if len(filtersJSON) > 0 {
			json.Unmarshal(filtersJSON, &task.Filters)
		}
		if len(filterLimitsJSON) > 0 {
			json.Unmarshal(filterLimitsJSON, &task.FilterLimits)
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (db *DB) UpdateTaskLastTimestamp(taskID int64) error {
	query := `UPDATE monitoring."Tasks" SET "LastTimestamp" = NOW() WHERE "ID" = $1`
	_, err := db.conn.Exec(query, taskID)
	return err
}

func (db *DB) CreateMonitoringTask(task *MonitoringTask) error {
	filtersJSON, _ := json.Marshal(task.Filters)
	filterLimitsJSON, _ := json.Marshal(task.FilterLimits)

	query := `
		INSERT INTO monitoring."Tasks" 
		("SocialNetworkType", "OwnerType", "OwnerID", "Period", "Filters", "FilterLimits", "AccountGroupID")
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING "ID"
	`

	return db.conn.QueryRow(query,
		task.SocialNetworkType,
		task.OwnerType,
		task.OwnerID,
		task.Period,
		filtersJSON,
		filterLimitsJSON,
		task.AccountGroupID,
	).Scan(&task.ID)
}

func (db *DB) DeleteMonitoringTask(taskID int64) error {
	query := `DELETE FROM monitoring."Tasks" WHERE "ID" = $1`
	_, err := db.conn.Exec(query, taskID)
	return err
}

func (db *DB) ListMonitoringTasks() ([]MonitoringTask, error) {
	query := `
		SELECT "ID", "SocialNetworkType", "OwnerType", "OwnerID", "Period", 
		       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID"
		FROM monitoring."Tasks"
		ORDER BY "ID" DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []MonitoringTask
	for rows.Next() {
		var task MonitoringTask
		var filtersJSON, filterLimitsJSON []byte

		err := rows.Scan(
			&task.ID,
			&task.SocialNetworkType,
			&task.OwnerType,
			&task.OwnerID,
			&task.Period,
			&task.LastTimestamp,
			&filtersJSON,
			&filterLimitsJSON,
			&task.AccountGroupID,
		)
		if err != nil {
			return nil, err
		}

		if len(filtersJSON) > 0 {
			json.Unmarshal(filtersJSON, &task.Filters)
		}
		if len(filterLimitsJSON) > 0 {
			json.Unmarshal(filterLimitsJSON, &task.FilterLimits)
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
