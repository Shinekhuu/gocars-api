package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
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
