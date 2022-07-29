package repositories_test

import (
	"testing"
	"time"

	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/framework/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestVideoRepository_InsertFind(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repo := repositories.VideoRepository{DB: db}
	repo.Insert(video)

	v, err := repo.Find(video.ID)

	require.Nil(t, err)
	require.NotEmpty(t, v.ID)
	require.Equal(t, v.ID, video.ID)
}
