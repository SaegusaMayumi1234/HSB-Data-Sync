package model

import (
	"encoding/json"
	"time"
)

type SystemKVModel struct {
	Id        int64           `db:"id"`
	Key       string          `db:"key"`
	Value     json.RawMessage `db:"value"`
	CreatedBy string          `db:"created_by"`
	CreatedAt time.Time       `db:"created_at"`
	UpdatedBy *string         `db:"updated_by"`
	UpdatedAt *time.Time      `db:"updated_at"`
	DeletedBy *string         `db:"deleted_by"`
	DeletedAt *time.Time      `db:"deleted_at"`
}