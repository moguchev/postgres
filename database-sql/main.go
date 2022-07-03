package main

import (
	"context"
	"database/sql" // https://go.dev/src/database/sql/doc.txt
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	// database/sql — это набор интерфейсов для работы с базой
	// Чтобы эти интерфейсы работали, для них нужна реализация. Именно за реализацию и отвечают драйверы.

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

		See more drivers: https://zchee.github.io/golang-wiki/SQLDrivers/
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
	db, err := sql.Open(stdPostgresDriverName, psqlConn) // returns *sql.DB, error
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

	fmt.Println("Connection with database successfully established!")

	/* Настройка пула соединений */
	db.SetConnMaxIdleTime(time.Minute) // время, в течение которого соединение может быть бездействующим.
	db.SetConnMaxLifetime(time.Hour)   // время, в течение которого соединение может быть повторно использовано.
	db.SetMaxIdleConns(2)              // максимум 2 простаивающих соединения
	db.SetMaxOpenConns(4)              // максимум 4 открытых соединений с БД

	/* статистика пула соединений */
	statistics := db.Stats()
	bytes, err := json.Marshal(statistics)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("db connection statistics: %s\n", string(bytes))

	/* примеры работы c БД */
	exampleQueryRow(db)
	exampleQuery(db)
	exampleExec(db)

	/* примеры работы c БД c контекстом */
	// Совет: используйте запросы с контекстом
	ctx := context.Background()

	exampleQueryRowContext(ctx, db)
	exampleQueryContext(ctx, db)
	exampleExecContext(ctx, db)

	/* примеры работы c транзакциями */
	exampleTransaction(ctx, db)

	/* пример работы с nullable полями*/
	exampleWithNullableFields(ctx, db)

	// Более подробный туториал (правда там с MySQL, но суть та же)
	// Go database/sql tutorial: http://go-database-sql.org/
}

func exampleQueryRow(db *sql.DB) {
	// Ex. 1
	row := db.QueryRow("SELECT count(*) FROM students")

	var totalStudents uint
	if err := row.Scan(&totalStudents); err != nil { // Обязательно передаем адрес переменной, куда будем сканировать значение.
		log.Fatal(err)
	}

	fmt.Printf("total students: %d\n", totalStudents)

	// Ex. 2
	var studentID int64
	row = db.QueryRow("SELECT id FROM students WHERE age = 10000") // такого "долгожителя" в нашей таблице может не быть
	if errors.Is(row.Err(), sql.ErrNoRows) {
		fmt.Println("Не найден в БД студент с age > 10000")
	}

	if err := row.Scan(&studentID); err != nil { // мы тут получим ошубку, так как нам ничего не вернулось из БД
		fmt.Println("db.QueryRow.Scan():", err) // нам вернется ошибка sql.ErrNoRows
		if errors.Is(err, sql.ErrNoRows) {      // при использовании QueryRow не забывайте обрабатывать ошибку на sql.ErrNoRows, так как отстуствие результата может быть стандартным кейсом
			fmt.Println("Не найден в БД студент с age > 10000")
		}
	}
}

