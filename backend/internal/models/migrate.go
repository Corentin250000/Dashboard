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

// SeedWidgetSizes inserts the fixed dashboard sizes required by widget
// instances. Existing rows are left unchanged so this is safe on startup.
func SeedWidgetSizes(db *gorm.DB) error {
	sizes := []WidgetSize{
		{Code: "small", Width: 1, Height: 1},
		{Code: "wide", Width: 2, Height: 1},
		{Code: "tall", Width: 1, Height: 2},
		{Code: "big", Width: 2, Height: 2},
	}
	for _, size := range sizes {
		if err := db.Where("code = ?", size.Code).FirstOrCreate(&size).Error; err != nil {
			return err
		}
	}
	return nil
}
