package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"servicesubs/internal/config"
	"servicesubs/internal/database/pgsql"
	"servicesubs/internal/models"
	"strconv"
	"strings"
	"time"
)

// type Meta struct {
// 	Total   int `json:"total"`
// 	Removed int `json:"removed"`
// 	Limit   int `json:"limit"`
// 	Offset  int `json:"offset"`
// }

// type StructGetGoods struct {
// 	Meta  Meta             `json:"meta"`
// 	Goods *[]modeldb.Goods `json:"goods"`
// }

func Init(mux *http.ServeMux, cfg *config.Config) {
	mux.HandleFunc("POST /sub", PostSub)
	mux.HandleFunc("GET /sub", GetSub)
	mux.HandleFunc("GET /subs", GetSubs)
	mux.HandleFunc("PUT /sub", PutSub)
	mux.HandleFunc("DELETE /sub", DeleteSub)
}

func writeJson(w http.ResponseWriter, data any, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	msg, _ := json.Marshal(data)
	io.Writer.Write(w, msg)
}

func CheckDate(str string) (month int, year int, err error) {
	if len(str) > 7 {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}
	arr := strings.Split(str, "-")
	if len(arr) > 2 {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}

	month, err = strconv.Atoi(arr[0])
	if err != nil {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}
	if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}
	year, err = strconv.Atoi(arr[1])
	if err != nil {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}
	if year < 1900 || year > 2100 {
		return 0, 0, fmt.Errorf("не валидный формат даты")
	}
	return month, year, nil
}

func PostSub(w http.ResponseWriter, r *http.Request) {
	slog.Info("PostSub")
	var buf bytes.Buffer
	var sub models.Sub

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &sub); err != nil {
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if sub.ServiceName == "" {
		writeJson(w, models.ResponseErr{Error: "не указано имя"}, http.StatusBadRequest)
		return
	}
	if sub.UserId == 0 {
		writeJson(w, models.ResponseErr{Error: "не указан пользователь подписки"}, http.StatusBadRequest)
		return
	}
	if sub.StartDate == "" {
		sub.StartDate = time.Now().Format("01-2006")
	} else {
		_, _, err := CheckDate(sub.StartDate)
		if err != nil {
			writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	}
	if sub.EndDate != "" {
		_, _, err := CheckDate(sub.EndDate)
		if err != nil {
			writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	}
	slog.Info("PostSub", "sub", sub)

	id, err := pgsql.Post(&sub)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJson(w, models.ResponseId{ID: id}, http.StatusOK)
}

func GetSubs(w http.ResponseWriter, r *http.Request) {
	subs, err := pgsql.GetSubs()
	if err != nil {
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	writeJson(w, subs, http.StatusOK)
}

func GetSub(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}

	var err error
	var sub models.Sub

	sub.ID, err = strconv.Atoi(id)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
		return
	}
	slog.Info("Api GetSub", "id", sub.ID)

	sub, err = pgsql.GetSub(sub.ID)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	writeJson(w, sub, http.StatusOK)
}

func PutSub(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	var sub models.Sub

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &sub); err != nil {
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if sub.ServiceName == "" {
		writeJson(w, models.ResponseErr{Error: "не указано имя"}, http.StatusBadRequest)
		return
	}
	if sub.UserId == 0 {
		writeJson(w, models.ResponseErr{Error: "не указано имя"}, http.StatusBadRequest)
		return
	}
	// 	good.ProjectId, err = strconv.Atoi(projectId)
	// 	if err != nil {
	// 		writeJson(w, modeldb.ResponseErr{Error: "не валидный projectId"}, http.StatusBadRequest)
	// 		return
	// 	}
	// 	good.ID, err = strconv.Atoi(id)
	// 	if err != nil {
	// 		writeJson(w, modeldb.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
	// 		return
	// 	}

	// 	err = pgsql.Update(&good)
	// 	if err != nil {
	// 		writeJson(w, modeldb.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
	// 		return
	// 	}

	// 	good, err = pgsql.GetGood(good.ID)
	// 	if err != nil {
	// 		writeJson(w, modeldb.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
	// 		return
	// 	}
	// 	// Инвалидируем данные Redis
	// 	err = redis.Set(&good)
	// 	if err != nil {
	// 		writeJson(w, modeldb.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
	// 		return
	// 	}

	// writeJson(w, good, http.StatusOK)
}

func DeleteSub(w http.ResponseWriter, r *http.Request) {
	idRaw := r.URL.Query().Get("id")
	if idRaw == "" {
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		writeJson(w, models.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
		return
	}

	err = pgsql.Delete(int64(id))
	if err != nil {
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	writeJson(w, "", http.StatusOK)
}
