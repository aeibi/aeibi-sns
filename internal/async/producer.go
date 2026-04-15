package async

import (
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type Producer struct {
	Client *river.Client[pgx.Tx]
}

func New(client *river.Client[pgx.Tx]) *Producer {
	return &Producer{
		Client: client,
	}
}
