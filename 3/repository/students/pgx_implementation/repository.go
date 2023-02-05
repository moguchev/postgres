package pgximplementation

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lib/pq"
	"github.com/moguchev/postgres/3/models"
	"github.com/moguchev/postgres/3/repository"
)

// проверка удовлетворению интерфейса repository.StudentsRepository
var _ repository.StudentsRepository = (*studentsRepository)(nil)

type studentsRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool /*logger*/) *studentsRepository {
	return &studentsRepository{
		pool: pool,
	}
}

func (r *studentsRepository) GetStudent(ctx context.Context, id int64) (models.Student, error) {
	const query = `
	SELECT id, first_name, last_name, age 
	FROM students
	WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)

	var student models.Student
	if err := row.Scan(
		&student.ID,
		&student.FirstName,
		&student.LastName,
		&student.Age,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) { // обязательно обрабатываем(перехватываем) известные нам ошибки уровня БД и отдаем наружу уже обработанные
			return models.Student{}, models.ErrNotFound
		}
		log.Printf("get student %d: database error: %s", id, err) // логируем внутренние ошибки и НЕ пробрасываем их наверх
		return models.Student{}, models.ErrInternal
	}

	return student, nil
}

func (r *studentsRepository) GetStudents(ctx context.Context, ids ...int64) ([]models.Student, error) {
	const query = `
	SELECT id, first_name, last_name, age 
	FROM students
	WHERE ids = ANY($1)`

	rows, err := r.pool.Query(ctx, query, pq.Array(ids))
	if err != nil {
		log.Printf("get students %v: database error: %s", ids, err) // логируем внутренние ошибки и НЕ пробрасываем их наверх
		return nil, models.ErrInternal
	}
	defer rows.Close()

	students := make([]models.Student, 0, len(ids))
	for rows.Next() {
		var student models.Student
		if err = rows.Scan(
			&student.ID,
			&student.FirstName,
			&student.LastName,
			&student.Age,
		); err != nil {
			log.Printf("get students %v: scan error: %s", ids, err) // логируем внутренние ошибки и НЕ пробрасываем их наверх
			return nil, models.ErrInternal
		}
		students = append(students, student)
	}

	if err = rows.Err(); err != nil {
		log.Printf("get students %v: rows error: %s", ids, err) // логируем внутренние ошибки и НЕ пробрасываем их наверх
		return nil, models.ErrInternal
	}

	return students, nil
}
