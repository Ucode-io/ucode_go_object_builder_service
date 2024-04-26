package helper

import "encoding/json"

var VIEW_TYPES = map[string]string{
	"TABLE":            "TABLE",
	"CALENDAR":         "CALENDAR",
	"CALENDAR HOUR":    "CALENDAR HOUR",
	"GANTT":            "GANTT",
	"TREE":             "TREE",
	"BOARD":            "BOARD",
	"FINANCE CALENDAR": "FINANCE CALENDAR",
}

var (
	FIELD_TYPES = map[string]string{
		"SINGLE_LINE":                 "VARCHAR",
		"MULTI_LINE":                  "VARCHAR",
		"PICK_LIST":                   "VARCHAR",
		"LOOKUP":                      "VARCHAR",
		"EMAIL":                       "VARCHAR",
		"PHOTO":                       "VARCHAR",
		"PHONE":                       "VARCHAR",
		"UUID":                        "VARCHAR",
		"INCREMENT_ID":                "VARCHAR",
		"RANDOM_NUMBERS":              "VARCHAR",
		"PASSWORD":                    "VARCHAR",
		"FILE":                        "VARCHAR",
		"CODABAR":                     "VARCHAR",
		"INTERNATIONAL_PHONE":         "VARCHAR",
		"FORMULA_FRONTEND":            "VARCHAR",
		"DATE":                        "DATE",
		"TIME":                        "TIME",
		"DATE_TIME":                   "TIMESTAMP",
		"DATE_TIME_WITHOUT_TIME_ZONE": "TIMESTAMP",
		"NUMBER":                      "FLOAT",
		"MONEY":                       "FLOAT",
		"FLOAT":                       "FLOAT",
		"FLOAT_NOLIMIT":               "FLOAT",
		"FORMULA":                     "FLOAT",
		"CHECKBOX":                    "BOOL",
		"SWITCH":                      "BOOL",
		"MULTISELECT":                 "TEXT[]",
		"LOOKUPS":                     "TEXT[]",
		"DYNAMIC":                     "TEXT[]",
		"LANGUAGE_TYPE":               "TEXT[]",
		"MULTI_IMAGE":                 "TEXT[]",
	}
)

var (
	TYPE_REGEXP = map[string]string{
		"VARCHAR":   "^.{0,255}$",
		"DATE":      `^\d{4}-\d{2}-\d{2}$`,
		"TIME":      `^\d{2}:\d{2}(:\d{2})?$`,
		"TIMESTAMP": `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?$`,
		"FLOAT":     `^-?\d+(\.\d+)?$`,
		"BOOL":      "^(true|false)$",
		"TEXT[]":    "^{.*}$",
	}
)

var (
	TYPE_DEFAULT = map[string]interface{}{
		"VARCHAR":   "",
		"DATE":      "CURRENT_DATE",
		"TIME":      "CURRENT_TIME",
		"TIMESTAMP": "CURRENT_TIMESTAMP",
		"FLOAT":     0.0,
		"BOOL":      false,
		"TEXT[]":    "{}",
	}
)

func GetDataType(t string) string {

	val, ok := FIELD_TYPES[t]
	if !ok {
		return "VARCHAR"
	}

	return val
}

func GetRegExp(t string) string {

	val, ok := TYPE_REGEXP[t]
	if !ok {
		return "^.{0,255}$"
	}

	return val
}

func GetDefault(t string) interface{} {

	val, ok := TYPE_DEFAULT[t]
	if !ok {
		return ""
	}

	return val
}

func MarshalToStruct(data interface{}, resp interface{}) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, resp)
	if err != nil {
		return err
	}

	return nil
}
