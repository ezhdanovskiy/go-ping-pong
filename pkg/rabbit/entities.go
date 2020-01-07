package rabbit

type Event struct {
	ID int `json:"id"`
}

const (
	DefaultPingsQueueName = "pings"
	DefaultPongsQueueName = "pongs"
)
