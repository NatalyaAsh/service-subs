package pgsql

import (
	"database/sql"
	"fmt"
	"log/slog"
	"servicesubs/internal/config"
	"servicesubs/internal/models"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func Init(cfg *config.Config) error {
	var err error
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s sslmode=disable",
		cfg.PGS.User, cfg.PGS.Name, cfg.PGS.Password, cfg.PGS.Host)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	slog.Info("Start db PostgreSQL")

	err = db.Ping()
	if err != nil {
		// panic(err)
		return err
	}
	slog.Info("Start db PostgreSQL: Успешное подключение к базе данных!")

	_, err = db.Exec(models.Schema_subs)
	if err != nil {
		// panic(err)
		return err
	}
	slog.Info("Start db PostgreSQL: Таблица SUBS успешно создана или уже была!")

	return nil
}

func CloseDB() {
	db.Close()
}

func Post(sub *models.Sub) (int64, error) {
	var id int64
	startDate, err := time.Parse("01-2006", sub.StartDate)
	if err != nil {
		return 0, err
	}
	if sub.EndDate == "" {
		query := `INSERT INTO subs (service_name, price, user_id, start_date) VALUES ($1, $2, $3, $4) RETURNING id;`
		err = db.QueryRow(query, sub.ServiceName, sub.Price, sub.UserId, startDate).Scan(&id)
	} else {
		query := `INSERT INTO subs (service_name, price, user_id, start_date, end_date) VALUES ($1, $2, $3, $4, $5) RETURNING id;`
		err = db.QueryRow(query, sub.ServiceName, sub.Price, sub.UserId, startDate, sub.EndDate).Scan(&id)
	}
	slog.Info("pgsql Post Exec: insert")
	if err != nil {
		return 0, err
	}
	return id, nil
}

func Update(sub *models.Sub) error {
	query := `UPDATE subs SET service_name=$1, price=$2, start_date=$3 WHERE id=$4`
	if sub.ServiceName != "" {
		query = query + `service_name=$1`
	}
	res, err := db.Exec(query, sub.ServiceName, sub.Price, sub.StartDate, sub.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func Delete(id int64) error {
	query := `DELETE FROM subs WHERE id=$1`
	res, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func GetSub(id int) (models.Sub, error) {
	slog.Info("PostgreSQL: GetSub", "id", id)
	row := db.QueryRow(`SELECT id, service_name, price, user_id, start_date, end_date FROM subs WHERE id=$1`, id)
	if row == nil {
		return models.Sub{}, fmt.Errorf("good not found")
	}

	var sub models.Sub
	var endDateRaw sql.NullString
	err := row.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserId, &sub.StartDate, &endDateRaw)
	if err != nil {
		return models.Sub{}, err
	}
	if endDateRaw.Valid {
		sub.EndDate = endDateRaw.String
	} else {
		sub.EndDate = ""
	}

	return sub, nil
}

func GetSubs() (*[]models.Sub, error) {
	slog.Info("PostgreSQL: GetSubs")
	rows, err := db.Query(`SELECT id, service_name, price, user_id, start_date, end_date FROM subs`)
	if err != nil {
		slog.Error(err.Error())
		return &[]models.Sub{}, err
	}
	defer rows.Close()

	subs := []models.Sub{}
	for rows.Next() {
		var sub models.Sub
		var endDateRaw sql.NullString
		err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserId, &sub.StartDate, &endDateRaw)
		if err != nil {
			slog.Error(err.Error())
			return &[]models.Sub{}, err
		}
		if endDateRaw.Valid {
			sub.EndDate = endDateRaw.String
		} else {
			sub.EndDate = ""
		}
		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		slog.Error(err.Error())
		return &[]models.Sub{}, err
	}

	return &subs, nil
}
