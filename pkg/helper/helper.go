package helper

import (
	"encoding/json"
	"os"
)

// func ReplaceQueryParams(namedQuery string, params map[string]interface{}) (string, []interface{}) {
// 	var (
// 		i    int = 1
// 		args []interface{}
// 	)

// 	for k, v := range params {
// 		if k != "" {
// 			oldsize := len(namedQuery)
// 			namedQuery = strings.ReplaceAll(namedQuery, ":"+k, "$"+strconv.Itoa(i))

// 			if oldsize != len(namedQuery) {
// 				args = append(args, v)
// 				i++
// 			}
// 		}
// 	}

// 	return namedQuery, args
// }

func ChangeHostname(data []byte) ([]byte, error) {

	var (
		isChangedByHost = map[string]bool{}
	)

	if err := json.Unmarshal(data, &isChangedByHost); err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	isChangedByHost[hostname] = true

	data, err = json.Marshal(isChangedByHost)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// func TableVersion(conn pgxpool.Pool)
