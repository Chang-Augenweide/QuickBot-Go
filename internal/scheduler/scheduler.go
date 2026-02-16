package scheduler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/robfig/cron/v3"
	_ "github.com/mattn/go-sqlite3"
)

// Task represents a scheduled task
type Task struct {
	ID        string
	Name      string
	SessionID string
	Status    string
	Payload   map[string]interface{}
	NextRun   time.Time
	CreatedAt time.Time
}

// TaskPayload represents task payload structure
type TaskPayload struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
	RemindAt string `json:"remind_at,omitempty"`
}

// Scheduler represents task scheduler
type Scheduler struct {
	conn     *sql.DB
	cron     *cron.Cron
	handlers map[string]func(*Task)
}

// NewScheduler creates a new scheduler instance
func NewScheduler(dbPath string) (*Scheduler, error) {
	// Create database file if doesn't exist
	_, err := os.Create(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create database file: %w", err)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	scheduler := &Scheduler{
		conn:     conn,
		cron:     cron.New(cron.WithSeconds()),
		handlers: make(map[string]func(*Task)),
	}

	err = scheduler.initDB()
	if err != nil {
		return nil, err
	}

	// Load and schedule existing tasks
	err = scheduler.loadTasks()
	if err != nil {
		return nil, err
	}

	return scheduler, nil
}

// initDB initializes database schema
func (s *Scheduler) initDB() error {
	_, err := s.conn.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			session_id TEXT NOT NULL,
			status TEXT NOT NULL,
			payload TEXT,
			next_run DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %w", err)
	}

	// Create index
	_, err = s.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_tasks_next_run
		ON tasks(next_run)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("✓ Scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cron.Stop()
	if s.conn != nil {
		s.conn.Close()
	}
	log.Println("✓ Scheduler stopped")
}

// AddTask adds a new task
func (s *Scheduler) AddTask(name, sessionID string, payload map[string]interface{}, nextRun time.Time) (string, error) {
	id := fmt.Sprintf("%d", time.Now().UnixNano())

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = s.conn.Exec(`
		INSERT INTO tasks (id, name, session_id, status, payload, next_run)
		VALUES (?, ?, ?, ?, ?, ?)
	`, id, name, sessionID, "scheduled", string(payloadJSON), nextRun)

	if err != nil {
		return "", fmt.Errorf("failed to insert task: %w", err)
	}

	// Add to cron
	cronExpr := formatCronExpression(nextRun)
	entryID, err := s.cron.AddFunc(cronExpr, func() {
		s.executeTask(id)
	})
	if err != nil {
		return "", fmt.Errorf("failed to schedule task: %w", err)
	}

	log.Printf("Task added: %s ( Cron entry: %d )", name, entryID)
	return id, nil
}

// AddReminder adds a reminder task
func (s *Scheduler) AddReminder(sessionID, message, remindAt string) (string, error) {
	// Parse time (HH:MM format)
	var parsedTime time.Time
	var err error

	if len(remindAt) == 5 && remindAt[2] == ':' {
		// Format: HH:MM
		now := time.Now()
		hour := time.Hour * time.Duration(now.Hour())
		minute := time.Minute * time.Duration(now.Minute())

		// Parse the target time
		parts := make([]time.Duration, 2)
		fmt.Sscanf(remindAt, "%d:%d", &parts[0], &parts[1])
		targetTime := parts[0] + parts[1]

		// Set target time for today
		parsedTime = time.Date(now.Year(), now.Month(), now.Day(), int(targetTime/time.Hour), int((targetTime%time.Hour)/time.Minute), 0, 0, time.Local)

		// If target time has passed, schedule for tomorrow
		if parsedTime.Before(now) {
			parsedTime = parsedTime.Add(24 * time.Hour)
		}
	} else {
		// Try parsing as full datetime
		parsedTime, err = time.Parse(time.RFC3339, remindAt)
		if err != nil {
			return "", fmt.Errorf("invalid time format: %w", err)
		}
	}

	return s.AddTask("reminder", sessionID, TaskPayload{
		Type:    "reminder",
		Message: message,
	}, parsedTime)
}

