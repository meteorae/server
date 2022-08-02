package database

import "time"

type JobType string

func (t JobType) String() string {
	return string(t)
}

var (
	JobTypeLibraryScan    JobType = "library.scan"
	JobTypeMetadataUpdate JobType = "metadata.update"
	JobTypeMediaAnalysis  JobType = "media.analysis"
)

type JobStatus string

func (s JobStatus) String() string {
	return string(s)
}

var (
	JobStatusPending  JobStatus = "pending"
	JobStatusRunning  JobStatus = "running"
	JobStatusComplete JobStatus = "complete"
	JobStatusFailed   JobStatus = "failed"
)

type Job struct {
	ID        uint `gorm:"primary_key"`
	JobType   JobType
	Status    JobStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func CreateJob(job *Job) error {
	if result := db.Create(&job); result.Error != nil {
		return result.Error
	}

	return nil
}

func UpdateJobStatus(job *Job, status JobStatus) error {
	job.Status = status

	if result := db.Save(&job); result.Error != nil {
		return result.Error
	}

	return nil
}
