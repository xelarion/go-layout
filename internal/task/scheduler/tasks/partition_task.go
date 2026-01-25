// Package tasks provides scheduler task implementations.
package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/xelarion/go-layout/internal/task"
	"github.com/xelarion/go-layout/internal/task/scheduler"
	"github.com/xelarion/go-layout/pkg/errs"
)

// PartitionConfig defines configuration for a table's time-based partitioning
type PartitionConfig struct {
	// TableName is the parent table name that's partitioned
	TableName string
	// PartitionKeyColumn is the column used for RANGE partitioning (typically a timestamp column)
	PartitionKeyColumn string
	// PartitionInterval defines the time interval for each partition ('MONTH', 'YEAR', etc.)
	PartitionInterval string
	// PartitionNameFormat defines how partition names are generated, must contain a time format string
	// Example: "%s_%s" combined with TimeFormat will determine the partition name
	PartitionNameFormat string
	// TimeFormat defines the format for the time portion of partition names (e.g., "200601" for YYYYMM)
	TimeFormat string
	// PartitionsAhead defines how many future partitions to create in advance
	PartitionsAhead int
	// SubPartitionColumn is optional, for tables that need sub-partitioning
	SubPartitionColumn string
	// SubPartitionType is the type of sub-partitioning, either "LIST" or "HASH"
	SubPartitionType string
	// SubPartitionCount is the number of sub-partitions if using HASH
	SubPartitionCount int
	// UseUTC determines if partition boundaries should use UTC time (true) or local time (false)
	UseUTC bool
}

// PartitionHandler is a partition scheduler task handler.
type PartitionHandler struct {
	deps   *task.Dependencies
	logger *zap.Logger
	// Configurations for tables that need partitioning
	partitionConfigs []PartitionConfig
}

// NewPartitionHandler creates a new partition handler.
func NewPartitionHandler(deps *task.Dependencies, logger *zap.Logger) *PartitionHandler {
	handler := &PartitionHandler{
		deps:   deps,
		logger: logger.Named("partition-handler"),
		partitionConfigs: []PartitionConfig{
			{
				// Configuration for user_notifications
				TableName:           "user_notifications",
				PartitionKeyColumn:  "created_at",
				PartitionInterval:   "MONTH",
				PartitionNameFormat: "%s_%s",
				TimeFormat:          "200601", // YYYYMM
				PartitionsAhead:     3,
				UseUTC:              true, // 使用UTC时间作为分区边界，这样确保与数据库的TIMESTAMP WITH TIME ZONE一致
			},
			// Add more table configurations as needed
		},
	}
	return handler
}

// Register registers this handler with the scheduler.
func (t *PartitionHandler) Register(s *scheduler.Scheduler) error {
	// Runs every day at midnight
	if err := s.Register("partition-task", "0 0 0 * * *", t.Execute); err != nil {
		return err
	}
	return nil
}

// Execute runs the partition task handler.
func (t *PartitionHandler) Execute(ctx context.Context) error {
	db := t.deps.Data().DB(ctx)

	var err error
	for _, config := range t.partitionConfigs {
		// Create partitions based on configuration
		if e := t.createPartitions(ctx, db, config); e != nil {
			err = e
			continue
		}
	}

	return err
}

// createPartitions creates time-based partitions for a given table configuration
func (t *PartitionHandler) createPartitions(ctx context.Context, db *gorm.DB, config PartitionConfig) error {
	// 获取当前时间，根据配置决定使用UTC还是本地时区
	var now time.Time
	if config.UseUTC {
		now = time.Now().UTC()
	} else {
		now = time.Now()
	}

	for i := 0; i < config.PartitionsAhead; i++ {
		var startTime, endTime time.Time
		// 确定使用的时区
		tz := time.Local
		if config.UseUTC {
			tz = time.UTC
		}

		// Calculate partition time boundaries based on interval
		switch config.PartitionInterval {
		case "MONTH":
			startTime = time.Date(now.Year(), now.Month()+time.Month(i+1), 1, 0, 0, 0, 0, tz)
			endTime = time.Date(now.Year(), now.Month()+time.Month(i+2), 1, 0, 0, 0, 0, tz)
		case "YEAR":
			startTime = time.Date(now.Year()+i+1, 1, 1, 0, 0, 0, 0, tz)
			endTime = time.Date(now.Year()+i+2, 1, 1, 0, 0, 0, 0, tz)
		default:
			return errs.NewBusiness(fmt.Sprintf("unsupported partition interval: %s", config.PartitionInterval))
		}

		// Format partition name
		partitionName := fmt.Sprintf(config.PartitionNameFormat, config.TableName, startTime.Format(config.TimeFormat))

		// Check if partition already exists
		var exists bool
		checkSQL := `
			SELECT EXISTS (
				SELECT 1 FROM pg_class c
				JOIN pg_namespace n ON n.oid = c.relnamespace
				WHERE c.relname = ?
			)
		`

		if err := db.WithContext(ctx).Raw(checkSQL, partitionName).Scan(&exists).Error; err != nil {
			return errs.WrapInternal(err, "failed to check if partition exists")
		}

		if exists {
			continue
		}

		// Create partition
		createSQL := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
			FOR VALUES FROM ('%s') TO ('%s')
		`, partitionName, config.TableName, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

		// Add sub-partitioning if configured
		if config.SubPartitionColumn != "" {
			switch config.SubPartitionType {
			case "HASH":
				createSQL += fmt.Sprintf(` PARTITION BY HASH (%s)`, config.SubPartitionColumn)

				// Execute the partition creation
				if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
					return errs.WrapInternal(err, "failed to create partition")
				}

				// Create hash sub-partitions
				for j := 0; j < config.SubPartitionCount; j++ {
					subPartitionName := fmt.Sprintf("%s_p%d", partitionName, j)
					subPartitionSQL := fmt.Sprintf(`
						CREATE TABLE IF NOT EXISTS %s PARTITION OF %s
						FOR VALUES WITH (modulus %d, remainder %d)
					`, subPartitionName, partitionName, config.SubPartitionCount, j)

					if err := db.WithContext(ctx).Exec(subPartitionSQL).Error; err != nil {
						return errs.WrapInternal(err, "failed to create sub-partition")
					}
				}
			default:
				return errs.NewBusiness(fmt.Sprintf("unsupported sub-partition type: %s", config.SubPartitionType))
			}
		} else {
			// Execute the simple partition creation without sub-partitioning
			if err := db.WithContext(ctx).Exec(createSQL).Error; err != nil {
				return errs.WrapInternal(err, "failed to create partition")
			}
		}

		t.logger.Info("Created partition", zap.String("partition", partitionName))
	}

	return nil
}
