package database

import (
	"encoding/json"
	"time"
	"database/sql"
	"strings"
	
	"github.com/lib/pq"
)

type MonitoringTask struct {
	ID                 int64
	SocialNetworkType  string
	OwnerType          OwnerType
	OwnerID            int64
	Period             int
	LastTimestamp      time.Time
	Filters            map[string]interface{}
	FilterLimits       map[string]interface{}
	AccountGroupID     int
	IsUnlockable       bool
	UnlockIDs          []int64
}

func (db *DB) GetDueMonitoringTasks() ([]MonitoringTask, error) {
	query := `
		SELECT now(), "ID", "SocialNetworkType", "OwnerType", "OwnerID", "Period",
		       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID",
		       "IsUnlocked" IS NOT NULL, "UnlockIDs"
		FROM monitoring."Tasks"
		WHERE (("IsUnlocked" IS NULL AND ("LastTimestamp" + ("Period" * INTERVAL '1 minute')) <= now()) OR "IsUnlocked" = true)
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
		var now time.Time

		err := rows.Scan(
			&now,
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
	var queryParts []string
	var args []interface{}
	
	// Build UPDATE query for LastTimestamp
	queryParts = append(queryParts, `UPDATE monitoring."Tasks" SET "LastTimestamp" = NOW()`)
	
	if success && task.IsUnlockable {
		queryParts = append(queryParts, `, "IsUnlocked" = false`)
	}
	
	queryParts = append(queryParts, ` WHERE "ID" = $1;`)
	args = append(args, task.ID)
	
	// Add unlock query if needed
	if success && len(task.UnlockIDs) > 0 {
		queryParts = append(queryParts, `UPDATE monitoring."Tasks" SET "IsUnlocked" = true WHERE "ID" = ANY($2);`)
		// Convert []int64 to pq.Array for PostgreSQL
		args = append(args, pq.Array(task.UnlockIDs))
	}
	
	// Execute the combined query
	finalQuery := strings.Join(queryParts, "")
	_, err := db.conn.Exec(finalQuery, args...)
	return err
}	return nil

// GetNextMonitoringTask получает следующую доступную задачу с использованием PostgreSQL advisory lock
func (db *DB) GetNextMonitoringTask() (*MonitoringTask, error) {
	query := `
		SELECT now(), "ID", "SocialNetworkType", "OwnerType", "OwnerID", "Period",
		       "LastTimestamp", "Filters", "FilterLimits", "AccountGroupID",
		       "IsUnlocked" IS NOT NULL, "UnlockIDs"
		FROM monitoring."Tasks"
		WHERE (
			  ("IsUnlocked" IS NULL AND ("LastTimestamp" + ("Period" * INTERVAL '1 minute')) <= now()) 
			  OR "IsUnlocked" = true
		  )
		  AND pg_try_advisory_lock(hashtext('monitoring_tasks'), "ID")
		ORDER BY "LastTimestamp" ASC
		LIMIT 1
	`

	var task MonitoringTask
	var now time.Time
	var filtersJSON, filterLimitsJSON, unlockIDsJSON []byte

	err := db.conn.QueryRow(query).Scan(
		&now,
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

	if err == sql.ErrNoRows {
		return nil, nil // Нет доступных задач
	}
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(filtersJSON) > 0 {
		json.Unmarshal(filtersJSON, &task.Filters)
	}
	if len(filterLimitsJSON) > 0 {
		json.Unmarshal(filterLimitsJSON, &task.FilterLimits)
	}
	if len(unlockIDsJSON) > 0 {
		json.Unmarshal(unlockIDsJSON, &task.UnlockIDs)
	}

	return &task, nil
}

// ReleaseTaskLock освобождает PostgreSQL advisory lock для задачи
func (db *DB) ReleaseTaskLock(taskID int64) error {
	_, err := db.conn.Exec(
		`SELECT pg_advisory_unlock(hashtext('monitoring_tasks'), $1)`,
		taskID,
	)
	return err
}

// TriggerTaskNow запускает задачу немедленно, сбрасывая LastTimestamp
func (db *DB) TriggerTaskNow(taskID int64) error {
	_, err := db.conn.Exec(
		`UPDATE monitoring."Tasks" 
		 SET "LastTimestamp" = now() - INTERVAL '1 year' 
		 WHERE "ID" = $1`,
		taskID,
	)
	return err
}
}
