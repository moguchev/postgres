package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

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
