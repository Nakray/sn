package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"
	"context"

	"github.com/redis/go-redis/v9"

	"github.com/Nakray/sn/internal/config"
	"github.com/Nakray/sn/internal/database"
	"github.com/Nakray/sn/internal/vk"
)

type Service struct {
	db      *database.DB
	config  *config.Config
	stopCh  chan struct{}
	wg      sync.WaitGroup
	running bool
	mu      sync.Mutex
	redis  *redis.Client
}

rdb *redis.Client, ervice {
	return &Service{
		db:     db,
		config: cfg,
		stopCh: make(chan struct{}),
		redis:  rdb,
	}
}

func (s *Service) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	// Start worker pool
	for i := 0; i < s.config.Monitoring.Workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopCh)
	s.wg.Wait()
}

func (s *Service) worker(workerID int) {
	defer s.wg.Done()

	ticker := time.NewTicker(time.Duration(s.config.Monitoring.IntervalMinutes) * time.Minute)
	defer ticker.Stop()

	log.Printf("Monitoring worker %d started\n", workerID)

	// Run immediately on start
	s.processNextTask(workerID)
	for {
		select {
		case <-s.stopCh:
			log.Printf("Monitoring worker %d stopped\n", workerID)
			return
		case <-ticker.C:
			s.processNextTask(workerID)		}
	}
}

func (s *Service) processNextTask(workerID int) {
	// Get next monitoring task with advisory lock
	task, err := s.db.GetNextMonitoringTask()
	if err != nil {
		log.Printf("Worker %d: failed to get next task: %v", workerID, err)
		return
	}

	// No tasks available
	if task == nil {
		return
	}

	// Release advisory lock when done
	defer s.db.ReleaseTaskLock(task.ID)

	// Check if task is enabled in Redis
	if !s.isTaskEnabled(task.ID) {
		log.Printf("Worker %d: task %d is disabled, skipping", workerID, task.ID)
		return
	}

	// Check if task is in cooldown period
	if s.isTaskInCooldown(task.ID) {
		log.Printf("Worker %d: task %d is in cooldown, skipping", workerID, task.ID)
		return
	}

	// Process the task
	success := true
	if err := s.processTask(*task); err != nil {
		success = false
		log.Printf("Worker %d: task %d failed: %v", workerID, task.ID, err)
	}

	// Update task timestamp
	if err := s.db.UpdateTaskLastTimestamp(task, success); err != nil {
		log.Printf("Worker %d: failed to update task %d timestamp: %v", workerID, task.ID, err)
	}
}
func (s *Service) processTask(task database.MonitoringTask) error {
	// Get available account
	account, err := s.db.GetAvailableAccount(task.SocialNetworkType, task.AccountGroupID)
	if err != nil {
		return err
	}

	// Create VK client
	accessToken := "" // Extract from account.Session
	if account.Session != nil {
		if token, ok := account.Session["access_token"].(string); ok {
			accessToken = token
		}
	}

	if accessToken == "" {
		return fmt.Errorf("no access token for account %d", account.ID)
	}

	client, err := vk.NewClient(accessToken, account.Proxy)
	if err != nil {
		return err
	}

	collector := vk.NewCollector(client, s.db)

	// Collect entity
	return collector.CollectEntity(task.OwnerType, task.OwnerID)
}

// isTaskEnabled проверяет разрешен ли запуск задачи через Redis
func (s *Service) isTaskEnabled(taskID int64) bool {
	key := fmt.Sprintf("monitoring:task:%d:enabled", taskID)
	val, err := s.redis.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return true // По умолчанию включено
	}
	return val == "true"
}

// isTaskInCooldown проверяет находится ли задача в cooldown
func (s *Service) isTaskInCooldown(taskID int64) bool {
	key := fmt.Sprintf("monitoring:task:%d:cooldown", taskID)
	exists, _ := s.redis.Exists(context.Background(), key).Result()
	return exists > 0
}

// DisableTask выключает задачу
func (s *Service) DisableTask(taskID int64) error {
	key := fmt.Sprintf("monitoring:task:%d:enabled", taskID)
	return s.redis.Set(context.Background(), key, "false", 0).Err()
}

// EnableTask включает задачу
func (s *Service) EnableTask(taskID int64) error {
	key := fmt.Sprintf("monitoring:task:%d:enabled", taskID)
	return s.redis.Set(context.Background(), key, "true", 0).Err()
}

// SetTaskCooldown устанавливает cooldown для задачи
func (s *Service) SetTaskCooldown(taskID int64, duration time.Duration) error {
	key := fmt.Sprintf("monitoring:task:%d:cooldown", taskID)
	return s.redis.Set(context.Background(), key, "1", duration).Err()
}

// TriggerTaskNow запускает задачу немедленно
func (s *Service) TriggerTaskNow(taskID int64) error {
	// Убираем cooldown
	key := fmt.Sprintf("monitoring:task:%d:cooldown", taskID)
	s.redis.Del(context.Background(), key)
	
	// Сбрасываем LastTimestamp
	return s.db.TriggerTaskNow(taskID)
}
