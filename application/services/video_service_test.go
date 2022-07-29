package services_test

import (
	"log"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/application/services"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/framework/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func init() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func prepare() (*domain.Video, repositories.IVideoRepository) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "example.mp4"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepository{DB: db}
	// repo.Insert(video)

	return video, repo
}

func TestVideoService(t *testing.T) {
	video, repo := prepare()

	videoService := services.NewVideoService()
	videoService.Video = video
	videoService.VideoRepository = repo

	err := videoService.Download("encoder-v-bucket")
	require.Nil(t, err)

	err = videoService.Fragment()
	require.Nil(t, err)

	err = videoService.Encode()
	require.Nil(t, err)

	err = videoService.Finish()
	require.Nil(t, err)
}
