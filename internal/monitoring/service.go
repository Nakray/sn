package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"

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
}

func NewService(db *database.DB, cfg *config.Config) *Service {
	return &Service{
		db:     db,
		config: cfg,
		stopCh: make(chan struct{}),
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
	s.processTasks(workerID)

	for {
		select {
		case <-s.stopCh:
			log.Printf("Monitoring worker %d stopped\n", workerID)
			return
		case <-ticker.C:
			s.processTasks(workerID)
		}
	}
}

func (s *Service) processTasks(workerID int) {
	tasks, err := s.db.GetDueMonitoringTasks()
	if err != nil {
		log.Printf("Worker %d: failed to get due tasks: %v\n", workerID, err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Printf("Worker %d: processing %d tasks\n", workerID, len(tasks))

	for _, task := range tasks {
		if err := s.processTask(task); err != nil {
			log.Printf("Worker %d: task %d failed: %v\n", workerID, task.ID, err)
		} else {
			log.Printf("Worker %d: task %d completed successfully\n", workerID, task.ID)
		}

		// Update task timestamp
		if err := s.db.UpdateTaskLastTimestamp(task.ID); err != nil {
			log.Printf("Worker %d: failed to update task %d timestamp: %v\n", workerID, task.ID, err)
		}
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
