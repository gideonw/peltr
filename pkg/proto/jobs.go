package proto

type Job struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Req         int    `json:"req"`
	Concurrency int    `json:"concurrency"`
	Duration    int    `json:"duration"`
	Rate        int    `json:"rate"`
}
