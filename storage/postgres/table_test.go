package postgres_test

import (
	"context"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func createTable(t *testing.T) string {

	usage := &nb.CreateTableRequest{
		ProjectId:         "6075011d-7191-4d4d-9d45-76cdbe998b32",
		Slug:              "okay",
		Label:             "okay",
		Icon:              "okay.svg",
		Description:       "okay",
		ShowInMenu:        true,
		SubtitleFieldSlug: "okay_item",
		IsCached:          true,
		IncrementId: &nb.IncrementID{
			WithIncrementId: false,
			DigitNumber:     0,
		},
		SoftDelete: false,
	}

	table, err := strg.Table().Create(context.Background(), usage)
	assert.NoError(t, err)
	assert.NotEmpty(t, table)

	return table.Id
}

func TestCreateTable(t *testing.T) {
	createTable(t)
}

func TestTableGetSingle(t *testing.T) {
	expectedTable := &nb.Table{
		Id:                "401dd843-0f3e-474d-a6f6-6dd2770d6e93",
		Slug:              "okay",
		Label:             "okay",
		Icon:              "okay.svg",
		Description:       "okay",
		ShowInMenu:        true,
		SubtitleFieldSlug: "okay_item",
		IsCached:          true,
		IncrementId: &nb.IncrementID{
			DigitNumber:     0,
			WithIncrementId: false,
		},
	}

	resp, err := strg.Table().GetByID(context.Background(), &nb.TablePrimaryKey{Id: expectedTable.Id, ProjectId: ""})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedTable, resp)
}

func TestTableUpdate(t *testing.T) {

	data := &nb.UpdateTableRequest{
		Id:                "401dd843-0f3e-474d-a6f6-6dd2770d6e93",
		Slug:              "okay",
		Label:             "okay",
		Icon:              "okay.svg",
		Description:       "okay",
		ShowInMenu:        true,
		SubtitleFieldSlug: "okay_item",
		IsCached:          true,
		IncrementId: &nb.IncrementID{
			DigitNumber:     0,
			WithIncrementId: false,
		},
	}

	newData, err := strg.Table().Update(context.Background(), data)
	assert.NoError(t, err)
	assert.NotEmpty(t, newData)
}

func TestDeleteTable(t *testing.T) {
	deleteReq := &nb.TablePrimaryKey{Id: "54be8e02-19a0-4194-ac62-a5c579f68b57"}

	err := strg.Table().Delete(context.Background(), deleteReq)

	assert.NoError(t, err)
}
