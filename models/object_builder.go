package models

import (
	"google.golang.org/protobuf/types/known/structpb"
)

type ClientType struct {
	Guid              string   `json:"guid"`
	ProjectId         string   `json:"project_id"`
	Name              string   `json:"name"`
	SelfRegister      bool     `json:"self_register"`
	SelfRecover       bool     `json:"self_recover"`
	ClientPlatformIds []string `json:"client_platform_ids"`
	ConfirmBy         string   `json:"confirm_by"`
	IsSystem          bool     `json:"is_system"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	TableSlug         string   `json:"table_slug"`
	DefaultPage       string   `json:"default_page"`
}

type Connection struct {
	Guid          string `json:"guid"`
	TableSlug     string `json:"table_slug"`
	ViewSlug      string `json:"view_slug"`
	ViewLabel     string `json:"view_label"`
	Name          string `json:"name"`
	Type          string `json:"type"`
	Icon          string `json:"icon"`
	MainTableSlug string `json:"main_table_slug"`
	FieldSlug     string `json:"field_slug"`
	ClientTypeId  string `json:"client_type_id"`
}

type Role struct {
	Guid             string `json:"guid"`
	Name             string `json:"name"`
	ProjectId        string `json:"project_id"`
	ClientPlatformId string `json:"client_platform_id"`
	ClientTypeId     string `json:"client_type_id"`
	IsSystem         bool   `json:"is_system"`
}

type ClientPlatform struct {
	Guid      string `json:"guid"`
	ProjectId string `json:"project_id"`
	Name      string `json:"name"`
	Subdomain string `json:"subdomain"`
}

type Field struct {
	Id                  string           `json:"id"`
	TableId             string           `json:"table_id"`
	TableSlug           string           `json:"table_slug"`
	Required            bool             `json:"required"`
	Slug                string           `json:"slug"`
	Label               string           `json:"label"`
	Default             string           `json:"default"`
	Type                string           `json:"type"`
	Index               string           `json:"index"`
	Attributes          *structpb.Struct `json:"attributes"`
	IsVisible           bool             `json:"is_visible"`
	AutofillField       string           `json:"autofill_field"`
	AutofillTable       string           `json:"autofill_table"`
	Unique              bool             `json:"unique"`
	Automatic           bool             `json:"automatic"`
	RelationId          string           `json:"relation_id"`
	ViewFields          []Field          `json:"view_fields"`
	PathSlug            string           `json:"path_slug"`
	EnableMultilanguage bool             `json:"enable_multilanguage"`
	Column              int32            `json:"column"`
	RelationType        string           `json:"relation_type"`
	ShowLabel           bool             `json:"show_label"`
	Order               int32            `json:"order"`
	IsEditable          bool             `json:"is_editable"`
	IsVisibleLayout     bool             `json:"is_visible_layout"`
}

type Relation struct {
	Id         string   `json:"id"`
	TableFrom  string   `json:"table_from"`
	TableTo    string   `json:"table_to"`
	Type       string   `json:"type"`
	FieldFrom  string   `json:"field_from"`
	ViewFields []string `json:"view_fields"`
}

type View struct {
	Id         string                 `json:"id"`
	Attributes map[string]interface{} `json:"attributes"`
	TableSlug  string                 `json:"table_slug"`
	Type       string                 `json:"type"`
	Columns    []string               `json:"columns"`
}

type ViewPermission struct {
	Guid   string `json:"guid"`
	RoleId string `json:"role_id"`
	ViewId string `json:"view_id"`
	View   bool   `json:"view"`
	Edit   bool   `json:"edit"`
	Delete bool   `json:"delete"`
}

type Table struct {
	Id              string `json:"id"`
	Slug            string `json:"slug"`
	Label           string `json:"label"`
	IsLoginTable    bool   `json:"is_login_table"`
	FromAuthService bool   `json:"from_auth_service"`
	SoftDelete      bool   `json:"soft_delete"`
}

type FieldPermission struct {
	Guid           string `json:"guid"`
	RoleId         string `json:"role_id"`
	Label          string `json:"label"`
	TableSlug      string `json:"table_slug"`
	FieldId        string `json:"field_id"`
	EditPermission bool   `json:"edit_permission"`
	ViewPermission bool   `json:"view_permission"`
}
