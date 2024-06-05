package psqlpool

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

var PsqlPool = make(map[string]*pgxpool.Pool) // there we save psql connections by project_id

func Add(projectId string, conn *pgxpool.Pool) {
	if projectId == "" {
		return
	}

	_, ok := PsqlPool[projectId]
	if ok {
		return
	}

	PsqlPool[projectId] = conn
}

func Get(projectId string) (conn *pgxpool.Pool) {
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

func Override(projectId string, conn *pgxpool.Pool) {
	if projectId == "" {
		return
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		return
	}

	PsqlPool[projectId] = conn
}
