package main

import (
	"database/sql" // https://go.dev/src/database/sql/doc.txt
	"fmt"
	"log"

	_ "github.com/lib/pq" // импортируем драйвер для postgres
	// Обратите внимание, что мы загружаем драйвер анонимно, присвоив его квалификатору пакета псевдоним, _ ,
	// чтобы ни одно из его экспортированных имен не было видно нашему коду.
	// Под капотом драйвер регистрирует себя как доступный для пакета database/sql
)

const (
	// название регистрируемоего драйвера github.com/lib/pq
	stdPostgresDriverName = "postgres"
	/*
		PostgreSQL:
			* github.com/lib/pq -> postgres
			* github.com/jackc/pgx -> pgx
		MySQL:
			* github.com/go-sql-driver/mysql -> mysql
		SQLite3:
			* github.com/mattn/go-sqlite3 -> sqlite3
		Oracle:
			* github.com/godror/godror -> godror
		MS SQL:
			* github.com/denisenkom/go-mssqldb -> sqlserver

		See more: https://zchee.github.io/golang-wiki/SQLDrivers/
	*/
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "playground"
)

func main() {
	// connection string
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open(stdPostgresDriverName, psqlConn) // *sql.DB, error
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close() // Обязательно при завершении работы приложения мы должны освободить все ресурсы, иначе соединения к базе останутся висеть.

	/*
		sql.DB - не является соединением с базой данных! Это абстракция интерфейса.

		sql.DB выполняет некоторые важные задачи для вас за кулисами:
		 	* открывает и закрывает соединения с фактической базовой базой данных через драйвер.
			* управляет пулом соединений по мере необходимости.

		Абстракция sql.DB предназначена для того, чтобы вы не беспокоились о том, как управлять одновременным
		доступом к базовому хранилищу данных. Соединение помечается как используемое, когда вы используете
		его для выполнения задачи, а затем возвращается в доступный пул, когда оно больше не используется.
	*/

	// После установления соединеия пингуем базу. Проверяем, что она отвечает нашему приложению.
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("connection with database successfully established")
}
