package model

import "time"

type GCJobStatus string
type GCJobTrigger string

const (
	GCJobStatusPending   GCJobStatus = "pending"
	GCJobStatusRunning   GCJobStatus = "running"
	GCJobStatusCompleted GCJobStatus = "completed"
	GCJobStatusFailed    GCJobStatus = "failed"
)

const (
	GCJobTriggerAutomatic GCJobTrigger = "automatic"
	GCJobTriggerManual    GCJobTrigger = "manual"
)

type GCJob struct {
	ID         string       `gorm:"primaryKey"`
	Status     GCJobStatus  `gorm:"not null;default:pending"`
	Trigger    GCJobTrigger `gorm:"not null"`
	StartedAt  *time.Time
	FinishedAt *time.Time
	CreatedAt  time.Time

	Logs []GCJobLog `gorm:"foreignKey:JobID"`
}

func (j *GCJob) StartedAtOrZero() time.Time {
	if j.StartedAt == nil {
		return time.Time{}
	}
	return *j.StartedAt
}

func (j *GCJob) FinishedAtOrZero() time.Time {
	if j.FinishedAt == nil {
		return time.Time{}
	}
	return *j.FinishedAt
}

type GCJobLog struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	JobID     string `gorm:"not null"`
	Level     string `gorm:"not null"`
	Message   string `gorm:"not null"`
	CreatedAt time.Time
}

type GCJobListParams struct {
	Page    int
	Limit   int
	Status  GCJobStatus
	Trigger GCJobTrigger
}
