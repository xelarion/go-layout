package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/pkg/errs"
)

const (
	PgDuplicateKeyErrorCode = "23505" // Postgres 唯一性校验错误
)

// IsExists is a generic function to check if a record exists in the database.
func IsExists(ctx context.Context, db *gorm.DB, model any, filters map[string]any, notFilters map[string]any) (bool, error) {
	query := db.WithContext(ctx).Model(model)

	for field, value := range filters {
		if value == nil {
			continue
		}

		query = query.Where(field+" = ?", value)
	}

	for field, value := range notFilters {
		if value == nil {
			continue
		}

		query = query.Where(field+" != ?", value)
	}

	existFlag := 0
	if err := query.Select("1").Take(&existFlag).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errs.WrapInternal(err, "failed to check record exists")
	}

	return true, nil
}

// Count is a generic function to count the number of records in the database.
func Count(ctx context.Context, db *gorm.DB, model any, filters map[string]any, notFilters map[string]any) (int64, error) {
	query := db.WithContext(ctx).Model(model)

	for field, value := range filters {
		if value == nil {
			continue
		}

		query = query.Where(field+" = ?", value)
	}

	for field, value := range notFilters {
		if value == nil {
			continue
		}

		query = query.Where(field+" != ?", value)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, errs.WrapInternal(err, "failed to count records")
	}

	return count, nil
}

// ApplyORConditions 应用 OR 条件到查询
// 用于数据权限过滤，支持 OR 逻辑组合多个条件
func ApplyORConditions(db *gorm.DB, conditions []map[string]any) *gorm.DB {
	if len(conditions) == 0 {
		return db
	}

	// 构建 OR 条件组
	orQuery := db.Session(&gorm.Session{NewDB: true})
	for i, condition := range conditions {
		subQuery := db.Session(&gorm.Session{NewDB: true})
		for key, value := range condition {
			subQuery = subQuery.Where(key+" = ?", value)
		}

		if i == 0 {
			orQuery = subQuery
		} else {
			orQuery = orQuery.Or(subQuery)
		}
	}

	return db.Where(orQuery)
}

// IsDuplicateKeyError 是否是 PG 唯一性校验错误
func IsDuplicateKeyError(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		if pgError.Code == PgDuplicateKeyErrorCode {
			return true
		}
	}
	return false
}
