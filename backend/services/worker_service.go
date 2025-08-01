package services

import (
	"encoding/json"
	"fmt"
	"github.com/dedegunawan/backend-ujian-telp-v5/models"
	"github.com/dedegunawan/backend-ujian-telp-v5/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type WorkerService interface {
	StartWorker(workerName string)
	ProcessJob(job *models.JobQueue) error
	EnqueueJob(jobType string, payload interface{}, delay time.Duration) error
}

type workerService struct {
	db *gorm.DB
}

func NewWorkerService(db *gorm.DB) WorkerService {
	return &workerService{db: db}
}

func (ws *workerService) StartWorker(workerName string) {
	db := ws.db
	for {
		tx := db.Begin()

		var job models.JobQueue
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}). // SKIP LOCKED
			Where("status = ? AND run_at <= now()", "queued").
			Order("run_at").
			Limit(1).
			Take(&job).Error

		if err != nil {
			tx.Rollback()
			time.Sleep(1 * time.Second)
			continue
		}

		job.Status = "processing"
		job.UpdatedAt = time.Now()
		tx.Save(&job)
		tx.Commit()

		go func(job models.JobQueue) {
			log.Printf("[%s] Memproses job #%d - %s\n", workerName, job.ID, job.Type)
			err := ws.ProcessJob(&job)
			if err != nil {
				job.Retries++
				job.LastError = utils.Ptr(err.Error())
				if job.Retries >= job.MaxRetries {
					job.Status = "failed"
				} else {
					job.Status = "queued"
					job.RunAt = time.Now().Add(30 * time.Second) // retry delay
				}
			} else {
				job.Status = "done"
			}
			job.UpdatedAt = time.Now()
			db.Save(&job)
		}(job)

		time.Sleep(200 * time.Millisecond) // tunggu sebentar sebelum ambil job lagi
	}
}

func (ws *workerService) ProcessJob(job *models.JobQueue) error {
	switch job.Type {
	case "send_email":
		var payload map[string]interface{}
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}
		email := payload["to"].(string)
		// Kirim email...
		log.Println("Mengirim email ke:", email)
		return nil
	default:
		return fmt.Errorf("unknown job type: %s", job.Type)
	}
}

func (ws *workerService) EnqueueJob(jobType string, payload interface{}, delay time.Duration) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	job := models.JobQueue{
		Type:       jobType,
		Payload:    datatypes.JSON(jsonPayload),
		Status:     "queued",
		Retries:    0,
		MaxRetries: 3,
		RunAt:      time.Now().Add(delay),
	}
	db := ws.db
	return db.Create(&job).Error
}
