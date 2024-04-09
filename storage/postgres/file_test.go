package postgres_test

import (
	"context"
	"testing"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func deleteFile(t *testing.T, id string) {
	deleteRequest := &nb.FileDeleteRequest{Ids: []string{id}}
	err := strg.File().Delete(context.Background(), deleteRequest)
	assert.NoError(t, err)
}

func createFile(t *testing.T) string {
	usage := &nb.CreateFileRequest{
		Title:            "Product",
		Description:      fakeData.Name(),
		Tags:             []string{fakeData.Name()},
		Storage:          fakeData.Name(),
		FileNameDisk:     fakeData.LastName(),
		FileNameDownload: fakeData.Name(),
		Link:             fakeData.URL(),
		FileSize:         fakeData.Rand.Int63n(12312),
		ProjectId:        "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	file, err := strg.File().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	return file.Id
}

func TestCreateFile(t *testing.T) {
	createFile(t)
}

func TestGetSingleFile(t *testing.T) {
	fileID := createFile(t)

	resp, err := strg.File().GetSingle(context.Background(), &nb.FilePrimaryKey{Id: fileID})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, fileID, resp.Id)
}

func TestGetListFiles(t *testing.T) {
	req := &nb.GetAllFilesRequest{
		Search: "Product",
		Sort:   "desc",
	}
	resp, err := strg.File().GetList(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Files)
}

func TestUpdateFile(t *testing.T) {
	fileID := createFile(t)
	defer deleteFile(t, fileID)

	err := strg.File().Update(context.Background(), &nb.File{
		Id:    fileID,
		Title: "Updated Title",
	})
	assert.NoError(t, err)
}

func TestDeleteFile(t *testing.T) {
	fileID := createFile(t)

	err := strg.File().Delete(context.Background(), &nb.FileDeleteRequest{Ids: []string{fileID, createFile(t)}})
	assert.NoError(t, err)
}
