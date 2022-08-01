package domain

import (
	"time"

	"github.com/leandro-koller-bft/video-encoder/local_constants"
	uuid "github.com/satori/go.uuid"

	"github.com/asaskevich/govalidator"
)

type Job struct {
	ID               string    `json:"job_id" valid:"uuid" gorm:"type:uuid;primary_key"`
	OutputBucketPath string    `json:"output_bucket_path" valid:"notnull"`
	Status           string    `json:"status" valid:"notnull"`
	Video            *Video    `json:"video" valid:"-"`
	VideoID          string    `json:"-" valid:"-" gorm:"column:video_id;type:uuid;notnull"`
	Error            string    `json:"error" valid:"-"`
	CreatedAt        time.Time `json:"created_at" valid:"-"`
	UpdatedAt        time.Time `json:"updated_at" valid:"-"`
}

func NewJob(output string, video *Video) (*Job, error) {
	job := Job{
		OutputBucketPath: output,
		Video:            video,
	}
	job.prepare()

	err := job.Validate()
	if err != nil {
		return nil, err
	}

	return &job, nil
}

func (job *Job) prepare() {
	job.ID = uuid.NewV4().String()
	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	job.Status = local_constants.JOB_STARTING
}

func (job *Job) Validate() error {
	_, err := govalidator.ValidateStruct(job)
	if err != nil {
		return err
	}

	return nil
}
