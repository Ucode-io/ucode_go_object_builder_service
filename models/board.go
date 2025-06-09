package models

type BoardDataParams struct {
	GroupBy    GroupBy    `json:"group_by"`
	SubgroupBy SubgroupBy `json:"subgroup_by"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
	Fields     []string   `json:"fields"`
}

type GroupBy struct {
	Field string `json:"field"`
}

type SubgroupBy struct {
	Field string `json:"field"`
}

type BoardGroup struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type BoardSubgroup struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
