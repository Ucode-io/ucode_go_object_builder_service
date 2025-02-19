package postgres

import (
	"context"
	"encoding/json"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
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

func (l *languageRepo) Create(ctx context.Context, req *nb.CreateLanguageRequest) (*nb.Language, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.Create")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.GetProjectId())

	query := `INSERT INTO language (id, key, translations, category, platform) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id, key, translations, category, platform`
	var translations pgtype.JSONB

	jsonData, err := json.Marshal(req.GetTranslations())
	if err != nil {
		return nil, err
	}
	translations.Set(jsonData)

	var lang nb.Language
	var storedTranslations pgtype.JSONB

	err = conn.QueryRow(
		ctx,
		query,
		uuid.NewString(),
		req.GetKey(),
		translations,
		req.Category,
		req.Platform,
	).Scan(&lang.Id, &lang.Key, &storedTranslations, &lang.Category, &lang.Platform)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(storedTranslations.Bytes, &lang.Translations); err != nil {
		return nil, err
	}

	return &lang, nil
}

func (l *languageRepo) GetById(ctx context.Context, req *nb.PrimaryKey) (*nb.Language, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.GetByID")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.ProjectId)

	query := `
		SELECT 
			id, 
			key, 
			translations,
			category,
			platform
		FROM language WHERE id = $1`
	var lang nb.Language
	var translations pgtype.JSONB

	err := conn.QueryRow(
		ctx,
		query,
		req.Id,
	).Scan(&lang.Id, &lang.Key, &translations, &lang.Category, &lang.Platform)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(translations.Bytes, &lang.Translations); err != nil {
		return nil, err
	}

	return &lang, nil
}

func (l *languageRepo) GetList(ctx context.Context, req *nb.GetListLanguagesRequest) (resp *nb.GetListLanguagesResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.GetList")
	defer dbSpan.Finish()

	resp = &nb.GetListLanguagesResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	query := `
		SELECT 
			COUNT(*) OVER() as count, 
			id, 
			key, 
			translations,
			category,
			platform
		FROM language
		WHERE platform = $1
		ORDER BY category
	`

	var args []interface{}
	args = append(args, req.GetSearch())

	if req.GetLimit() > 0 {
		query += " LIMIT $2 OFFSET $3"
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

		err := rows.Scan(
			&resp.Count,
			&lang.Id,
			&lang.Key,
			&translations,
			&lang.Category,
			&lang.Platform,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(translations.Bytes, &lang.Translations); err != nil {
			return nil, err
		}

		languages = append(languages, &lang)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &nb.GetListLanguagesResponse{Languages: languages}, nil
}

func (l *languageRepo) UpdateLanguage(ctx context.Context, req *nb.UpdateLanguageRequest) (resp *nb.Language, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.UpdateLanguage")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.GetProjectId())

	query := `
        UPDATE language
        SET 
			key = $1, 
			translations = $2, 
			category = $4,
			platform = $5,
			updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, key, translations, category, platform
    `

	var lang nb.Language
	var translations pgtype.JSONB

	translationsBytes, err := json.Marshal(req.GetTranslations())
	if err != nil {
		return nil, err
	}
	translations.Set(translationsBytes)

	err = conn.QueryRow(
		ctx,
		query,
		req.GetKey(),
		translations,
		req.GetId(),
		req.Category,
		req.Platform,
	).Scan(
		&lang.Id, &lang.Key, &translations, &lang.Category, &lang.Platform,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(translations.Bytes, &lang.Translations); err != nil {
		return nil, err
	}

	return &lang, nil
}

func (l *languageRepo) Delete(ctx context.Context, req *nb.PrimaryKey) error {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "language.Delete")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.ProjectId)

	query := `DELETE FROM language WHERE id = $1`
	_, err := conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}
