package models

// Student - наша бизнес сущность (не зависит от того где и как хранится)
type Student struct {
	ID         int64
	FirstName  string
	LastName   string
	Age        uint
	OtherField string
}