// GetTask retrieves a task by ID
func (s *Scheduler) GetTask(id string) (*Task, error) {
	var task Task
	var payload string

	err := s.conn.QueryRow(`
		SELECT id, name, session_id, status, payload, next_run, created_at
		FROM tasks WHERE id = ?
	`, id).Scan(&task.ID, &task.Name, &task.SessionID, &task.Status,
		&payload, &task.NextRun, &task.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	err = json.Unmarshal([]byte(payload), &task.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &task, nil
}

// GetAllTasks returns all tasks
func (s *Scheduler) GetAllTasks() ([]Task, error) {
	rows, err := s.conn.Query(`
		SELECT id, name, session_id, status, payload, next_run, created_at
		FROM tasks
		ORDER BY next_run ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var payload string

		err := rows.Scan(&task.ID, &task.Name, &task.SessionID, &task.Status,
			&payload, &task.NextRun, &task.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		err = json.Unmarshal([]byte(payload), &task.Payload)
		if err != nil {
			continue
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetDueTasks returns tasks that are due for execution
func (s *Scheduler) GetDueTasks() []Task {
	tasks, err := s.GetAllTasks()
	if err != nil {
		return nil
	}

	var dueTasks []Task
	now := time.Now()
	for _, task := range tasks {
		if task.Status == "scheduled" && task.NextRun.Before(now) {
			dueTasks = append(dueTasks, task)
		}
	}

	return dueTasks
}

// DeleteTask deletes a task
func (s *Scheduler) DeleteTask(id string) error {
	_, err := s.conn.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// SetTaskHandler sets a handler for task execution
func (s *Scheduler) SetTaskHandler(handler func(*Task)) {
	s.handlers["default"] = handler
}

// executeTask executes a task
func (s *Scheduler) executeTask(id string) {
	task, err := s.GetTask(id)
	if err != nil {
		log.Printf("Failed to get task %s: %v", id, err)
		return
	}

	if task == nil {
		return
	}

	log.Printf("Executing task: %s", task.Name)

	// Call handler
	if handler, ok := s.handlers["default"]; ok {
		handler(task)
	}

	// Delete the task after execution
	err = s.DeleteTask(id)
	if err != nil {
		log.Printf("Failed to delete task %s: %v", id, err)
	}
}

// loadTasks loads existing tasks from database and schedules them
func (s *Scheduler) loadTasks() error {
	tasks, err := s.GetAllTasks()
	if err != nil {
		return err
	}

	for _, task := range tasks {
		if task.Status == "scheduled" && task.NextRun.After(time.Now()) {
			cronExpr := formatCronExpression(task.NextRun)
			_, err := s.cron.AddFunc(cronExpr, func() {
				s.executeTask(task.ID)
			})
			if err != nil {
				log.Printf("Failed to schedule task %s: %v", task.ID, err)
			}
		}
	}

	return nil
}

// formatCronExpression formats a time as a cron expression
func formatCronExpression(t time.Time) string {
	return fmt.Sprintf("%d %d %d %d %d ?",
		t.Second(),
		t.Minute(),
		t.Hour(),
		t.Day(),
		t.Month(),
	)
}

// TestScheduler runs tests on the scheduler module
func TestScheduler() error {
	log.Println("Testing Scheduler module...")

	scheduler, err := NewScheduler("test_scheduler.db")
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}
	defer scheduler.Stop()

	// Add a test task
	testTime := time.Now().Add(2 * time.Minute)
	taskID, err := scheduler.AddTask("test", "session1", TaskPayload{
		Type:    "test",
		Message: "test message",
	}, testTime)
	if err != nil {
		return fmt.Errorf("failed to add task: %w", err)
	}
	log.Printf("✓ Test task added: %s", taskID)

	// Get task
	task, err := scheduler.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}
	log.Println("✓ Task retrieved")

	// Get all tasks
	tasks, err := scheduler.GetAllTasks()
	if err != nil {
		return fmt.Errorf("failed to get all tasks: %w", err)
	}
	log.Printf("✓ Retrieved %d tasks", len(tasks))

	// Cleanup
	os.Remove("test_scheduler.db")
	log.Println("✓ Scheduler module tests passed")
	return nil
}

func main() {
	log.Println("QuickBot Go Scheduler Module")
	TestScheduler()
}
