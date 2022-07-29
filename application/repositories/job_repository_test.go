package repositories_test

import (
	"testing"
	"time"

	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/constants"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/framework/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

func TestJobRepository_InsertFind(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoV := repositories.VideoRepository{DB: db}
	repoV.Insert(video)

	job, err := domain.NewJob("output_path", "status", video)
	require.Nil(t, err)

	repoJ := repositories.JobRepository{DB: db}
	repoJ.Insert(job)

	j, err := repoJ.Find(job.ID)

	require.Nil(t, err)
	require.NotEmpty(t, j.ID)
	require.Equal(t, j.ID, job.ID)
	require.Equal(t, j.VideoID, video.ID)
}

func TestJobRepository_Update(t *testing.T) {
	db := database.NewDBTest()
	defer db.Close()

	video := domain.NewVideo()
	video.ID = uuid.NewV4().String()
	video.FilePath = "path"
	video.CreatedAt = time.Now()

	repoV := repositories.VideoRepository{DB: db}
	repoV.Insert(video)

	job, err := domain.NewJob("output_path", constants.PENDING_STATUS, video)
	require.Nil(t, err)

	repoJ := repositories.JobRepository{DB: db}
	repoJ.Insert(job)

	job.Status = constants.COMPLETE_STATUS
	repoJ.Update(job)

	j, err := repoJ.Find(job.ID)

	require.Nil(t, err)
	require.Equal(t, j.Status, constants.COMPLETE_STATUS)
}
