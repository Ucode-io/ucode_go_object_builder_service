package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func createMenu(t *testing.T) (*nb.Menu, error) {
	name := fakeData.Name()
	usage := &nb.CreateMenuRequest{
		Label:    name,
		Icon:     "apple-whole.svg",
		ParentId: "c57eedc3-a954-4262-a0af-376c65b5a284",
		Type:     "FOLDER",
		Attributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"label":    {Kind: &structpb.Value_StringValue{StringValue: ""}},
				"label_en": {Kind: &structpb.Value_StringValue{StringValue: name}},
			},
		},
	}

	menu, err := strg.Menu().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, menu)

	return menu, nil
}

func TestCreateMenu(t *testing.T) {
	menu, err := createMenu(t)
	assert.NoError(t, err)
	assert.NotNil(t, menu)
	fmt.Println(menu)
}

func TestMenuGetById(t *testing.T) {
	req := &nb.MenuPrimaryKey{Id: "fa9a8ce6-47a9-440c-8d90-6359666f3e7f"}
	resp, err := strg.Menu().GetById(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	fmt.Println("Response->", resp)
}

func TestDeleteMenu(t *testing.T) {
	id := "c57eedc3-a954-4262-a0af-376c65b5a280"

	err := strg.Menu().Delete(context.Background(), &nb.MenuPrimaryKey{Id: id})
	assert.NoError(t, err)
}

// {
// 	"label": "Hello World",
// 	"icon": "apple-whole.svg",
// 	"parent_id": "c57eedc3-a954-4262-a0af-376c65b5a284",
// 	"type": "FOLDER",
// 	"id": "0a293a8d-c522-4cc7-9efb-73a0dee00dca",
// 	"attributes": {
// 	  "label": "",
// 	  "label_en": "Hello World"
// 	}
//   }
