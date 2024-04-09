package postgres_test

// import (
// 	"context"
// 	"testing"
// 	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

// 	"github.com/stretchr/testify/assert"
// )

// func CreateCustomErrorMessage(t *testing.T) string {
// 	usage := &nb.CreateCustomErrorMessage{
// 		TableId:    "990825b6-6243-44de-944c-769a4b89a33e", // change to dynamic
// 		Message:    fakeData.Name(),
// 		ErrorId:    CreateRandomId(t),
// 		Code:       32,
// 		LanguageId: CreateRandomId(t),
// 		ActionType: fakeData.Name(),
// 	}

// 	customErrMes, err := strg.CustomErrorMessage().Create(context.Background(), usage)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, customErrMes)

// 	return customErrMes.Id
// }

// func TestCreateCustomErr(t *testing.T) {
// 	CreateCustomErrorMessage(t)
// }
// func TestGetSingleCustomErrMess(t *testing.T) {
// 	customErrMessId := CreateCustomErrorMessage(t)

// 	resp, err := strg.CustomErrorMessage().GetSingle(context.Background(), &nb.CustomErrorMessagePK{Id: customErrMessId})
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.Equal(t, customErrMessId, resp.Id)
// }

// func TestGetListCustomErrMess(t *testing.T) {

// 	req := &nb.GetCustomErrorMessageListRequest{
// 		TableId: "990825b6-6243-44de-944c-769a4b89a33e", // change to dynamic
// 	}
// 	resp, err := strg.CustomErrorMessage().GetList(context.Background(), req)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotEmpty(t, resp.CustomErrorMessages)
// }

// func Test_customMessRepo_Update(t *testing.T) {
// 	CreateCustomErrorMessage(t)

// 	newData := &nb.CustomErrorMessage{
// 		Id:         CreateCustomErrorMessage(t),
// 		TableId:    CreateRandomId(t),
// 		Message:    "update",
// 		ErrorId:    CreateRandomId(t),
// 		Code:       33,
// 		LanguageId: CreateRandomId(t),
// 		ActionType: "Antwon Kshlerin II",
// 	}

// 	err := strg.CustomErrorMessage().Update(context.Background(), newData)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, newData)
// }

// func TestDeleteCustomErrMessage(t *testing.T) {
// 	customErrMess := CreateCustomErrorMessage(t)

// 	err := strg.CustomErrorMessage().Delete(context.Background(), &nb.CustomErrorMessagePK{Id: customErrMess})
// 	assert.NoError(t, err)
// }

// func TestGetListForObjectRequestt(t *testing.T) {

// 	req := &nb.GetListForObjectRequest{
// 		TableId: "990825b6-6243-44de-944c-769a4b89a33e", // change to dynamic
// 	}
// 	resp, err := strg.CustomErrorMessage().GetListForObject(context.Background(), req)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, resp)
// 	assert.NotEmpty(t, resp.CustomErrorMessages)
// }
