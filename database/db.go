package database

import (
	"fmt"
	"gocars-api/config"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg config.Config) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DB_USER,
		cfg.DB_PASSWORD,
		cfg.DB_HOST,
		cfg.DB_PORT,
		cfg.DB_NAME,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 🔥 IMPORTANT: configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get DB instance:", err)
	}

	// ✅ limits
	sqlDB.SetMaxOpenConns(10)                 // max connections
	sqlDB.SetMaxIdleConns(5)                  // idle connections
	sqlDB.SetConnMaxLifetime(time.Hour)       // max lifetime
	sqlDB.SetConnMaxIdleTime(5 * time.Minute) // idle timeout

	log.Println("Database initialized successfully")
}
