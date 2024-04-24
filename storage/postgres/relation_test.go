package postgres_test

import (
	"context"
	"fmt"
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
		TableTo:                "nickolas",
		TableFrom:              "palma",
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
	relations, err := strg.Relation().GetList(context.Background(), &nb.GetAllRelationsRequest{
		TableSlug: "nannie",
		Limit:     10,
		Offset:    0,
	})
	fmt.Printf("RELATIONS: %+v\n", relations)
	assert.NoError(t, err)
}

func TestGetRelation(t *testing.T) {
	// id := createRelation(t)
	id := "cd9a64f4-8c67-4f0a-92ab-20256c92fe3b"
	relation, err := strg.Relation().GetByID(context.Background(), &nb.RelationPrimaryKey{Id: id})
	fmt.Println("RELATION: ", relation)
	assert.NoError(t, err)
	assert.NotEmpty(t, relation)
}
