package models

type SectionFields struct {
	Id    string `json:"id"`
	Order int    `json:"order"`
}
type RelationFields struct {
	Guid             string `json:"guid"`
	RoleId           string `json:"role_id"`
	RelationId       string `json:"relation_id"`
	TableSlug        string `json:"table_slug"`
	ViewPermission   bool   `json:"view_permission"`
	CreatePermission bool   `json:"create_permission"`
	EditPermission   bool   `json:"edit_permission"`
	DeletePermission bool   `json:"delete_permission"`
}

type AutofillField struct {
	FieldFrom     string `json:"field_from"`
	FieldTo       string `json:"field_to"`
	FieldSlug     string `json:"field_slug"`
	TableSlug     string `json:"table_slug"`
	AutoFillTable string `json:"autofill_table"`
	Automatic     bool   `json:"automatic"`
}
