package utils_test

import (
	"testing"

	"github.com/leandro-koller-bft/video-encoder/framework/utils"
	"github.com/stretchr/testify/require"
)

func TestIs(t *testing.T) {
	json := `{
		"id": "not-a-id",
		"file_path": "not/a/path",
		"isValid": true
	}`

	err := utils.IsJson(json)
	require.Nil(t, err)

	json = `was`
	err = utils.IsJson(json)
	require.Error(t, err)
}
