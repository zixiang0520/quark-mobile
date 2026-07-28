package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"quark-mobile/internal/model"
	"quark-mobile/internal/service"
)

type Manager struct {
	mu      sync.RWMutex
	tasks   map[string]*model.TransferTask
	queue   chan *model.TransferTask
	service *service.TransferService
	maxConc int
	wg      sync.WaitGroup
}

func NewManager(svc *service.TransferService, maxConcurrent int) *Manager {
	m := &Manager{
		tasks:   make(map[string]*model.TransferTask),
		queue:   make(chan *model.TransferTask, maxConcurrent*2),
		service: svc,
		maxConc: maxConcurrent,
	}
	m.start()
	return m
}

func (m *Manager) start() {
	for i := 0; i < m.maxConc; i++ {
		m.wg.Add(1)
		go m.worker(i)
	}
}

func (m *Manager) worker(id int) {
	defer m.wg.Done()
	for task := range m.queue {
		m.runTask(task)
	}
}

func (m *Manager) runTask(task *model.TransferTask) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Minute)
	defer cancel()

	task.Status = "running"
	task.UpdatedAt = time.Now()

	if err := m.service.ExecuteTransfer(ctx, task); err != nil {
		task.Status = "failed"
		task.Error = err.Error()
	}

	task.UpdatedAt = time.Now()
}

func (m *Manager) CreateTask(req model.TransferRequest) *model.TransferTask {
	m.mu.Lock()
	defer m.mu.Unlock()

	task := m.service.NewTask(req)
	m.tasks[task.ID] = task
	m.queue <- task
	return task
}

func (m *Manager) GetTask(id string) (*model.TransferTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	return task, nil
}

func (m *Manager) ListTasks() []*model.TransferTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*model.TransferTask, 0, len(m.tasks))
	for _, t := range m.tasks {
		result = append(result, t)
	}
	return result
}

func (m *Manager) CancelTask(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[id]
	if !ok {
		return fmt.Errorf("task not found: %s", id)
	}
	if task.Status == "running" {
		task.Status = "cancelled"
		task.Error = "cancelled by user"
		task.UpdatedAt = time.Now()
	}
	return nil
}

func (m *Manager) Stop() {
	close(m.queue)
	m.wg.Wait()
}
