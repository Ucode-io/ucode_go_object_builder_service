package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func TestGetAllMenuPermission(t *testing.T) {
	menus, err := strg.Permission().GetAllMenuPermissions(
		context.Background(),
		&nb.GetAllMenuPermissionsRequest{
			ParentId: "a70be675-2c58-4efe-80e2-40dfdf8c14ea",
			RoleId:   "9a31fec6-1cd3-477a-ab7d-11a4281222bb",
		},
	)
	assert.NoError(t, err)
	fmt.Println("Menus->", menus)
}
