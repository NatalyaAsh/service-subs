package models

type ResponseId struct {
	ID int64 `json:"id"`
}

type ResponseErr struct {
	Error string `json:"error"`
}

type Sub struct {
	ID          int    `json:"id"`
	UserId      int    `json:"user_id"`
	ServiceName string `json:"service_name"`
	Price       int    `json:"price"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
}

const (
	Schema_subs = `CREATE TABLE IF NOT EXISTS subs (
    id SERIAL PRIMARY KEY,
		user_id VARCHAR(128) NOT NULL DEFAULT '',
    service_name VARCHAR(128) NOT NULL DEFAULT '',
		price INTEGER NOT NULL DEFAULT 0,
		start_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		end_date TIMESTAMP);

		CREATE INDEX IF NOT EXISTS idxSubsId ON subs (id);`
)
