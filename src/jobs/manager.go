package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yhonda-ohishi/browser_render_go/src/browser"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	ID           string                   `json:"id"`
	Status       JobStatus                `json:"status"`
	CreatedAt    time.Time                `json:"created_at"`
	CompletedAt  *time.Time               `json:"completed_at,omitempty"`
	Error        string                   `json:"error,omitempty"`
	VehicleCount int                      `json:"vehicle_count,omitempty"`
	HonoResponse *browser.HonoAPIResponse `json:"hono_response,omitempty"`
}

type Manager struct {
	jobs     map[string]*Job
	mu       sync.RWMutex
	renderer *browser.Renderer
}

func NewManager(renderer *browser.Renderer) *Manager {
	return &Manager{
		jobs:     make(map[string]*Job),
		renderer: renderer,
	}
}

func (m *Manager) CreateJob() string {
	jobID := uuid.New().String()

	m.mu.Lock()
	m.jobs[jobID] = &Job{
		ID:        jobID,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
	}
	m.mu.Unlock()

	// Start processing in background
	go m.processJob(jobID)

	return jobID
}

func (m *Manager) processJob(jobID string) {
	// Update status to running
	m.updateJobStatus(jobID, JobStatusRunning)

	// Create independent context for background processing
	ctx := context.Background()

	// Call the renderer with fixed parameters
	vehicleData, _, honoAPIResponse, err := m.renderer.GetVehicleData(
		ctx,
		"",         // Session ID
		"00000000", // Branch ID
		"0",        // Filter ID
		false,      // Force login
	)

	// Update job with results
	m.mu.Lock()
	job := m.jobs[jobID]
	if job != nil {
		now := time.Now()
		job.CompletedAt = &now

		if err != nil {
			job.Status = JobStatusFailed
			job.Error = err.Error()
			log.Printf("Job %s failed: %v", jobID, err)
		} else {
			job.Status = JobStatusCompleted
			job.VehicleCount = len(vehicleData)
			job.HonoResponse = honoAPIResponse
			log.Printf("Job %s completed successfully with %d vehicles", jobID, len(vehicleData))

			if honoAPIResponse != nil {
				log.Printf("Hono API Response for job %s - Success: %v, Records: %d/%d",
					jobID,
					honoAPIResponse.Success,
					honoAPIResponse.RecordsAdded,
					honoAPIResponse.TotalRecords)
			}
		}
	}
	m.mu.Unlock()

	// Clean up old jobs after 10 minutes
	go func() {
		time.Sleep(10 * time.Minute)
		m.mu.Lock()
		delete(m.jobs, jobID)
		m.mu.Unlock()
		log.Printf("Job %s cleaned up", jobID)
	}()
}

func (m *Manager) updateJobStatus(jobID string, status JobStatus) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if job, exists := m.jobs[jobID]; exists {
		job.Status = status
		log.Printf("Job %s status updated to %s", jobID, status)
	}
}

func (m *Manager) GetJob(jobID string) (*Job, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	job, exists := m.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	// Return a copy to avoid race conditions
	jobCopy := *job
	return &jobCopy, nil
}

func (m *Manager) GetAllJobs() []*Job {
	m.mu.RLock()
	defer m.mu.RUnlock()

	jobs := make([]*Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobCopy := *job
		jobs = append(jobs, &jobCopy)
	}

	return jobs
}