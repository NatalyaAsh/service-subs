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

	mux.HandleFunc("GET /sum", GetSumSubs)
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

// @Summary		PostSub
// @Tags			Api Subs
// @Description	post sub
// @Accept json
// @Produce json
// @Param input body models.Sub true "sub info for insert"
// @Success 200 {integer} integer 1
// @Router			/api/PostSub [post]
func PostSub(w http.ResponseWriter, r *http.Request) {
	slog.Debug("PostSub")
	var buf bytes.Buffer
	var sub models.Sub

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &sub); err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if sub.ServiceName == "" {
		slog.Error("не указано имя")
		writeJson(w, models.ResponseErr{Error: "не указано имя"}, http.StatusBadRequest)
		return
	}
	//	if sub.UserId == 0 {
	if sub.UserId == "" {
		slog.Error("не указан пользователь подписки")
		writeJson(w, models.ResponseErr{Error: "не указан пользователь подписки"}, http.StatusBadRequest)
		return
	}
	if sub.StartDate == "" {
		sub.StartDate = time.Now().Format("01-2006")
		slog.Debug("PostSub: StartDate := Now")
	} else {
		_, _, err := CheckDate(sub.StartDate)
		if err != nil {
			slog.Error(err.Error())
			writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	}
	if sub.EndDate != "" {
		_, _, err := CheckDate(sub.EndDate)
		if err != nil {
			slog.Error(err.Error())
			writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
			return
		}
	}
	slog.Debug("PostSub", "sub", sub)

	id, err := pgsql.Post(&sub)
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	sub.ID = int(id)
	slog.Info("Добавили запись: ", "sub", sub)
	writeJson(w, models.ResponseId{ID: id}, http.StatusOK)
}

// @Summary		GetSubs
// @Tags			Api Subs
// @Description	get subS
// @Accept json
// @Produce json
// @Success 200 {integer} integer 1
// @Router			/api/Subs [get]
func GetSubs(w http.ResponseWriter, r *http.Request) {
	subs, err := pgsql.GetSubs()
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	slog.Debug("Api GetSubs")
	writeJson(w, subs, http.StatusOK)
}

// @Summary		GetSub
// @Tags			Api Subs
// @Description	get sub
// @Accept json
// @Produce json
// @Success 200 {integer} integer 1
// @Router			/api/Sub [get]
func GetSub(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		slog.Error("не указан Id подписки")
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}

	var err error
	var sub models.Sub

	sub.ID, err = strconv.Atoi(id)
	if err != nil {
		slog.Error("не валидный Id")
		writeJson(w, models.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
		return
	}
	slog.Info("Api GetSub", "id", sub.ID)

	sub, err = pgsql.GetSub(sub.ID)
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	slog.Debug("Api GetSub")
	writeJson(w, sub, http.StatusOK)
}

// @Summary		PutSub
// @Tags			Api Subs
// @Description	update sub
// @Accept json
// @Produce json
// @Param input body models.Sub true "sub info for update"
// @Success 200 {integer} integer 1
// @Router			/api/PutSub [put]
func PutSub(w http.ResponseWriter, r *http.Request) {
	idRaw := r.URL.Query().Get("id")
	if idRaw == "" {
		slog.Error("не указан Id подписки")
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		slog.Error("не валидный Id")
		writeJson(w, models.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer
	var sub models.Sub

	_, err = buf.ReadFrom(r.Body)
	if err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &sub); err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	sub.ID = id

	if sub.ServiceName != "" {
		sub.ServiceNameUpdate = true
		slog.Debug("Api PutSub: Update ServiceName")
	}
	if sub.Price != 0 {
		sub.PriceUpdate = true
		slog.Debug("Api PutSub: Update Price")
	}
	if sub.StartDate != "" {
		sub.StartDateUpdate = true
		slog.Debug("Api PutSub: Update StartDate")
	}
	if sub.EndDate != "" {
		sub.EndDateUpdate = true
		slog.Debug("Api PutSub: Update EndDate")
	}

	err = pgsql.Update(&sub)
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	writeJson(w, "", http.StatusOK)
}

// @Summary		DeleteSub
// @Tags			Api Subs
// @Description	delete sub
// @Accept json
// @Produce json
// @Success 200 {integer} integer 1
// @Router			/api/DeleteSub [delete]
func DeleteSub(w http.ResponseWriter, r *http.Request) {
	idRaw := r.URL.Query().Get("id")
	if idRaw == "" {
		slog.Error("не указан Id подписки")
		writeJson(w, models.ResponseErr{Error: "не указан Id подписки"}, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idRaw)
	if err != nil {
		slog.Error("не валидный Id")
		writeJson(w, models.ResponseErr{Error: "не валидный Id"}, http.StatusBadRequest)
		return
	}

	err = pgsql.Delete(int64(id))
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	slog.Info("Удалили подписку с", "id", id)
	writeJson(w, "", http.StatusOK)
}

// @Summary		GetSumSubs
// @Tags			Api Subs
// @Description	sum subs of user
// @Accept json
// @Produce json
// @Success 200 {integer} integer 1
// @Router			/api/GetSumSubs [get]
func GetSumSubs(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	var sub models.Sub

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if err = json.Unmarshal(buf.Bytes(), &sub); err != nil {
		slog.Error("ошибка передачи данных")
		writeJson(w, models.ResponseErr{Error: "ошибка передачи данных"}, http.StatusBadRequest)
		return
	}
	if sub.UserId == "" {
		slog.Error("не задан id пользователя")
		writeJson(w, models.ResponseErr{Error: "не задан id пользователя"}, http.StatusBadRequest)
		return
	}

	sum, err := pgsql.GetSumSubs(&sub)
	if err != nil {
		slog.Error(err.Error())
		writeJson(w, models.ResponseErr{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	slog.Debug("Api GetSumSubs", "sum", sum, "user_id", sub.UserId)
	writeJson(w, models.ResponseSum{Sum: sum}, http.StatusOK)
}
