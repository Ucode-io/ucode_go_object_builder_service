package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func createFile(t *testing.T) string {
	usage := &nb.CreateFileRequest{
		Title:            fakeData.JobTitle(),
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
func Test_fileRepo_GetSingle(t *testing.T) {
	expectedFile := &nb.File{
		Id:               "1c00c6a0-a706-4059-a667-ee976d183e14",
		Title:            "Product Mobility Agent",
		Description:      "Kimberly Koelpin",
		Tags:             []string{"Aaliyah Effertz IV"},
		Storage:          "Geo Hane",
		FileNameDisk:     "Wunsch",
		FileNameDownload: "Chadrick Schmeler",
		Link:             "http://sawayn.net/khalid",
		FileSize:         4880,
	}

	resp, err := strg.File().GetSingle(context.Background(), &nb.FilePrimaryKey{
		Id: expectedFile.Id,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedFile, resp)
}

func Test_fileRepo_GetList(t *testing.T) {
	req := &nb.GetAllFilesRequest{
		Search: "Product",
		Sort:   "desc",
	}

	resp, err := strg.File().GetList(context.Background(), req)
	fmt.Println(resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Files)
}

func Test_fileRepo_Update(t *testing.T) {
	existingFile := &nb.File{
		Id:               "1c00c6a0-a706-4059-a667-ee976d183e14",
		Title:            "Product Mobility Agent",
		Description:      "Kimberly Koelpin",
		Tags:             []string{"Aaliyah Effertz IV"},
		Storage:          "Geo Hane",
		FileNameDisk:     "Wunsch",
		FileNameDownload: "Chadrick Schmeler",
		Link:             "http://sawayn.net/khalid",
		FileSize:         4880,
	}

	newData := &nb.File{
		Id:               existingFile.Id,
		Title:            "Updated Title",
		Description:      "Updated Description",
		Tags:             []string{"Tag1", "Tag2", "Tag3"},
		Storage:          "Updated Storage",
		FileNameDisk:     "Updated FileNameDisk",
		FileNameDownload: "Updated FileNameDownload",
		Link:             "http://updated-url.com",
		FileSize:         10000,
	}

	err := strg.File().Update(context.Background(), newData)
	assert.NoError(t, err)
}

func TestDelete(t *testing.T) {

	// Construct the FileDeleteRequest with the ID to delete
	deleteRequest := &nb.FileDeleteRequest{Ids: []string{"57a03df1-0187-4a09-ac9d-b1c3349c37cc"}}

	// Attempt to delete the file
	err := strg.File().Delete(context.Background(), deleteRequest)

	// Check if there's any error
	assert.NoError(t, err)

}
