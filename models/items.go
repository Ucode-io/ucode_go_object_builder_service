package models

type IsExistsBody struct {
	TableSlug  string
	FieldSlug  string
	FieldValue any
}

type CreateBody struct {
	FieldMap   map[string]FieldBody
	TableId    string
	Fields     []Field
	TableSlugs []string
}
