package db

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

type DB struct {
	db.Session

	host      string
	port      int
	user, pwd string
}

func New(host string, port int, user, pwd string) (*DB, error) {
	settings, err := mysql.ParseURL(fmt.Sprintf("%s:%s@tcp(%s:%d)/atec2022?maxAllowedPacket=1073741824&multiStatements=true", user, pwd, host, port))
	if err != nil {
		return nil, err
	}
	sess, err := mysql.Open(settings)
	if err != nil {
		return nil, err
	}
	sess.SetMaxIdleConns(1200)
	sess.SetMaxOpenConns(1200)
	db.LC().SetLevel(db.LogLevelPanic)
	return &DB{
		Session: sess,
		host:    host,
		port:    port,
		user:    user,
		pwd:     pwd,
	}, nil
}

func (d *DB) ShowVariables() error {
	rows, err := d.SQL().Query("show variables;")
	if err != nil {
		return err
	}
	defer rows.Close()
	a, b := "", ""
	for rows.Next() {
		err = rows.Scan(&a, &b)
		if err != nil {
			panic(err)
			return err
		}
		fmt.Println(a, b)
	}
	return nil
}
