package models

import "google.golang.org/protobuf/types/known/structpb"

type TablePermission struct {
	Id         string
	Slug       string
	Label      string
	ShowInMenu bool
	IsChanged  bool
	IsCached   bool
	Icon       string
	IsSystem   bool
	Attributes *structpb.Struct
}

type RecordPermission struct {
	Read            string
	Write           string
	Update          string
	Delete          string
	IsHaveCondition bool
	IsPublic        bool
	RoleID          string
	TableSlug       string
	LanguageBtn     string
	Automation      string
	Settings        string
	ShareModal      string
	ViewCreate      string
	PDFAction       string
	AddField        string
	DeleteAll       string
}
