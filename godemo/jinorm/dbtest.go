package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main()  {
	db , _ := sql.Open("sqlite3", "test.db")
	defer func() {
		_=db.Close()
	}()
	_,_=db.Exec("DROP TABLE IF EXISTS User;")
	_,_=db.Exec("CREATE TABLE User(Name, text);")
	result, err:= db.Exec("INSERT INTO User(`Name`) values (?),(?)","toms", "Sam")
	if err == nil {
		affected, _ := result.RowsAffected()
		log.Println(affected)
	}
	row := db.QueryRow("SELECT  Name FROM User LIMIT 1")
	var name string
	if err:= row.Scan(&name); err == nil {
		log.Println(name)
	}

}
