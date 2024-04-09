package postgres_test

import (
	"context"
	"testing"

	pb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func createFunction(t *testing.T) string {
	usage := &pb.CreateFunctionRequest{
		Path:             fakeData.City(),
		Name:             fakeData.Name(),
		Type:             "MICRO_FRONTEND",
		Description:      fakeData.Name(),
		ProjectId:        CreateRandomId(t),
		EnvironmentId:    CreateRandomId(t),
		FunctionFolderId: CreateRandomId(t),
		Url:              fakeData.URL(),
		Password:         fakeData.CellPhoneNumber(),
		SshUrl:           fakeData.URL(),
		GitlabId:         fakeData.Country(),
		GitlabGroupId:    fakeData.Country(),
	}

	function, err := strg.Function().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, function)

	return function.Id
}

func TestCreateFunction(t *testing.T) {
	createFunction(t)
}

func TestGetSingleFunc(t *testing.T) {
	funcPk := createFunction(t)

	resp, err := strg.Function().GetSingle(context.Background(), &pb.FunctionPrimaryKey{Id: funcPk})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, funcPk, resp.Id)
}

func TestGetListFunc(t *testing.T) {

	req := &pb.GetAllFunctionsRequest{
		Type: "FUNCTION",
	}
	resp, err := strg.Function().GetList(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Functions)
}

func TestUpdateFunc(t *testing.T) {
	fileID := createFunction(t)

	
	err := strg.Function().Update(context.Background(), &pb.Function{
		Id:               fileID,
		EnvironmentId:    CreateRandomId(t),
		ProjectId:        CreateRandomId(t),
		FunctionFolderId: CreateRandomId(t),
		Type:             "MICRO_FRONTEND",
	})
	assert.NoError(t, err)
}

func TestDeleteFunc(t *testing.T) {
	funcPk := createFunction(t)

	err := strg.Function().Delete(context.Background(), &pb.FunctionPrimaryKey{Id: funcPk})
	assert.NoError(t, err)
}
