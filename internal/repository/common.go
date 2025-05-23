package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/xelarion/go-layout/pkg/errs"
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