func exampleQuery(db *sql.DB) {
	const minAge = 18

	rows, err := db.Query("SELECT first_name, last_name, age FROM students WHERE age >= $1", minAge)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close() // Обязательно закрываем иначе соединение с БД повиснет

	type Student struct {
		FirstName string
		LastName  string
		Age       uint
	}

	var students []Student
	for rows.Next() {
		var st Student
		if err := rows.Scan(&st.FirstName, &st.LastName, &st.Age); err != nil {
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

func exampleExec(db *sql.DB) {
	// Ex. 1:
	const notExistedStudentID = 1234567
	result, err := db.Exec("UPDATE students SET age = age+1 WHERE id = $1", notExistedStudentID)
	if err != nil {
		log.Fatal(err)
	}

	var (
		rowsAffected, lastInsertId int64
	)

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		fmt.Println("sql.Result.RowsAffected():", err) // ok
	}
	lastInsertId, err = result.LastInsertId()
	if err != nil {
		fmt.Println("sql.Result.LastInsertId():", err) // LastInsertId is not supported by "postgres" driver
	}
	fmt.Printf("rows affected: %d, last insert id: %d\n", rowsAffected, lastInsertId)

	// Ex. 2:
	const studentID = 1
	result, err = db.Exec("UPDATE students SET age = age+1 WHERE id = $1", studentID)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		fmt.Println("sql.Result.RowsAffected():", err)
	}
	fmt.Printf("rows affected: %d, last insert id: %d\n", rowsAffected, lastInsertId)
}

func exampleQueryRowContext(ctx context.Context, db *sql.DB) {
	row := db.QueryRowContext(ctx, "SELECT count(*) FROM students")

	var totalStudents uint
	if err := row.Scan(&totalStudents); err != nil { // Обязательно передаем адрес переменной, куда будем сканировать значение.
		log.Fatal(err)
	}

	fmt.Printf("total students: %d\n", totalStudents)
}

func exampleQueryContext(ctx context.Context, db *sql.DB) {
	const minAge = 18

	rows, err := db.QueryContext(ctx, "SELECT first_name, last_name, age FROM students WHERE age >= $1", minAge)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close() // Обязательно закрываем иначе соединение с БД повиснет

	type Student struct {
		FirstName string
		LastName  string
		Age       uint
	}

	var students []Student
	for rows.Next() {
		var st Student
		if err := rows.Scan(&st.FirstName, &st.LastName, &st.Age); err != nil {
			log.Fatal(err)
		}
		students = append(students, st)
	}
	// Внутри драйвера мы получаем данные, накапливая их в буфер размером 4KB.
	// rows.Next() порождает поход в сеть и наполняет буфер. Если буфера не хватает,
	// то мы идём в сеть за оставшимися данными. Больше походов в сеть – меньше скорость обработки.

	fmt.Printf("students: %v\n", students)
}

func exampleExecContext(ctx context.Context, db *sql.DB) {
	const studentID = 1
	result, err := db.ExecContext(ctx, "UPDATE students SET age = age+1 WHERE id = $1", studentID)
	if err != nil {
		log.Fatal(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println("sql.Result.RowsAffected():", err)
	}
	fmt.Printf("rows affected: %d\n", rowsAffected)
}

func exampleTransaction(ctx context.Context, db *sql.DB) {
	// создаем транзакцию
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable, // указываем в опциях уровень изоляции
		ReadOnly:  false,                 // можем указать, что транзакции только для чтения
	})
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback() // если на любом из этапов произойдет ошибка, то мы откатим изменения при выходе из функции.
	// После вызова tx.Commit() вызов tx.Rollback() ничего уже не откатит, а просто вернет ошибку sql.ErrTxDone

	rows, err := tx.QueryContext(ctx, "SELECT 1") // у sql.Tx все те же селекторы что и у sql.DB
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	_, err = tx.ExecContext(ctx, "SELECT 1")
	if err != nil {
		log.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("transaction is commited")
}

func exampleWithNullableFields(ctx context.Context, db *sql.DB) {
	var number int
	if err := db.QueryRowContext(ctx, "SELECT null").Scan(&number); err != nil {
		fmt.Println("error scan:", err) // вернет ошибку, так как нельзя NULL сложить в нессылочные типы
	}
	// Как быть?

	// Вариант 1: COALESCE(field, <default_value>)
	if err := db.QueryRowContext(ctx, "SELECT COALESCE(null, -1) AS some_field").Scan(&number); err != nil {
		log.Fatal(err)
	}
	fmt.Println("number =", number) // number = -1
	// Преимущества:
	// - ничего не меням в коде
	// Недостатски:
	// - можно забыть в запросе использовать COALESCE, и тогда запросы будут валиться с ошибкой
	// - иногда нам важно отличать NULL от значения по умолчанию

	// Вариант 2: давайте сделаем из int ссылочный тип - указатель
	var ptrNumber *int // теперь у нас не int, а указатель на int
	if err := db.QueryRowContext(ctx, "SELECT null").Scan(&ptrNumber); err != nil {
		log.Fatal(err)
	}
	fmt.Println("ptrNumber =", ptrNumber) // тут мы увидем, что numberCanStoreNul == nil
	if ptrNumber != nil {                 // делаем постоянную проверку на nil
		fmt.Println("value of ptrNumber =", *ptrNumber) // разыменовываем указатель
	}

	// Преимущества:
	// - легко из обычного типа сделать указатель
	// Недостатски:
	// - везде в приложении нам надо делать проверки на nil и разыменовывать указатель
	// - Наша переменная будет теперь аллоцировать в куче, что создает накладки для работы приложения на Go

	// Вариант 3: за нас уже позаботились и сделали специальные типы в пакете database/sql:
	/*
		sql.NullInt16
		sql.NullInt32
		sql.NullInt64
		sql.NullByte
		sql.NullBool
		sql.NullFloat64
		sql.NullString
		sql.NullTime
	*/

	var sqlNullNumber sql.NullInt32
	if err := db.QueryRowContext(ctx, "SELECT null").Scan(&sqlNullNumber); err != nil {
		log.Fatal(err)
	}
	fmt.Println("sqlNullNumber =", sqlNullNumber) // sqlNullNumber это структура
	if sqlNullNumber.Valid {                      // поле Valid = true сообщает нам, что поле не Null и можно брать занчение
		fmt.Println("value of ptrNumber =", sqlNullNumber.Int32) // получаем значение
	}
	// Преимущества:
	// - выделяется на стеке переменная
	// - все так же имеем возможность отличить NULL от 0
	// Недостатки:
	// - А что если мы хоти такое поведение для нашего кастомного типа? Нам не хватает стандартного набора.

	// Нам необходмо, чтобы наш тип удовлетоварял sql.Scanner интерфейсу
}

type MyCustomType struct {
	Number int
	Valid  bool
}

// Scan implements the Scanner interface.
func (n *MyCustomType) Scan(src interface{}) error {
	// The src value will be of one of the following types:
	//
	//    int64
	//    float64
	//    bool
	//    []byte
	//    string
	//    time.Time
	//    nil - for NULL values
	if src == nil {
		n.Number, n.Valid = 0, false
		return nil
	}
	n.Valid = true

	// some fantastic logic here
	switch src := src.(type) {
	case int64:
		n.Number = int(src)
	case bool:
		n.Number = 1
	default:
		return fmt.Errorf("can't scan %#v into MyCustomType", src)
	}

	return nil
}

var _ sql.Scanner = (*MyCustomType)(nil) // наш тип MyCustomType удовлетовряет интерфейсу sql.Scanner

// Если мы хоти наш тип как-то мапить в null/valu, то реализуем Value()

// Value implements the driver Valuer interface.
func (n MyCustomType) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}
	return int64(n.Number), nil
}

var _ driver.Valuer = (*MyCustomType)(nil) // наш тип MyCustomType удовлетовряет интерфейсу driver.Valuer
