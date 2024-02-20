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

// SQLite connection.
// Create new instance with NewSqlite().
type SqliteCon struct {
	db   *sql.DB
	errs *log.Logger
}

// Creates new instance of SQLite connection
func NewSqlite() DBConnection {
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

// Implementation of DBConnection.SetPoints
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

// Implementation of DBConnection.AddPoints
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

// Implementation of DBConnection.TryRemovePoints
func (db *SqliteCon) TryRemovePoints(username string, points int) bool {
	username = strings.ToLower(username)
	if db.GetPoints(username) < points {
		return false
	}
	db.AddPoints(username, -points)
	return true
}

// Implementation of DBConnection.GetPoints
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

// Implementation of DBConnection.GetPointsTop20
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

// Implementation of DBConnection.GetPointsTop5
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

// Implementation of DBConnection.MassAddPoints
func (db *SqliteCon) MassAddPoints(usernames []string, toAdd int) {
	tx, err := db.db.BeginTx(context.Background(), nil)
	if err != nil {
		db.errs.Println(err)
		return
	}

	for _, username := range usernames {
		username = strings.ToLower(username)

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
