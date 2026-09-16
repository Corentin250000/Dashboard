package models

import (
	"gorm.io/gorm"
)

/*
  AutoMigrate creates/updates tables from All().
  It is a dev-convenience helper only: db/migrations/*.sql stays the source
  of truth for reviews (CHECK constraints, indexes, triggers), and production
  schema changes must go through versioned SQL migrations.
*/
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(All()...)
}
