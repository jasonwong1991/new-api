package model

import (
	"fmt"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	originalDB := DB
	originalUsingSQLite := common.UsingSQLite
	originalUsingMySQL := common.UsingMySQL
	originalUsingPostgreSQL := common.UsingPostgreSQL

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	initCol()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	DB = db

	t.Cleanup(func() {
		DB = originalDB
		common.UsingSQLite = originalUsingSQLite
		common.UsingMySQL = originalUsingMySQL
		common.UsingPostgreSQL = originalUsingPostgreSQL
		initCol()
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})

	return db
}

func TestMigrateDBCreatesFAQBoardPostTable(t *testing.T) {
	db := setupMigrationTestDB(t)

	if err := migrateDB(); err != nil {
		t.Fatalf("migrateDB failed: %v", err)
	}

	if !db.Migrator().HasTable(&FAQBoardPost{}) {
		t.Fatalf("expected faq_board_posts table to be created")
	}
}
