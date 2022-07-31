package services

import (
	"errors"
	"os"
	"strconv"

	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
)

type JobService struct {
	Job           *domain.Job
	JobRepository repositories.IJobRepository
	VideoService  VideoService
}

func (j *JobService) Start() error {
	err := j.changeJobStatus(local_constants.JOB_DOWNLOADING)
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Download(local_constants.STORAGE_NAME)
	if err != nil {
		return j.failJob(err)
	}
	err = j.changeJobStatus(local_constants.JOB_FRAGMENTING)
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Fragment()
	if err != nil {
		return j.failJob(err)
	}
	err = j.changeJobStatus(local_constants.JOB_ENCODING)
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Encode()
	if err != nil {
		return j.failJob(err)
	}

	err = j.performUpload()
	if err != nil {
		return j.failJob(err)
	}

	err = j.changeJobStatus(local_constants.JOB_FINISHING)
	if err != nil {
		return j.failJob(err)
	}
	err = j.VideoService.Finish()
	if err != nil {
		return j.failJob(err)
	}
	err = j.changeJobStatus(local_constants.JOB_COMPLETED)
	if err != nil {
		return j.failJob(err)
	}

	return nil
}

func (j *JobService) performUpload() error {
	err := j.changeJobStatus(local_constants.JOB_UPLOADING)
	if err != nil {
		return j.failJob(err)
	}

	videoUpload := NewVideoUpload()
	videoUpload.OutputBucket = local_constants.STORAGE_NAME
	videoUpload.VideoPath = os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + j.VideoService.Video.ID
	concurrency, _ := strconv.Atoi(os.Getenv(local_constants.CONCURRENCY_ENV))
	doneUpload := make(chan string)

	go videoUpload.ProcessUpload(concurrency, doneUpload)

	uploadResult := <-doneUpload

	if uploadResult != local_constants.UPLOAD_COMPLETE_MSG {
		return j.failJob(errors.New(uploadResult))
	}

	return err
}

func (j *JobService) changeJobStatus(status string) error {
	var err error

	j.Job.Status = status
	j.Job, err = j.JobRepository.Update(j.Job)
	if err != nil {
		return j.failJob(err)
	}
	return nil
}

func (j *JobService) failJob(err error) error {
	j.Job.Status = local_constants.JOB_FAILED
	j.Job.Error = err.Error()

	_, err = j.JobRepository.Update(j.Job)

	return err
}
