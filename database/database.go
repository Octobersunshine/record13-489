package database

import (
	"log"
	"redpacket/config"
	"redpacket/models"

	_ "modernc.org/sqlite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(sqlite.Dialector{
		DSN:                       cfg.DBPath,
		DriverName:                "sqlite",
	}, &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	err = DB.AutoMigrate(
		&models.RedPacketActivity{},
		&models.RedPacketRecord{},
		&models.UserRedPacket{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")
}
