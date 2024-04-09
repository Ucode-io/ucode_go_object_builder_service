package helper

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

func GetDataType(t string) string {

	val, ok := FIELD_TYPES[t]
	if !ok {
		return "VARCHAR"
	}

	return val
}
