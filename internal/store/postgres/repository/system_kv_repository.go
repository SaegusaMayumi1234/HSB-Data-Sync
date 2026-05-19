package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/saegusamayumi1234/hsb-data-sync/internal/store/postgres/model"
)

type SystemKVRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSystemKVRepository(db *pgxpool.Pool) *SystemKVRepository {
	return &SystemKVRepository{db: db}
}

func (r *SystemKVRepository) GetValueByKey(ctx context.Context, key string) (*model.SystemKVModel, error) {
	query := `
		SELECT
			id,
			key,
			value,
			created_by,
			created_at,
			updated_by,
			updated_at,
			deleted_by,
			deleted_at
		FROM system_kv
		WHERE key = $1
	`

	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		return nil, err
	}

	item, err := pgx.CollectOneRow(rows, pgx.RowToAddrOfStructByName[model.SystemKVModel])
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (r *SystemKVRepository) GetValuesByKeys(ctx context.Context, keys []string) ([]*model.SystemKVModel, error) {
	if len(keys) == 0 {
		return []*model.SystemKVModel{}, nil
	}

	query := `
		SELECT
			id,
			key,
			value,
			created_by,
			created_at,
			updated_by,
			updated_at,
			deleted_by,
			deleted_at
		FROM system_kv
		WHERE key = ANY($1::text[])
	`

	rows, err := r.db.Query(ctx, query, keys)
	if err != nil {
		return nil, err
	}

	items, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[model.SystemKVModel])
	if err != nil {
		return nil, err
	}

	return items, nil
}
