package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteCon struct {
	db   *sql.DB
	errs *log.Logger
}

func New() DBConnection {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	dbFile := path.Join(pwd, "viewers.db")
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(createTableSql)
	if err != nil {
		panic(err)
	}

	fmt.Println("Database ready")

	return &SqliteCon{
		db:   db,
		errs: log.New(os.Stdout, "[DB ERR] ", log.LUTC),
	}
}

func (db *SqliteCon) SetPoints(username string, points int) {
	username = strings.ToLower(username)
	// can't find user
	if db.GetPoints(username) == -1 {
		_, err := db.db.Exec(insertPointsViewerSql, points)
		if err != nil {
			db.errs.Println(err)
		}
		return
	}
	_, err := db.db.Exec(setPointsViewerSql, points)
	if err != nil {
		db.errs.Println(err)
	}
}

func (db *SqliteCon) AddPoints(username string, points int) {
	username = strings.ToLower(username)
	// can't find user
	if db.GetPoints(username) == -1 {
		_, err := db.db.Exec(insertPointsViewerSql, points)
		if err != nil {
			db.errs.Println(err)
		}
		return
	}
	_, err := db.db.Exec(addPointsViewerSql, points)
	if err != nil {
		db.errs.Println(err)
	}
}

func (db *SqliteCon) TryRemovePoints(username string, points int) bool {
	username = strings.ToLower(username)
	if db.GetPoints(username) < points {
		return false
	}
	db.AddPoints(username, -points)
	return true
}

func (db *SqliteCon) GetPoints(username string) int {
	username = strings.ToLower(username)
	var points int

	row := db.db.QueryRow(getPointsViewerSql, username)
	if err := row.Scan(&points); err != nil {
		db.errs.Println(err)
		return -1
	}
	return points
}

func (db *SqliteCon) GetPointsTop20() map[string]int {
	res := map[string]int{}

	var name string
	var points int

	rows, err := db.db.Query(getPointsViewerMax20Sql)
	if err != nil {
		db.errs.Println(err)
		return nil
	}

	for i := 0; rows.Next() && i < 20; i++ {
		err = rows.Scan(&name, &points)
		if err != nil {
			db.errs.Println(err)
			continue
		}
		res[name] = points
	}

	return res
}

func (db *SqliteCon) GetPointsTop5(owner string) map[string]int {
	owner = strings.ToLower(owner)

	res := map[string]int{}

	var name string
	var points int

	rows, err := db.db.Query(getPointsTopSql, owner)
	if err != nil {
		db.errs.Println(err)
		return nil
	}
	for i := 0; rows.Next() && i < 5; i++ {
		err = rows.Scan(&name, &points)
		if err != nil {
			db.errs.Println(err)
			continue
		}
		res[name] = points
	}

	return res
}

func (db *SqliteCon) MassAddPoints(usernames []string, toAdd int) {
	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		db.errs.Println(err)
		return
	}

	for _, username := range usernames {
		username = strings.ToLower(username)
		// TODO:mb should change
		if db.GetPoints(username) == -1 {
			_, err := tx.Exec(insertPointsViewerSql, username, toAdd)
			if err != nil {
				db.errs.Println(err)
			}
			return
		}
		_, err := tx.Exec(addPointsViewerSql, username, toAdd)
		if err != nil {
			db.errs.Println(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		db.errs.Println(err)
		return
	}
}
