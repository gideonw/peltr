package server

type Job struct {
	ID          string
	Req         int
	Concurrency int
	Duration    int
	Rate        int
}
