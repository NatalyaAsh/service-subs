package pgsql

import (
	"database/sql"
	"fmt"
	"log/slog"
	"servicesubs/internal/config"
	"servicesubs/internal/models"
	"strings"
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
	slog.Debug("Start db PostgreSQL")

	err = db.Ping()
	if err != nil {
		// panic(err)
		return err
	}
	slog.Debug("Start db PostgreSQL: Успешное подключение к базе данных!")

	_, err = db.Exec(models.Schema_subs)
	if err != nil {
		// panic(err)
		return err
	}
	slog.Debug("Start db PostgreSQL: Таблица SUBS успешно создана или уже была!")

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
	slog.Debug("pgsql Insert")
	if err != nil {
		return 0, err
	}
	return id, nil
}

func Update(sub *models.Sub) error {
	flag := false
	fields := ""
	var params []interface{}
	i := 1
	//query := `UPDATE subs SET service_name=$1, price=$2, start_date=$3 WHERE id=$4`
	query := `UPDATE subs SET `
	if sub.ServiceNameUpdate {
		fields = fmt.Sprintf("service_name = $%d", i)
		params = append(params, sub.ServiceName)
		i++
		flag = true
	}
	if sub.PriceUpdate {
		if flag {
			fields += ", "
			//params += ", "
		}
		fields += fmt.Sprintf("price = $%d", i)
		//params += strconv.Itoa(sub.Price)
		params = append(params, sub.Price)
		i++
		flag = true
	}
	if sub.StartDateUpdate {
		if flag {
			fields += ", "
			//params += ", "
		}
		fields += fmt.Sprintf("start_date = $%d", i)
		//params += sub.StartDate
		params = append(params, sub.StartDate)
		i++
		flag = true
	}
	if sub.EndDateUpdate {
		if flag {
			fields += ", "
			//params += ", "
		}
		fields += fmt.Sprintf("end_date = $%d", i)
		//params += sub.EndDate
		params = append(params, sub.EndDate)
		i++
		flag = true
	}

	query = query + fields + fmt.Sprintf(` WHERE id=$%d`, i)
	params = append(params, sub.ID)
	slog.Debug("pgsql Update ", "query", query)
	slog.Debug("pgsql Update ", "params", params)

	res, err := db.Exec(query, params...)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		slog.Error(err.Error())
		return err
	}
	if count == 0 {
		slog.Error(`incorrect id for updating sub`)
		return fmt.Errorf(`incorrect id for updating sub`)
	}
	slog.Debug("pgsql Update")
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
		return fmt.Errorf(`incorrect id for deleting sub`)
	}
	slog.Debug("pgsql Delete")
	return nil
}

func GetSub(id int) (models.Sub, error) {
	slog.Debug("PostgreSQL: GetSub", "id", id)
	row := db.QueryRow(`SELECT id, service_name, price, user_id, start_date, end_date FROM subs WHERE id=$1`, id)
	if row == nil {
		return models.Sub{}, fmt.Errorf("sub not found")
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
	slog.Debug("PostgreSQL: GetSubs")
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
		return &[]models.Sub{}, err
	}
	return &subs, nil
}

func GetSumSubs(sub *models.Sub) (int64, error) {
	//flag := false
	var fields []string
	var params []interface{}
	i := 2
	fields = append(fields, `SELECT sum(price) FROM subs WHERE user_id=$1 `)
	params = append(params, sub.UserId)
	if sub.StartDate != "" {
		fields = append(fields, fmt.Sprintf(" AND start_date >= $%d", i))
		params = append(params, sub.StartDate)
		i++
	}
	if sub.EndDate != "" {
		fields = append(fields, fmt.Sprintf(" AND start_date <= $%d", i))
		params = append(params, sub.EndDate)
		i++
	}
	if sub.ServiceName != "" {
		fields = append(fields, fmt.Sprintf(" AND service_name LIKE $%d", i))
		params = append(params, sub.ServiceName)
	}

	query := strings.Join(fields, "")
	slog.Debug("pgsql Update ", "query", query)
	slog.Debug("pgsql Update ", "params", params)

	row := db.QueryRow(query, params...)
	if row == nil {
		slog.Error("подписки этого пользователя не найдены")
		return 0, fmt.Errorf(`подписки этого пользователя не найдены`)
	}
	var sumSubs int64
	err := row.Scan(&sumSubs)
	if err != nil {
		slog.Error(err.Error())
		return 0, err
	}
	slog.Debug("pgsql GetSum>Sub", "user_id", sub.UserId, "sum", sumSubs)
	return sumSubs, nil
}
