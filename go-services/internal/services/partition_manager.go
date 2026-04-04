package services

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

// EnsurePartitions calls the database function ensure_partitions() for each
// partitioned table. It should be called once at application startup to
// guarantee that monthly partitions exist for the upcoming months.
//
// monthsAhead controls how far into the future partitions are pre-created.
// A value of 3–6 is typical for daily scheduled services.
func EnsurePartitions(db *gorm.DB, monthsAhead int) error {
	if db == nil {
		slog.Warn("EnsurePartitions: database is nil, skipping")
		return nil
	}

	tables := []struct {
		name   string
		column string
	}{
		{"technical_analysis", "timestamp"},
		{"sentiment_analysis", "analyzed_at"},
	}

	for _, t := range tables {
		var created int
		err := db.Raw(
			"SELECT ensure_partitions(?, ?, ?)",
			t.name, t.column, monthsAhead,
		).Scan(&created).Error

		if err != nil {
			return fmt.Errorf("ensure_partitions(%s): %w", t.name, err)
		}
		if created > 0 {
			slog.Info("Created new partitions",
				"table", t.name,
				"count", created,
				"months_ahead", monthsAhead,
			)
		}
	}

	return nil
}
