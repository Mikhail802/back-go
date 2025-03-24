package initializers

import (
	"fmt"
	"log"
	

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go_back/internal/models"
)

var DB *gorm.DB

func ConnectDB(config *Config) {
	dsn := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		config.PSQLUser, config.PSQLPassword, config.PSQLHost, config.PSQLPort, config.PSQLDbName)
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// Если соединение не удалось, выводим ошибку и завершаем программу
		log.Fatal("DB connection failed: ", err.Error())
	}
	// Если соединение успешно, выводим сообщение
	log.Println("DB connection successful")

	// Настроим логирование для GORM с уровнем "Info"
	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Println("Running migrations...")
	DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";")
	// Выполняем миграцию моделей в базе данных
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Room{},
		&models.Column{},
		&models.Task{},
		&models.Entry{},
		
	); err != nil {
		log.Fatal("Migrations failed: ", err.Error())
	}

	log.Println("Migrations completed successfully.")
}