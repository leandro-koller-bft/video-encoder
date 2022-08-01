package services

import (
	"encoding/json"

	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/framework/utils"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type JobWorkerResult struct {
	Job     domain.Job
	Message *amqp.Delivery
	Error   error
}

func JobWorker(
	messageChannel chan amqp.Delivery,
	returnChannel chan JobWorkerResult,
	jobService JobService,
	job domain.Job,
	workerID int) {

	for message := range messageChannel {
		// validate if body is a json
		err := utils.IsJson(string(message.Body))
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		// get message from body
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		// validate video
		err = jobService.VideoService.Video.Validate()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		// insert video on database
		err = jobService.VideoService.InsertVideo()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		// prepare job
		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = local_constants.STORAGE_NAME
		job.Status = local_constants.JOB_STARTING

		_, err = jobService.JobRepository.Insert(&job)
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}
		jobService.Job = &job
		// start job
		err = jobService.Start()
		if err != nil {
			returnChannel <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		returnChannel <- returnJobResult(job, message, nil)
	}
}

func returnJobResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	return JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err,
	}
}
