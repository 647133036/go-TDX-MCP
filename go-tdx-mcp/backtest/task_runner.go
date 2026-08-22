package backtest

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

func shortUUID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

type TaskStatus string

const (
	TaskPending TaskStatus = "pending"
	TaskRunning TaskStatus = "running"
	TaskDone    TaskStatus = "done"
	TaskFailed  TaskStatus = "failed"
)

type BacktestTask struct {
	ID        string
	Strategy  string
	Params    map[string]float64
	Cash      float64
	Market    int
	Code      string
	Period    string
	Count     int
	Job       func() *Result
	Status    TaskStatus
	Result    *Result
	Error     string
	CreatedAt time.Time
	DoneAt    time.Time
}

type TaskResult struct {
	ID        string     `json:"id"`
	Status    TaskStatus `json:"status"`
	Result    *Result    `json:"result,omitempty"`
	Error     string     `json:"error,omitempty"`
	CreatedAt string     `json:"created_at"`
	DoneAt    string     `json:"done_at,omitempty"`
}

const maxCachedTasks = 100

type TaskRunner struct {
	mu        sync.RWMutex
	tasks     map[string]*BacktestTask
	order     []string
	maxCached int
}

func NewTaskRunner(workers int) *TaskRunner {
	tr := &TaskRunner{
		tasks:     make(map[string]*BacktestTask),
		order:     make([]string, 0),
		maxCached: maxCachedTasks,
	}
	if workers <= 0 {
		workers = 4
	}
	for i := 0; i < workers; i++ {
		go tr.worker()
	}
	return tr
}

func (tr *TaskRunner) worker() {
	for {
		tr.mu.Lock()
		var task *BacktestTask
		for _, id := range tr.order {
			if t, ok := tr.tasks[id]; ok && t.Status == TaskPending {
				t.Status = TaskRunning
				task = t
				break
			}
		}
		if task == nil {
			tr.mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			continue
		}
		tr.mu.Unlock()

		result := task.Job()
		tr.mu.Lock()
		if result != nil {
			task.Result = result
			task.Status = TaskDone
		} else {
			if task.Error == "" {
				task.Error = "回测执行失败: 策略无效或数据不足"
			}
			task.Status = TaskFailed
		}
		task.DoneAt = time.Now()
		tr.mu.Unlock()
	}
}

func (tr *TaskRunner) Submit(job func() *Result) *BacktestTask {
	task := &BacktestTask{
		ID:        shortUUID(),
		Job:       job,
		Status:    TaskPending,
		CreatedAt: time.Now(),
	}
	tr.mu.Lock()
	tr.tasks[task.ID] = task
	tr.order = append(tr.order, task.ID)
	for len(tr.order) > tr.maxCached {
		oldest := tr.order[0]
		tr.order = tr.order[1:]
		delete(tr.tasks, oldest)
	}
	tr.mu.Unlock()
	return task
}

func (tr *TaskRunner) SubmitJob(task *BacktestTask) string {
	if task.Status == "" {
		task.Status = TaskPending
	}
	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now()
	}
	tr.mu.Lock()
	tr.tasks[task.ID] = task
	tr.order = append(tr.order, task.ID)
	for len(tr.order) > tr.maxCached {
		oldest := tr.order[0]
		tr.order = tr.order[1:]
		delete(tr.tasks, oldest)
	}
	tr.mu.Unlock()
	return task.ID
}

func (tr *TaskRunner) Peek(id string) *TaskResult {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	task, ok := tr.tasks[id]
	if !ok {
		return nil
	}
	result := &TaskResult{
		ID:        task.ID,
		Status:    task.Status,
		Result:    task.Result,
		Error:     task.Error,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
	}
	if !task.DoneAt.IsZero() {
		result.DoneAt = task.DoneAt.Format(time.RFC3339)
	}
	return result
}

func (tr *TaskRunner) ListRecent(limit int) []*TaskResult {
	if limit <= 0 {
		limit = 20
	}
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	n := len(tr.order)
	if n > limit {
		n = limit
	}
	results := make([]*TaskResult, 0, n)
	for i := n; i > 0; i-- {
		id := tr.order[len(tr.order)-i]
		if t, ok := tr.tasks[id]; ok {
			item := &TaskResult{
				ID:        t.ID,
				Status:    t.Status,
				Error:     t.Error,
				CreatedAt: t.CreatedAt.Format(time.RFC3339),
			}
			if !t.DoneAt.IsZero() {
				item.DoneAt = t.DoneAt.Format(time.RFC3339)
			}
			results = append(results, item)
		}
	}
	return results
}
