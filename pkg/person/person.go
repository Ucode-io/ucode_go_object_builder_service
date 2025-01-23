package person

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/config"

	"github.com/jackc/pgx/v5"
)

type CreateSyncWithLoginTableRequest struct {
	Ctx                context.Context
	Data               map[string]any
	Tx                 pgx.Tx
	Guid               string
	UserIdAuth         string
	LoginTableSlug     string
	TableAttributesMap map[string]any
	Update             bool
}

func CreateSyncWithLoginTable(req CreateSyncWithLoginTableRequest) error {
	var (
		authInfo                     = cast.ToStringMap(req.TableAttributesMap["auth_info"])
		clientTypeID, clientTypeIDOk = authInfo[config.CLIENT_TYPE_ID].(string)
		roleID, roleIDOk             = authInfo[config.ROLE_ID].(string)
		phone, phoneOk               = authInfo["phone"].(string)
		email, emailOk               = authInfo["email"].(string)
		login, loginOk               = authInfo["login"].(string)
		password, passwordOk         = authInfo["password"].(string)

		queryFields = []string{"guid, user_id_auth"}
		queryValues = []string{"$1, $2"}
		args        = []any{req.Guid, req.UserIdAuth}
		argCount    = 3
	)

	if clientTypeIDOk {
		queryFields = append(queryFields, clientTypeID)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data[config.CLIENT_TYPE_ID])
		argCount++
	}

	if roleIDOk {
		queryFields = append(queryFields, roleID)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data[config.ROLE_ID])
		argCount++
	}

	if phoneOk {
		queryFields = append(queryFields, phone)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data["phone_number"])
		argCount++
	}

	if emailOk {
		queryFields = append(queryFields, email)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data["email"])
		argCount++
	}

	if loginOk {
		queryFields = append(queryFields, login)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data["login"])
		argCount++
	}

	if passwordOk {
		queryFields = append(queryFields, password)
		queryValues = append(queryValues, fmt.Sprintf("$%d", argCount))
		args = append(args, req.Data["password"])
		argCount++
	}

	query := fmt.Sprintf(`
			INSERT INTO "%s" (%s) VALUES (%s)`,
		req.LoginTableSlug,
		strings.Join(queryFields, ", "),
		strings.Join(queryValues, ", "),
	)

	_, err := req.Tx.Exec(req.Ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "when insert to login table")
	}

	return nil
}
