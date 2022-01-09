package main

type Worker struct {
	Id         int
	FullName   string
	Email      string
	IdPosition int
}

type Bookkeeping struct {
	IdPosition   int
	PositionName string
	Salary       float64
}

type PageDataWorkers struct {
	Title   string
	Columns []string
	Items   []Worker
}

type PageDataBookkeeping struct {
	Title   string
	Columns []string
	Items   []Bookkeeping
}
