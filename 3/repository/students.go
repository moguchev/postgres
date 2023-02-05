package repository

import (
	"context"

	"github.com/moguchev/postgres/3/models"
)

type StudentsRepository interface {
	GetStudent(ctx context.Context, id int64) (models.Student, error)
	GetStudents(ctx context.Context, ids ...int64) ([]models.Student, error)
}
