package postgres_test

import (
	"context"
	"fmt"
	"testing"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

// func deleteLayout(t *testing.T, id string) {
// 	deleteRequest := &nb.LayoutPrimaryKey{Id: id}
// 	err := strg.Layout().Delete(context.Background(), deleteRequest)
// 	assert.NoError(t, err)``
// }

func createLayout(t *testing.T) string {
	usage := &nb.LayoutRequest{
		Id:      CreateRandomId(t),
		Label:   fakeData.Email(),
		Order:   232,
		TableId: "40958790-0d6f-4a56-bc19-cf1d0d6f7372",
		Type:    "relation",
		Tabs: []*nb.TabRequest{
			{
				Label:      fakeData.City(),
				Type:       "relation",
				Order:      1,
				RelationId: "c168f2df-c6a2-4921-8c56-56e708a8d766",
				Sections: []*nb.Section{
					{
						Label: fakeData.City(),
						Order: 1,
						Fields: []*nb.FieldForSection{
							{
								Id:           CreateRandomId(t),
								Order:        1,
								RelationType: "Many2One",
							},
						},
					},
				},
			},
		},
		Icon:             CreateRandomId(t),
		IsDefault:        true,
		IsModal:          false,
		IsVisibleSection: true,
		Attributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				fakeData.City():        {Kind: &structpb.Value_StringValue{StringValue: fakeData.LastName()}},
				fakeData.Characters(1): {Kind: &structpb.Value_StringValue{StringValue: fakeData.LastName()}},
			},
		},
		MenuId: "12d1af5a-2938-4160-88df-933182cea688",

		ProjectId: "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	layout, err := strg.Layout().Update(context.Background(), usage)
	assert.NoError(t, err)
	fmt.Println(layout.Id, layout)
	return layout.Id
}

func TestCreateLayout(t *testing.T) {
	createLayout(t)
}

func TestGetSingleLayout(t *testing.T) {
	// layoutID := createLayout(t)

	resp, err := strg.Layout().GetSingleLayout(context.Background(), &nb.GetSingleLayoutRequest{
		TableId:         "19a864bb-9fc9-46a9-a7a2-8e89e4842fcb",
		RoleId:          "b8190175-d6cb-4631-906c-eb015d81a930",
		MenuId:          "0afb3303-2cc1-463e-9b63-7a49f5e75997",
		LanguageSetting: "en",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	fmt.Println(resp)
}

func TestGetListLayouts(t *testing.T) {

	req := &nb.GetListLayoutRequest{
		TableId: "40958790-0d6f-4a56-bc19-cf1d0d6f7372",
		MenuId:  "12d1af5a-2938-4160-88df-933182cea688",
		RoleId:  "9a31fec6-1cd3-477a-ab7d-11a4281222bb",
	}
	resp, err := strg.Layout().GetAll(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Layouts)
}

// func TestUpdateLayout(t *testing.T) {
// 	layoutID := createLayout(t)
// 	// defer deleteLayout(t, layoutID)

// 	_, err := strg.Layout().Update(context.Background(), &nb.Layout{
// 		Id:                  layoutID,
// 		TableSlug:           fakeData.City(),
// 		Type:                fakeData.City(),
// 		GroupFields:         []string{CreateRandomId(t)},
// 		LayoutFields:        []string{CreateRandomId(t)},
// 		MainField:           fakeData.FirstName(),
// 		Users:               []string{CreateRandomId(t), CreateRandomId(t)},
// 		Name:                fakeData.FirstName(),
// 		CalendarFromSlug:    fakeData.FirstName(),
// 		CalendarToSlug:      fakeData.FirstName(),
// 		TimeInterval:        int32(fakeData.Rand.Int63n(2132)),
// 		MultipleInsert:      true,
// 		StatusFieldSlug:     fakeData.FirstName(),
// 		IsEditable:          false,
// 		RelationTableSlug:   fakeData.FirstName(),
// 		RelationId:          CreateRandomId(t),
// 		MultipleInsertField: fakeData.FirstName(),
// 		UpdatedFields:       []string{CreateRandomId(t)},
// 		TableLabel:          fakeData.FirstName(),
// 		DefaultLimit:        fakeData.FirstName(),
// 		Attributes: &structpb.Struct{
// 			Fields: map[string]*structpb.Value{
// 				fakeData.Language:      {Kind: &structpb.Value_StringValue{StringValue: ""}},
// 				fakeData.Characters(1): {Kind: &structpb.Value_StringValue{StringValue: fakeData.Language}},
// 			},
// 		},
// 		DefaultEditable: true,
// 		Order:           int32(len(fakeData.FirstName())),
// 		NameUz:          fakeData.FirstName(),
// 		NameEn:          fakeData.FirstName(),
// 		Columns:         []string{CreateRandomId(t)},
// 		AppId:           CreateRandomId(t),
// 		ProjectId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
// 	})
// 	assert.NoError(t, err)
// }

// func TestDeleteLayout(t *testing.T) {
// 	layoutID := createLayout(t)

// 	err := strg.Layout().Delete(context.Background(), &nb.LayoutDeleteRequest{Ids: []string{layoutID, createLayout(t)}})
// 	assert.NoError(t, err)
// }
