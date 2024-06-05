package postgres_test

import (
	"context"
	"encoding/json"
	"testing"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func createRelation(t *testing.T) string {
	relation, err := strg.Relation().Create(context.Background(), &nb.CreateRelationRequest{
		Id:                     CreateRandomId(t),
		Type:                   config.MANY2ONE,
		TableTo:                "relation3",
		TableFrom:              "relation4",
		RelationFieldId:        CreateRandomId(t),
		RelationToFieldId:      CreateRandomId(t),
		ViewFields:             []string{},
		RelationFieldSlug:      "",
		DynamicTables:          []*nb.DynamicTable{},
		Editable:               false,
		IsUserIdDefault:        false,
		Cascadings:             []*nb.Cascading{},
		ObjectIdFromJwt:        false,
		CascadingTreeTableSlug: "",
		CascadingTreeFieldSlug: "",
		Attributes:             &structpb.Struct{},
		GroupFields:            []string{},
		QuickFilters:           []*nb.QuickFilter{},
		Columns:                []string{},
		MultipleInsert:         false,
		IsEditable:             false,
		MultipleInsertField:    "",
		UpdatedFields:          []string{},
		DefaultLimit:           "",
		DefaultValues:          []string{},
		DefaultEditable:        false,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, relation)

	return relation.Id
}

func TestCreateRelation(t *testing.T) {
	createRelation(t)
}

func TestGetListRelations(t *testing.T) {
	_, err := strg.Relation().GetList(context.Background(), &nb.GetAllRelationsRequest{
		TableSlug: "nannie",
		Limit:     10,
		Offset:    0,
	})
	assert.NoError(t, err)
}

func TestGetRelation(t *testing.T) {
	// id := createRelation(t)
	id := "f435f72f-7ab0-4b28-831a-ed43c647c8a8"
	relation, err := strg.Relation().GetByID(context.Background(), &nb.RelationPrimaryKey{Id: id})
	json.Marshal(relation)
	assert.NoError(t, err)
	assert.NotEmpty(t, relation)
}

func TestDeleteRelation(t *testing.T) {
	err := strg.Relation().Delete(context.Background(), &nb.RelationPrimaryKey{Id: "fa954c9d-2dab-4f81-87c9-6c4d221ff81d"})
	assert.NoError(t, err)
}
