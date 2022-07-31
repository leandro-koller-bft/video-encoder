package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"cloud.google.com/go/storage"
	"github.com/leandro-koller-bft/video-encoder/application/repositories"
	"github.com/leandro-koller-bft/video-encoder/domain"
	"github.com/leandro-koller-bft/video-encoder/local_constants"
)

type VideoService struct {
	Video           *domain.Video
	VideoRepository repositories.IVideoRepository
}

func NewVideoService() VideoService {
	return VideoService{}
}

func (v *VideoService) Download(bucketName string) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	bkt := client.Bucket(bucketName)
	obj := bkt.Object(v.Video.FilePath)

	r, err := obj.NewReader(ctx)
	if err != nil {
		return err
	}
	defer r.Close()

	body, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f, err := os.Create(os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID + ".mp4")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(body)
	if err != nil {
		return err
	}

	log.Printf("video %v has been stored", v.Video.ID)

	return nil
}

func (v *VideoService) Fragment() error {
	err := os.Mkdir(os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV)+"/"+v.Video.ID, os.ModePerm)
	if err != nil {
		return err
	}
	fmt.Println(v.Video.FilePath)

	source := os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID + ".mp4"
	target := os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID + ".frag"
	cmd := exec.Command("mp4fragment", source, target)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	printOutput(output)

	return nil
}

func (v *VideoService) Encode() error {
	cmdArgs := []string{}
	cmdArgs = append(cmdArgs, os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV)+"/"+v.Video.ID+".frag")
	cmdArgs = append(cmdArgs, "--use-segment-timeline")
	cmdArgs = append(cmdArgs, "-o")
	cmdArgs = append(cmdArgs, os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV)+"/"+v.Video.ID)
	cmdArgs = append(cmdArgs, "-f")

	cmd := exec.Command("/usr/local/bin/wrappers/mp4dash", cmdArgs...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	printOutput(output)

	return nil
}

func (v *VideoService) Finish() error {
	err := os.Remove(os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID + ".mp4")
	if err != nil {
		log.Println("error removing ", v.Video.ID, ".mp4")
		return err
	}

	err = os.Remove(os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID + ".frag")
	if err != nil {
		log.Println("error removing ", v.Video.ID, ".frag")
		return err
	}

	err = os.RemoveAll(os.Getenv(local_constants.LOCAL_STORAGE_PATH_ENV) + "/" + v.Video.ID)
	if err != nil {
		log.Println("error removing /", v.Video.ID)
		return err
	}

	log.Println("files have been removed: " + v.Video.ID)

	return nil
}

func printOutput(out []byte) {
	if len(out) > 0 {
		log.Printf("=====> Output: %s\n", string(out))
	}
}
