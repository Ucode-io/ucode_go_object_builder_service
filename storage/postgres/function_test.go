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
		Type:             fakeData.Name(),
		Description:      fakeData.Name(),
		ProjectId:        CreateRandomId(t),
		EnvironmentId:    CreateRandomId(t),
		FunctionFolderId: CreateRandomId(t),
		Url:              fakeData.URL(),
		Password:         fakeData.CellPhoneNumber(),
		SshUrl:           fakeData.URL(),
		GitlabId:         CreateRandomId(t),
		GitlabGroupId:    CreateRandomId(t),
	}

	function, err := strg.Function().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, function)

	return function.Id
}

func TestCreateFunction(t *testing.T) {
	createFunction(t)
}
