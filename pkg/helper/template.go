package helper

import (
	"ucode/ucode_go_object_builder_service/models"
)

func CreateTemplate(roleID string) []models.FieldPermission {
	templates := []models.FieldPermission{
		{
			FieldId:        "834df8ef-edb7-4170-996c-9bd5431d9a62",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "789fdf49-5387-4cca-926f-1e963aaee7e0",
			RoleId:         roleID,
			Label:          "Таблица",
		},
		{
			FieldId:        "5dda58a1-84ac-4c50-8993-02e2cefcb29a",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "a14724d7-772b-435a-9aec-55328126c766",
			RoleId:         roleID,
			Label:          "Размер",
		},
		{
			FieldId:        "9772b679-33ec-4004-b527-317a1165575e",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "37d485d4-7cf4-4e35-9cc5-c0c4476712fd",
			RoleId:         roleID,
			Label:          "Название",
		},
		{
			FieldId:        "98279b02-10c0-409e-8303-14224fd76ec6",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "3dbc32ed-8842-4c6e-93bd-f4b01ae9e631",
			RoleId:         roleID,
			Label:          "HTML",
		},
		{
			FieldId:        "494e1ad3-fce8-4e6c-921f-850d0ec73cc4",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "21c47566-b873-4152-ab36-5a510d1aa725",
			RoleId:         roleID,
			Label:          "ID",
		},
		{
			FieldId:        "fd7f0fde-3de7-4073-b64d-bd3076c6e3fb",
			TableSlug:      "template",
			ViewPermission: true,
			EditPermission: true,
			Guid:           "b91908f9-6c77-4211-84d0-134f10094ae7",
			RoleId:         roleID,
			Label:          "FROM VersionTable2.1 TO template",
		},
	}

	return templates
}
