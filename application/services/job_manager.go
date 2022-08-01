package services

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/framework/queue"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
	"github.com/streadway/amqp"
)

type JobManager struct {
	DB               *gorm.DB
	Domain           domain.Job
	MessageChannel   chan amqp.Delivery
	JobReturnChannel chan JobWorkerResult
	RabbitMQ         *queue.RabbitMQ
	Mutex            sync.Mutex
}

type JobNotificationError struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func NewJobManager(
	db *gorm.DB,
	rabbitMQ *queue.RabbitMQ,
	jobReturnChannel chan JobWorkerResult,
	messageChannel chan amqp.Delivery,
) *JobManager {
	job, err := domain.NewJob(local_constants.STORAGE_NAME, domain.NewVideo())
	if err != nil {
		log.Fatalf("failed creating job and video")
	}
	return &JobManager{
		DB:               db,
		Domain:           *job,
		MessageChannel:   messageChannel,
		JobReturnChannel: jobReturnChannel,
		RabbitMQ:         rabbitMQ,
	}
}

func (j *JobManager) Start(ch *amqp.Channel) {
	videoService := NewVideoService()
	videoService.VideoRepository = repositories.VideoRepository{DB: j.DB}
	jobService := JobService{
		JobRepository: repositories.JobRepository{DB: j.DB},
		VideoService:  videoService,
	}

	concurrency, err := strconv.Atoi(os.Getenv(local_constants.CONCURRENCY_WORKERS_ENV))
	if err != nil {
		log.Fatalf("error loading var: CONCURRENCY_WORKERS")
	}

	for qtdProcess := 0; qtdProcess < concurrency; qtdProcess++ {
		go JobWorker(j.MessageChannel, j.JobReturnChannel, jobService, j.Domain, qtdProcess)
	}

	for jobResult := range j.JobReturnChannel {
		if jobResult.Error != nil {
			err = j.checkParseErrors(jobResult)
		} else {
			err = j.notifySuccess(jobResult, ch)
		}

		if err != nil {
			jobResult.Message.Reject(false)
		}
	}
}

func (j *JobManager) notifySuccess(jobResult JobWorkerResult, ch *amqp.Channel) error {
	j.Mutex.Lock()
	jobJson, err := json.Marshal(jobResult.Job)
	j.Mutex.Unlock()
	if err != nil {
		return err
	}

	err = j.notify(jobJson)
	if err != nil {
		return err
	}
	err = jobResult.Message.Ack(false)

	return err
}

func (j *JobManager) checkParseErrors(jobResult JobWorkerResult) error {
	if jobResult.Job.ID != "" {
		log.Panicf(
			"MessageID %v. Error with job: %v with video %v. Error: %v",
			jobResult.Message.DeliveryTag,
			jobResult.Job.ID,
			jobResult.Job.VideoID,
			jobResult.Error.Error(),
		)
	} else {
		log.Panicf(
			"MessageID %v. Error parsing message: %v.",
			jobResult.Message.DeliveryTag,
			jobResult.Error.Error(),
		)
	}

	errorMsg := JobNotificationError{
		Message: string(jobResult.Message.Body),
		Error:   jobResult.Error.Error(),
	}
	jobJson, err := json.Marshal(errorMsg)
	if err != nil {
		return err
	}
	err = j.notify(jobJson)
	if err != nil {
		return err
	}
	err = jobResult.Message.Reject(false)

	return err
}

func (j *JobManager) notify(jobJson []byte) error {
	err := j.RabbitMQ.Notify(
		string(jobJson),
		"application/json",
		os.Getenv(local_constants.RABBIT_MQ_NOTIFICATION_EX_ENV),
		os.Getenv(local_constants.RABBIT_MQ_NOTIFICATION_ROUTING_KEY_ENV),
	)

	return err
}
