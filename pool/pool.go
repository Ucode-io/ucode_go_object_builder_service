package psqlpool

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/opentracing/opentracing-go"
)

var PsqlPool = make(map[string]*Pool) // there we save psql connections by project_id

type Pool struct {
	Db *pgxpool.Pool
}

func (b *Pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "pgx.QueryRow")
	defer dbSpan.Finish()

	dbSpan.SetTag("sql", sql)
	dbSpan.SetTag("args", args)

	return b.Db.QueryRow(ctx, sql, args...)
}

func (b *Pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "pgx.Query")
	defer dbSpan.Finish()

	dbSpan.SetTag("sql", sql)
	dbSpan.SetTag("args", args)

	return b.Db.Query(ctx, sql, args...)
}

func (b *Pool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "pgx.Exec")
	defer dbSpan.Finish()

	dbSpan.SetTag("sql", sql)
	dbSpan.SetTag("args", arguments)

	return b.Db.Exec(ctx, sql, arguments...)
}

func (b *Pool) Begin(ctx context.Context) (pgx.Tx, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "pgx.Begin")
	defer dbSpan.Finish()

	tx, err := b.Db.Begin(ctx)
	if err != nil {
		dbSpan.SetTag("error", true)
		dbSpan.LogKV("error.message", err.Error())
		return nil, err
	}

	return tx, nil
}

func Add(projectId string, conn *Pool) {
	if projectId == "" {
		return
	}

	_, ok := PsqlPool[projectId]
	if ok {
		return
	}

	PsqlPool[projectId] = conn
}

func Get(projectId string) (conn *Pool) {
	if projectId == "" {
		return nil
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		return nil
	}

	return PsqlPool[projectId]
}

func Remove(projectId string) {
	if projectId == "" {
		return
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		return
	}

	delete(PsqlPool, projectId)
}

func Override(projectId string, conn *Pool) {
	if projectId == "" {
		return
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		return
	}

	PsqlPool[projectId] = conn
}
