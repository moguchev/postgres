package main

import (
	// https://go.dev/src/database/sql/doc.txt

	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // импортируем драйвер для postgres
)

const (
	// название регистрируемоего драйвера github.com/lib/pq
	stdPostgresDriverName = "postgres"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "playground"
)

func clear(c io.Closer) {
	if err := c.Close(); err != nil {
		fmt.Println("close:", err)
	}
}

func main() {
	// connection string
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	sqlDB, err := sql.Open(stdPostgresDriverName, psqlConn) // returns *sql.DB, error
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	// Можем обернуть существующий *sql.DB в *sqlx.DB; нужно обоязательно передать тот же драйвер, иначе не заведется
	sqlxDB := sqlx.NewDb(sqlDB, stdPostgresDriverName)
	clear(sqlxDB) // дальше не использую

	// А можем сразу создавать *sqlx.DB с помощью функции sqlx.Connect
	// sqlx.Connect = sql.Open + sql.Ping (2 в одном :))
	db, err := sqlx.Connect(stdPostgresDriverName, psqlConn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close() // так же не забываем освободить ресурсы

	fmt.Println("Connection with database successfully established!")

	// А если мы ленивые и не хотим обрабатывать ошибку, хотим сразу падать
	sqlxDB = sqlx.MustConnect(stdPostgresDriverName, psqlConn)
	clear(sqlxDB) // дальше не использую

	/*
		sqlx реализует стандартные функции: database/sql (ведут себя точно так же как стандартные, нет никаких преимуществ):
			* Exec(...) (sql.Result, error) - unchanged from database/sql
			* Query(...) (*sql.Rows, error) - unchanged from database/sql
			* QueryRow(...) *sql.Row - unchanged from database/sql

		Расширенные стандартные функции (позволяют использовать фичи sqlx):
			* MustExec() sql.Result -- Exec, but panic on error // не надо использовать в проде:)
			* Queryx(...) (*sqlx.Rows, error) - Query, but return an sqlx.Rows
			* QueryRowx(...) *sqlx.Row -- QueryRow, but return an sqlx.Row

		Новые:
			* Get(dest interface{}, ...) error
			* Select(dest interface{}, ...) error
			* NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	*/
	ctx := context.Background()

	exampleQueryx(ctx, db)

	exampleGet(ctx, db)
	exampleSelect(ctx, db)
	exampleNamedQuery(ctx, db)

	// See more: https://jmoiron.github.io/sqlx/
}

type Student struct {
	FirstName  string `db:"first_name"`
	LastName   string `db:"last_name"`
	Age        uint   `db:"age"`
	OtherField string `db:"-"`
}

func exampleQueryx(ctx context.Context, db *sqlx.DB) {
	const minAge = 18

	rows, err := db.QueryxContext(ctx, "SELECT first_name, last_name, age FROM students WHERE age >= $1", minAge)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close() // Обязательно закрываем иначе соединение с БД повиснет

	var students []Student
	for rows.Next() {
		var st Student

		// Варианта 1
		if err := rows.StructScan(&st); err != nil {
			log.Fatal(err)
		}

		students = append(students, st)
	}

	if err = rows.Err(); err != nil {
		// handle the error here
		log.Fatal(err)
	}

	fmt.Printf("students: %v\n", students)
}

func exampleGet(ctx context.Context, db *sqlx.DB) {
	var st Student
	// Get записывает в st данные из первой строки
	if err := db.GetContext(ctx, &st, "SELECT first_name, last_name,age FROM students LIMIT 1"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Printf("no students")
		} else {
			log.Fatal(err) // handle the error here
		}
	}

	fmt.Printf("student: %v\n", st)

	// sqlx.DB.Get ~ sql.DB.QueryRow
}

func exampleSelect(ctx context.Context, db *sqlx.DB) {
	const (
		minAge = 18
		query  = "SELECT first_name, last_name, age FROM students WHERE age >= $1"
	)

	var students []Student
	// Select записывает в students массив полученных строк.
	if err := db.SelectContext(ctx, &students, query, minAge); err != nil {
		log.Fatal(err) // handle the error here
	}

	fmt.Printf("students: %v\n", students)
	// sqlx.DB.Select ~ sql.DB.Query
}

func exampleNamedQuery(ctx context.Context, db *sqlx.DB) {
	st := Student{
		FirstName: "Bob",
		LastName:  "Brown",
	}

	const query = `
		SELECT 
			first_name, 
			last_name, 
			age 
		FROM students 
		WHERE first_name=:first_name 
		   OR last_name=:last_name
	`
	rows, err := db.NamedQueryContext(ctx, query, st)
	if err != nil {
		log.Fatal(err) // handle the error here
	}
	defer rows.Close()

	var students []Student
	for rows.Next() {
		var st Student
		if err := rows.StructScan(&st); err != nil {
			log.Fatal(err) // handle the error here
		}

		students = append(students, st)
	}

	if err = rows.Err(); err != nil {
		log.Fatal(err) // handle the error here
	}

	fmt.Printf("students: %v\n", students)
}
