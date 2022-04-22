package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// MariaDB
var DB_USER string = os.Getenv("DB_USER")
var DB_PASSWORD string = os.Getenv("DB_PASSWORD")
var DB_NAME string = os.Getenv("DB_NAME")
var DB_URL string = os.Getenv("DB_URL")

var DB_ADMIN string = os.Getenv("DB_ADMIN")
var DB_ADMIN_PASSWORD string = os.Getenv("DB_ADMIN_PASSWORD")

// Postgres
var Postgreshost string = os.Getenv("Postgreshost")
var Postgresport int = 5432
var Postgresuser string = os.Getenv("Postgresuser")
var Postgrespassword string = os.Getenv("Postgrespassword")
var Postgresdbname string = os.Getenv("Postgresdbname")

var path string = "/usr/local/bin/sql/"

func SetUpMariaDB_admin() *sql.DB {
	MariaDB, err := sql.Open("mysql", DB_ADMIN+":"+DB_ADMIN_PASSWORD+"@tcp("+DB_URL+":3306"+")/"+DB_NAME)
	if err != nil {
		panic(err.Error())
	}
	return MariaDB
}
func SetUpMariaDB() *sql.DB {
	MariaDB, err := sql.Open("mysql", DB_USER+":"+DB_PASSWORD+"@tcp("+DB_URL+":3306"+")/"+DB_NAME)
	if err != nil {
		panic(err.Error())
	}
	return MariaDB
}

func SetUpPostgres() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", Postgreshost, Postgresport, Postgresuser, Postgrespassword, Postgresdbname)

	PostgresDB, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	return PostgresDB
}

func ExecuteFile(database *sql.DB, filename string) {
	path := filepath.Join(path, filename)

	c, ioErr := ioutil.ReadFile(path)
	if ioErr != nil {
		// handle error.
		log.Fatal(ioErr)
	}
	sql := string(c)
	_, err := database.Exec(sql)

	if err != nil {
		log.Fatal(err)
	}

}
