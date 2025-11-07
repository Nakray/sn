package database

import (
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
	IsUnlockable      *bool
	UnlockIDs         []int64
}

func (db *DB) GetDueMonitoringTasks() ([]MonitoringTask, error) {
	query := `
		SELECT "ID", "SocialNetworkType", "OwnerType", "OwnerID", "Period", 
		       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID",
			              "IsUnlocked", "UnlockIDs"
		FROM monitoring."Tasks"
       WHERE (("IsUnlocked" IS NULL AND ("LastTimestamp" + ("Period" * INTERVAL '1 minute') <= NOW()) OR "IsUnlocked" = true)) ORDER BY "LastTimestamp" ASC
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
		var unlockIDsJSON []byte

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
			&task.IsUnlockable,
			&unlockIDsJSON,
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
		if len(unlockIDsJSON) > 0 {
			json.Unmarshal(unlockIDsJSON, &task.UnlockIDs)
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (db *DB) UpdateTaskLastTimestamp(task *MonitoringTask, success bool) error {
	query := `UPDATE monitoring."Tasks" SET "LastTimestamp" = NOW() WHERE "ID" = $1`
	if success && task.IsUnlockable != nil && *task.IsUnlockable {
		query += `, "IsUnlocked" = false`
	}
	query += ` WHERE "ID" = $1`

	_, err := db.conn.Exec(query, task.ID)
	if err != nil {
		return err
	}

	// Unlock dependent tasks if needed
	if success && len(task.UnlockIDs) > 0 {
		unlockQuery := `UPDATE monitoring."Tasks" SET "IsUnlocked" = true WHERE "ID" = ANY($1)`
		_, err = db.conn.Exec(unlockQuery, task.UnlockIDs)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) CreateMonitoringTask(task *MonitoringTask) error {
	filtersJSON, _ := json.Marshal(task.Filters)
	filterLimitsJSON, _ := json.Marshal(task.FilterLimits)
	unlockIDsJSON, _ := json.Marshal(task.UnlockIDs)

	query := `
				INSERT INTO monitoring."Tasks"
						("SocialNetworkType", "OwnerType", "OwnerID", "Period", "Filters", "FilterLimits", "AccountGroupID", "IsUnlocked", "UnlockIDs")
								VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
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
		task.IsUnlockable,
		unlockIDsJSON,
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
							       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID",
								   		       "IsUnlocked", "UnlockIDs"
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
		var unlockIDsJSON []byte

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
			&task.IsUnlockable,
			&unlockIDsJSON,
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
		if len(unlockIDsJSON) > 0 {
			json.Unmarshal(unlockIDsJSON, &task.UnlockIDs)
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
