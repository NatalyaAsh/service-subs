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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(statusCode)
	msg, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal JSON", "error", err)
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
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

// @Summary      Создание новой подписки
// @Description  Добавляет запись о новой подписке пользователя в базу данных
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body models.Sub true "Данные новой подписки (обязательно: user_id, service_name)"
// @Success      201  {object}  models.ResponseId  "ID созданной подписки"
// @Failure      400  {object}  models.ResponseErr "Ошибка валидации данных"
// @Failure      500  {object}  models.ResponseErr "Внутренняя ошибка сервера"
// @Router       /sub [post]
func PostSub(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	writeJson(w, models.ResponseId{ID: id}, http.StatusCreated)
}

// @Summary      Список всех подписок
// @Description  Возвращает массив всех существующих подписок
// @Tags         subscriptions
// @Produce      json
// @Success      200  {array}   models.Sub
// @Failure      500  {object}  models.ResponseErr
// @Router       /subs [get]
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

// @Summary      Получение подписки по ID
// @Description  Возвращает данные одной подписки
// @Tags         subscriptions
// @Produce      json
// @Param        id query int true "ID подписки"
// @Success      200  {object}  models.Sub
// @Failure      400  {object}  models.ResponseErr "Не указан или невалидный ID"
// @Failure      404  {object}  models.ResponseErr "Подписка не найдена"
// @Router       /sub [get]
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

// @Summary      Обновление подписки
// @Description  Обновляет поля существующей подписки. Передавайте ТОЛЬКО те поля, которые нужно изменить.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id      query int       true  "ID подписки для обновления"
// @Param        request body   models.Sub true  "Объект с полями для обновления"
// @Success      200     "Успешно обновлено"
// @Failure      400     {object} models.ResponseErr "Ошибка валидации"
// @Failure      500     {object} models.ResponseErr
// @Router       /sub [put]
func PutSub(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
	writeJson(w, nil, http.StatusOK)
}

// @Summary      Удаление подписки
// @Description  Удаляет запись о подписке по ID
// @Tags         subscriptions
// @Param        id query int true "ID подписки для удаления"
// @Success      200  "Успешно удалено"
// @Failure      400  {object} models.ResponseErr "Не указан или невалидный ID"
// @Failure      500  {object} models.ResponseErr
// @Router       /sub [delete]
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
	writeJson(w, nil, http.StatusOK)
}

// @Summary      Расчет суммы подписок
// @Description  Считает общую стоимость подписок пользователя ЗА период. ВНИМАНИЕ: этот эндпоинт ожидает JSON-тело с фильтрами, хотя использует GET-метод.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body   models.Sub true  "Фильтры: user_id (обяз.), service_name (опц.), start_date, end_date"
// @Success      200  {object}  models.ResponseSum "Сумма подписок"
// @Failure      400  {object}  models.ResponseErr
// @Router       /sum [get]
func GetSumSubs(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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
