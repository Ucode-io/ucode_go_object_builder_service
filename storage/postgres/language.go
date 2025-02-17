package postgres

import (
	"context"
	"encoding/json"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgtype"
	"github.com/opentracing/opentracing-go"
)

type languageRepo struct {
	db *psqlpool.Pool
}

func NewLanguageRepo(db *psqlpool.Pool) storage.LanguageRepoI {
	return &languageRepo{
		db: db,
	}
}

func (l *languageRepo) GetList(ctx context.Context, req *nb.GetListLanguagesRequest) (resp *nb.GetListLanguagesResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.GetList")
	defer dbSpan.Finish()

	resp = &nb.GetListLanguagesResponse{}

	conn := psqlpool.Get(req.GetProjectId())
	query := `SELECT COUNT(*) OVER() as count, id, key, translations FROM language`
	var args []interface{}

	if req.GetLimit() > 0 {
		query += ` LIMIT $1 OFFSET $2`
		args = append(args, req.GetLimit(), req.GetOffset())
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var languages []*nb.Language
	for rows.Next() {
		var lang nb.Language
		var translations pgtype.JSONB

		err := rows.Scan(&resp.Count, &lang.Id, &lang.Key, &translations)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(translations.Bytes, &lang.Translations); err != nil {
			return nil, err
		}

		languages = append(languages, &lang)
	}

	return &nb.GetListLanguagesResponse{Languages: languages}, nil
}

func (l *languageRepo) UpdateLanguage(ctx context.Context, req *nb.UpdateLanguageRequest) (resp *nb.Language, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.UpdateLanguage")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.GetProjectId())

	query := `
        UPDATE language
        SET key = $1, translations = $2, updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, key, translations
    `

	var lang nb.Language
	var translations pgtype.JSONB

	translationsBytes, err := json.Marshal(req.GetTranslations())
	if err != nil {
		return nil, err
	}
	translations.Set(translationsBytes)

	err = conn.QueryRow(ctx, query, req.GetKey(), translations, req.GetId()).Scan(
		&lang.Id, &lang.Key, &translations,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(translations.Bytes, &lang.Translations); err != nil {
		return nil, err
	}

	return &lang, nil
}
