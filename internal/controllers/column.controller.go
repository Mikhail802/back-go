package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go_back/internal/models"
	"go_back/internal/initializers"
)

// Получить все колонки в комнате
func GetColumns(c *fiber.Ctx) error {
	roomID := c.Query("roomId")
	var columns []models.Column
	initializers.DB.Preload("Tasks").Where("room_id = ?", roomID).Find(&columns)
	return c.JSON(columns)
}

// Создать колонку
func CreateColumn(c *fiber.Ctx) error {
	var column models.Column
	if err := c.BodyParser(&column); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	column.ID = uuid.New()
	initializers.DB.Create(&column)
	return c.JSON(column)
}

func DeleteColumn(c *fiber.Ctx) error {
	columnIDStr := c.Params("columnId")

	// Проверка UUID
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Некорректный UUID колонки"})
	}

	// Проверка существования
	var column models.Column
	if err := initializers.DB.First(&column, "id = ?", columnID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Колонка не найдена"})
	}

	// Удаление (каскадное удаление задач сработает автоматически)
	if err := initializers.DB.Delete(&column).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при удалении колонки"})
	}

	return c.JSON(fiber.Map{"message": "Колонка успешно удалена"})
}

func UpdateColumn(c *fiber.Ctx) error {
	columnIDStr := c.Params("columnId")
	columnID, err := uuid.Parse(columnIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Некорректный UUID колонки"})
	}

	var updateData struct {
		Title string `json:"title"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Неверный формат данных"})
	}

	var column models.Column
	if err := initializers.DB.First(&column, "id = ?", columnID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Колонка не найдена"})
	}

	column.Title = updateData.Title

	if err := initializers.DB.Save(&column).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при обновлении колонки"})
	}

	return c.JSON(fiber.Map{"message": "Колонка обновлена", "column": column})
}
