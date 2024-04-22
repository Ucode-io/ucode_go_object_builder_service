package models

type ReqForViewRelation struct {
	Id        string
	ProjectId string
	TableSlug string
	TableId   string
	RoleId    string
}

type RelationForView struct {
	Id                     string
	TableFrom              string
	TableTo                string
	FieldFrom              string
	FieldTo                string
	Type                   string
	ViewFields             []string
	Editable               bool
	RelationFieldSlug      string
	AutoFilters            []string
	IsUserIdDefault        bool
	Cascadings             []string
	ObjectIdFromJwt        string
	CascadingTreeTableSlug string
	CascadingTreeFieldSlug string
	CreatedAt              string
	UpdatedAt              string
	DynamicTables          []struct {
		TableSlug  string
		ViewFields []string
	}
}
