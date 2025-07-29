package models

import (
	"github.com/google/uuid"
)

type Subscription struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"service_name" binding:"required"`
	Price       int       `json:"price" db:"price" binding:"required"`
	UserID      uuid.UUID `json:"user_id" db:"user_id" binding:"required"`
	StartDate   string    `json:"start_date" db:"start_date" binding:"required"`
	EndDate     *string   `json:"end_date,omitempty" db:"end_date"`
}
