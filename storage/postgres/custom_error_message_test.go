package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func CreateCustomErrorMessage(t *testing.T) string {
	usage := &nb.CreateCustomErrorMessage{
		TableId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Message:    fakeData.Name(),
		ErrorId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Code:       32,
		LanguageId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		ActionType: fakeData.Name(),
	}

	file, err := strg.CustomErrorMessage().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, file)

	return file.Id
}

func TestCreateCustomErr(t *testing.T) {
	CreateCustomErrorMessage(t)
}
func Test_CustomRepo_GetSingle(t *testing.T) {
	expectedFile := &nb.CustomErrorMessage{
		Id:         "9ca2f22a-ae8a-45f5-ab44-e3b6120d01ea",
		TableId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Message:    "Lydia Watsica",
		ErrorId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Code:       32,
		LanguageId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		ActionType: "Antwon Kshlerin II",
	}

	resp, err := strg.CustomErrorMessage().GetSingle(context.Background(), &nb.CustomErrorMessagePK{
		Id: expectedFile.Id,
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedFile, resp)
}

func Test_customMessRepo_GetList(t *testing.T) {
	req := &nb.GetCustomErrorMessageListRequest{}

	resp, err := strg.CustomErrorMessage().GetList(context.Background(), req)
	fmt.Println(resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.CustomErrorMessages)
}

func Test_customMessRepo_Update(t *testing.T) {
	existingFile := &nb.CustomErrorMessage{
		Id:         "9ca2f22a-ae8a-45f5-ab44-e3b6120d01ea",
		TableId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Message:    "update",
		ErrorId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Code:       3,
		LanguageId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		ActionType: "Antwon Kshlerin II",
	}

	newData := &nb.CustomErrorMessage{
		Id:         existingFile.Id,
		TableId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Message:    "update",
		ErrorId:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		Code:       3,
		LanguageId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		ActionType: "Antwon Kshlerin II",
	}

	err := strg.CustomErrorMessage().Update(context.Background(), newData)
	assert.NoError(t, err)
	assert.NotEmpty(t, newData)
}

func TestDeleteCustomMessage(t *testing.T) {

	// Construct the FileDeleteRequest with the ID to delete
	deleteRequest := &nb.CustomErrorMessagePK{Id: "9ca2f22a-ae8a-45f5-ab44-e3b6120d01ea"}

	// Attempt to delete the file
	err := strg.CustomErrorMessage().Delete(context.Background(), deleteRequest)

	// Check if there's any error
	assert.NoError(t, err)

}
