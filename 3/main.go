package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/moguchev/postgres/3/repository"
	students_databasesql "github.com/moguchev/postgres/3/repository/students/database_sql_implementation"
	students_pgx "github.com/moguchev/postgres/3/repository/students/pgx_implementation"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "playground"
)

type StudentUsecase struct {
	repo repository.StudentsRepository
}

func main() {
	ctx := context.Background()

	// connection string
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database
	db, err := sql.Open("postgres", psqlConn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pool, err := pgxpool.Connect(ctx, psqlConn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal(err)
	}

	su := &StudentUsecase{} // наша бизнес логика

	// нашей бизнес логике всеравно что мы используем
	// мы можем спокойно подменять реализации(мигрировать с одной на другую без особых изменений кода)
	su.repo = students_databasesql.NewRepository(db)
	su.repo = students_pgx.NewRepository(pool)
}
