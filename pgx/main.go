// see: https://github.com/jackc/pgx
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/log/zapadapter" // adapter for go.uber.org/zap loggger
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap" // zap logger
	/*
		github.com/jackc/pgx — именно этот драйвер вы хотите использовать. Почему?

		- Активно поддерживается и развивается.
		- Может быть производительнее в случае использования без интерфейсов database/sql .
		- Поддержка более 60 типов PostgreSQL, которые PostgreSQL реализует вне стандарта SQL.
		- Возможность удобно реализовать логирование того, что происходит внутри драйвера.
		- У pgx человекопонятные ошибки, в то время как просто lib/pq бросает паники. Если не поймать панику, программа упадет. (Не стоит использовать паники в Go, это не то же самое, что исключения.)
		- С pgx у нас есть возможность независимо конфигурировать каждое соединение.
		- Есть поддержка протокола логической репликации PostgreSQL.

		Выбор между интерфейсами pgx и database/sql?
		Рекомендуется использовать интерфейс pgx, если:

		1) Приложение предназначено только для PostgreSQL.
		2) Никакие другие требуемые библиотеки database/sqlне используются.

		Интерфейс pgx быстрее и предоставляет больше возможностей.

		Интерфейс database/sql позволяет базовому драйверу возвращать или получать
		только следующие типы: int64, float64, bool, []byte, string, time.Time или nil.
		Для работы с другими типами требуется реализация интерфейсов
		database/sql.Scanner и database/sql/driver/driver.Valuer для передачи значений в текстовом формате.
		Двоичный формат может быть значительно быстрее, что и использует интерфейс pgx.
	*/)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "playground"
)

func main() {
	ctx, cacnel := context.WithCancel(context.Background())
	defer cacnel()

	zapLogger := zap.NewExample()
	defer zapLogger.Sync()
	logger := zapLogger.Sugar()

	// connection string
	psqlConn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// open database connection
	conn, err := pgx.Connect(ctx, psqlConn) // это именно одно соединение, не пул соединений
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer conn.Close(ctx)

	if err := conn.Ping(ctx); err != nil {
		logger.Fatal(err)
	}

	// *pgx.Conn.Exec(...) like *sql.DB.ExexContext(...)
	// *pgx.Conn.Query(...) like *sql.DB.QueryContext(...)
	// *pgx.Conn.QueryRow(...) like *sql.DB.QueryRowContext(...)
	// *pgx.Conn.BeginTx(...) like *sql.DB.BeginTx(...)

	// New:
	// *pgx.Conn.QueryFunc()
	// *pgx.Conn.SendBatch()

	fmt.Println("Connection with database successfully established!")

	// pgxpool - пул соединений с БД

	pool, err := pgxpool.Connect(ctx, psqlConn)
	if err != nil {
		logger.Fatal(err)
	}
	defer pool.Close()

	// настраиваем
	config := pool.Config()
	config.MaxConnIdleTime = time.Minute
	config.MaxConnLifetime = time.Hour
	config.MinConns = 2
	config.MaxConns = 10

	config.ConnConfig.Logger = zapadapter.NewLogger(zapLogger) // передаем наш zap логгер
	config.ConnConfig.LogLevel = pgx.LogLevelDebug             // уровень логирования выставляем

	exampleSelect(ctx, pool)

	// github.com/jackc/pgx/v4/stdlib - для совместимости с database/sql (теряем все преимущества pgx)

	// github.com/jackc/pgtype - Поддерживается более 70 типов PostgreSQL, включая uuid, hstore, json, bytea, numeric, interval

	// github.com/jackc/pgerrcode -  коды ошибок postgres

	// github.com/vgarvardt/pgx-google-uuid - поддержка github.com/google/uuid.
}

func exampleSelect(ctx context.Context, pool *pgxpool.Pool) {
	// github.com/georgysavva/scany - Библиотека для сканирования данных из базы данных в структуры  (like sqlx)

	type Student struct {
		FirstName  string `db:"first_name"`
		LastName   string `db:"last_name"`
		Age        uint   `db:"age"`
		OtherField string `db:"-"`
	}
	const query = `SELECT first_name, last_name, age FROM students`

	// Ex. 1
	var students []Student
	if err := pgxscan.Select(ctx, pool, &students, query); err != nil {
		log.Fatal(err)
	}
	log.Println(students)

	// Ex. 2
	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Ex. 2.1
	students = students[:0]
	if err := pgxscan.ScanAll(&students, rows); err != nil {
		log.Fatal(err)
	}
	log.Println(students)

	// Ex. 2.2
	students = students[:0]

	rs := pgxscan.NewRowScanner(rows)
	for rows.Next() {
		var st Student
		if err := rs.Scan(&st); err != nil {
			log.Fatal(err)
		}
		// do something here
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// see more examples: https://github.com/georgysavva/scany/blob/master/pgxscan/example_test.go
}
