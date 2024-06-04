package psqlpool

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var PsqlPool = make(map[string]*pgxpool.Pool) // there we save psql connections by project_id

func Add(projectId string, conn *pgxpool.Pool) {
	if projectId == "" {
		fmt.Println("WARNING!!! projectId is empty")
		return
	}

	_, ok := PsqlPool[projectId]
	if ok {
		fmt.Println("db conn with given projectId already exists")
		return
	}

	PsqlPool[projectId] = conn
}

func Get(projectId string) (conn *pgxpool.Pool) {
	fmt.Println("here project id >>>>> ", projectId)
	if projectId == "" {
		fmt.Println("WARNING!!! projectId is empty")
		return nil
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		fmt.Println("db conn with given projectId does not exists")
		return nil
	}

	return PsqlPool[projectId]
}

func Remove(projectId string) {
	if projectId == "" {
		fmt.Println("WARNING!!! projectId is empty")
		return
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		fmt.Println("db conn with given projectId does not exists")
		return
	}

	delete(PsqlPool, projectId)
}

func Override(projectId string, conn *pgxpool.Pool) {
	if projectId == "" {
		fmt.Println("WARNING!!! projectId is empty")
		return
	}

	_, ok := PsqlPool[projectId]
	if !ok {
		fmt.Println("db conn with given projectId does not exists")
		return
	}

	PsqlPool[projectId] = conn
}
