package postgres_test

import (
	"context"
	"testing"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

// func deleteView(t *testing.T, id string) {
// 	deleteRequest := &nb.ViewPrimaryKey{Id: id}
// 	err := strg.View().Delete(context.Background(), deleteRequest)
// 	assert.NoError(t, err)
// }

func createView(t *testing.T) string {
	usage := &nb.CreateViewRequest{
		TableSlug:           "STATIC_SLUG",
		Type:                fakeData.City(),
		GroupFields:         []string{CreateRandomId(t)},
		ViewFields:          []string{CreateRandomId(t)},
		MainField:           fakeData.FirstName(),
		Users:               []string{CreateRandomId(t), CreateRandomId(t)},
		Name:                fakeData.FirstName(),
		CalendarFromSlug:    fakeData.FirstName(),
		CalendarToSlug:      fakeData.FirstName(),
		TimeInterval:        int32(fakeData.Rand.Int63n(2132)),
		MultipleInsert:      true,
		StatusFieldSlug:     fakeData.FirstName(),
		IsEditable:          false,
		RelationTableSlug:   fakeData.FirstName(),
		RelationId:          CreateRandomId(t),
		MultipleInsertField: fakeData.FirstName(),
		UpdatedFields:       []string{CreateRandomId(t)},
		TableLabel:          fakeData.FirstName(),
		DefaultLimit:        fakeData.FirstName(),
		Attributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				fakeData.City():        {Kind: &structpb.Value_StringValue{StringValue: fakeData.LastName()}},
				fakeData.Characters(1): {Kind: &structpb.Value_StringValue{StringValue: fakeData.LastName()}},
			},
		},
		DefaultEditable: true,
		Order:           int32(len(fakeData.FirstName())),
		NameUz:          fakeData.FirstName(),
		NameEn:          fakeData.FirstName(),
		Columns:         []string{CreateRandomId(t)},
		AppId:           CreateRandomId(t),
		ProjectId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	view, err := strg.View().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, view)

	return view.Id
}

func TestCreateView(t *testing.T) {
	createView(t)
}

func TestGetSingleView(t *testing.T) {
	viewID := createView(t)

	resp, err := strg.View().GetSingle(context.Background(), &nb.ViewPrimaryKey{Id: viewID})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, viewID, resp.Id)
}

func TestGetListViews(t *testing.T) {

	createView(t)

	req := &nb.GetAllViewsRequest{
		TableSlug: "STATIC_SLUG",
	}
	resp, err := strg.View().GetList(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Views)
}

func TestUpdateView(t *testing.T) {
	viewID := createView(t)
	// defer deleteView(t, viewID)

	_, err := strg.View().Update(context.Background(), &nb.View{
		Id:                  viewID,
		TableSlug:           fakeData.City(),
		Type:                fakeData.City(),
		GroupFields:         []string{CreateRandomId(t)},
		ViewFields:          []string{CreateRandomId(t)},
		MainField:           fakeData.FirstName(),
		Users:               []string{CreateRandomId(t), CreateRandomId(t)},
		Name:                fakeData.FirstName(),
		CalendarFromSlug:    fakeData.FirstName(),
		CalendarToSlug:      fakeData.FirstName(),
		TimeInterval:        int32(fakeData.Rand.Int63n(2132)),
		MultipleInsert:      true,
		StatusFieldSlug:     fakeData.FirstName(),
		IsEditable:          false,
		RelationTableSlug:   fakeData.FirstName(),
		RelationId:          CreateRandomId(t),
		MultipleInsertField: fakeData.FirstName(),
		UpdatedFields:       []string{CreateRandomId(t)},
		TableLabel:          fakeData.FirstName(),
		DefaultLimit:        fakeData.FirstName(),
		Attributes: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				fakeData.Language:      {Kind: &structpb.Value_StringValue{StringValue: ""}},
				fakeData.Characters(1): {Kind: &structpb.Value_StringValue{StringValue: fakeData.Language}},
			},
		},
		DefaultEditable: true,
		Order:           int32(len(fakeData.FirstName())),
		NameUz:          fakeData.FirstName(),
		NameEn:          fakeData.FirstName(),
		Columns:         []string{CreateRandomId(t)},
		AppId:           CreateRandomId(t),
		ProjectId:       "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	})
	assert.NoError(t, err)
}

// func TestDeleteView(t *testing.T) {
// 	viewID := createView(t)

// 	err := strg.View().Delete(context.Background(), &nb.ViewDeleteRequest{Ids: []string{viewID, createView(t)}})
// 	assert.NoError(t, err)
// }
