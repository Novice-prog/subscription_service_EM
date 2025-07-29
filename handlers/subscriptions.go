package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"rest-service/models"
)

// Handler хранит ссылку на БД
type Handler struct {
	db *sqlx.DB
}

// ErrorResponse описывает формат ответа с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewHandler создаёт новый Handler
func NewHandler(db *sqlx.DB) *Handler {
	return &Handler{db: db}
}

// CreateSubscription создает новую подписку
// @Summary Создание подписки
// @Description Создает новую запись о подписке
// @Accept json
// @Produce json
// @Param subscription body models.Subscription true "Информация о подписке"
// @Success 201 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	var input models.Subscription
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Генерируем новый UUID
	input.ID = uuid.New()

	// Парсим дату начала в формате "MM-YYYY"
	start, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, use MM-YYYY"})
		return
	}
	startDate := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Парсим дату окончания, если она есть
	var endDate *time.Time
	if input.EndDate != nil {
		e, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_date format, use MM-YYYY"})
			return
		}
		eDate := time.Date(e.Year(), e.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = &eDate
	}

	// Вставляем в БД уже корректные time.Time
	query := `
        INSERT INTO subscriptions
        (id, service_name, price, user_id, start_date, end_date)
        VALUES
        (:id, :service_name, :price, :user_id, :start_date, :end_date)
    `
	params := map[string]interface{}{
		"id":           input.ID,
		"service_name": input.ServiceName,
		"price":        input.Price,
		"user_id":      input.UserID,
		"start_date":   startDate,
		"end_date":     endDate,
	}
	if _, err := h.db.NamedExec(query, params); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// ListSubscriptions возвращает список подписок
// @Summary Список подписок
// @Description Получить все подписки или отфильтрованные
// @Produce json
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Success 200 {array} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	var subs []models.Subscription
	userID := c.Query("user_id")
	service := c.Query("service_name")

	baseQuery := `SELECT * FROM subscriptions WHERE 1=1`
	args := map[string]interface{}{}

	if userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
			return
		}
		baseQuery += ` AND user_id = :user_id`
		args["user_id"] = uid
	}
	if service != "" {
		baseQuery += ` AND service_name ILIKE :service_name`
		args["service_name"] = "%" + service + "%"
	}

	stmt, err := h.db.PrepareNamed(baseQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if err := stmt.Select(&subs, args); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, subs)
}

// GetSubscription возвращает подписку по ID
// @Summary Подписка по ID
// @Description Получить подписку по ID
// @Produce json
// @Param id path string true "UUID подписки"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var sub models.Subscription
	if err := h.db.Get(&sub, `SELECT * FROM subscriptions WHERE id = $1`, uid); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, sub)
}

// UpdateSubscription обновляет подписку по ID
// @Summary Обновить подписку
// @Description Обновляет данные подписки по ID
// @Accept json
// @Produce json
// @Param id path string true "UUID подписки"
// @Param subscription body models.Subscription true "Новые данные"
// @Success 200 {object} models.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var input models.Subscription
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	input.ID = uid

	// Парсим start_date
	start, err := time.Parse("01-2006", input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid start_date format, use MM-YYYY"})
		return
	}
	startDate := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Парсим end_date
	var endDate *time.Time
	if input.EndDate != nil {
		e, err := time.Parse("01-2006", *input.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid end_date format, use MM-YYYY"})
			return
		}
		eDate := time.Date(e.Year(), e.Month(), 1, 0, 0, 0, 0, time.UTC)
		endDate = &eDate
	}

	// Обновляем в БД
	query := `
        UPDATE subscriptions
        SET service_name = :service_name,
            price        = :price,
            user_id      = :user_id,
            start_date   = :start_date,
            end_date     = :end_date
        WHERE id = :id
    `
	params := map[string]interface{}{
		"id":           input.ID,
		"service_name": input.ServiceName,
		"price":        input.Price,
		"user_id":      input.UserID,
		"start_date":   startDate,
		"end_date":     endDate,
	}
	res, err := h.db.NamedExec(query, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}

	c.JSON(http.StatusOK, input)
}

// DeleteSubscription удаляет подписку по ID
// @Summary Удалить подписку
// @Description Удаляет подписку по ID
// @Param id path string true "UUID подписки"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	res, err := h.db.Exec(`DELETE FROM subscriptions WHERE id = $1`, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetSummary считает суммарную стоимость подписок за период
// @Summary Суммарная стоимость
// @Description Подсчитывает сумму подписок за указанный период с фильтрацией
// @Produce json
// @Param user_id query string false "UUID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param from query string true "Начало периода (MM-YYYY)"
// @Param to query string true "Конец периода (MM-YYYY)"
// @Success 200 {object} map[string]int64
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /summary [get]
func (h *Handler) GetSummary(c *gin.Context) {
	userID := c.Query("user_id")
	service := c.Query("service_name")
	from := c.Query("from")
	to := c.Query("to")

	// Парсинг дат
	startDate, err := time.Parse("01-2006", from)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid from format, use MM-YYYY"})
		return
	}

	endDate, err := time.Parse("01-2006", to)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid to format, use MM-YYYY"})
		return
	}

	// Определение границ периода
	start := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(endDate.Year(), endDate.Month()+1, 1, 0, 0, 0, -1, time.UTC)

	// Формируем запрос
	query := `
      SELECT COALESCE(SUM(price),0)
      FROM subscriptions
      WHERE start_date <= $2
        AND (end_date IS NULL OR end_date >= $1)
    `
	args := []interface{}{start, end}

	// Фильтры
	if userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
			return
		}
		query += " AND user_id = $3"
		args = append(args, uid)
	}
	if service != "" {
		query += " AND service_name ILIKE $4"
		args = append(args, "%"+service+"%")
	}

	// Выполняем запрос
	var total sql.NullInt64
	if err := h.db.Get(&total, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	// Если нет записей, возвращаем 0
	if !total.Valid {
		total.Int64 = 0
	}
	c.JSON(http.StatusOK, gin.H{"total": total.Int64})
}
