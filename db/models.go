package db

type User struct {
	ID int64
}

type Category struct {
	ID       int
	Name     string
	UserID   int64
	LimitSum float64
}

type Expense struct {
	ID         int
	UserID     int64
	CategoryID int
	Amount     float64
	Date       string // YYYY-MM-DD
}
