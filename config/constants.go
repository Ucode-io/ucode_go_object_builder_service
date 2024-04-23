package config

import "time"

const (
	DatabaseQueryTimeLayout string = `'YYYY-MM-DD"T"HH24:MI:SS"."MS"Z"TZ'`
	DatabaseTimeLayout      string = time.RFC3339
	ErrTheSameId                   = "cannot use the same uuid for 'id' and 'parent_id' fields"
	ErrRpcNodFoundAndNoRows        = "rpc error: code = NotFound desc = no rows in result set"
	ErrNoRows                      = "no rows in result set"
	ErrObjectType                  = "object type error: code =  NodFound"
	ErrEnvNodFound                 = "No .env file found"
)

var (
	MENU_TYPES      = []string{"TABLE", "LAYOUT", "FOLDER", "MICROFRONTEND", "FAVOURITE", "HIDE", "WEBPAGE", "PIVOT", "REPORT_SETTING", "LINK", "MINIO_FOLDER", "WIKI", "WIKI_FOLDER"}
	STATIC_MENU_IDS = []string{
		"c57eedc3-a954-4262-a0af-376c65b5a284", //root
		"c57eedc3-a954-4262-a0af-376c65b5a282", //favorite
		"c57eedc3-a954-4262-a0af-376c65b5a280", //admin
		"c57eedc3-a954-4262-a0af-376c65b5a278", //analytics
		"c57eedc3-a954-4262-a0af-376c65b5a276", //pivot
		"c57eedc3-a954-4262-a0af-376c65b5a274", //report setting
		"7c26b15e-2360-4f17-8539-449c8829003f", //saved pivot
		"e96b654a-1692-43ed-89a8-de4d2357d891", //history pivot
		"a8de4296-c8c3-48d6-bef0-ee17057733d6", //admin => user and permission
		"d1b3b349-4200-4ba9-8d06-70299795d5e6", //admin => database
		"f7d1fa7d-b857-4a24-a18c-402345f65df8", //admin => code
		"f313614f-f018-4ddc-a0ce-10a1f5716401", //admin => resource
		"db4ffda3-7696-4f56-9f1f-be128d82ae68", //admin => api
		"3b74ee68-26e3-48c8-bc95-257ca7d6aa5c", // profile setting
		"8a6f913a-e3d4-4b73-9fc0-c942f343d0b9", //files menu id
		"744d63e6-0ab7-4f16-a588-d9129cf959d1", //wiki menu id
		"9e988322-cffd-484c-9ed6-460d8701551b", // users menu id
	}
	STRING_TYPES = []string{
		"SINGLE_LINE", "MULTI_LINE",
		"PICK_LIST", "DATE",
		"LOOKUP", "EMAIL",
		"PHOTO", "PHONE", "UUID", "DATE_TIME",
		"TIME", "INCREMENT_ID", "RANDOM_NUMBERS", "PASSWORD",
		"FILE", "CODABAR", "INTERNATIONAL_PHONE", "DATE_TIME_WITHOUT_TIME_ZONE",
	}
)

const (
	MANY2DYNAMIC = "Many2Dynamic"
	MANY2MANY    = "Many2Many"
	RECURSICE    = "Recursive"
	MANY2ONE     = "Many2One"
	ONE2ONE      = "One2One"
)
