package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go_back/internal/models"
	"go_back/internal/initializers"
)

// Получение задач по комнате
func GetTasks(c *fiber.Ctx) error {
	roomID := c.Query("roomId")
	var tasks []models.Task
	initializers.DB.Where("room_id = ?", roomID).Find(&tasks)
	return c.JSON(tasks)
}

// Создание задачи
func CreateTask(c *fiber.Ctx) error {
	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	task.ID = uuid.New()
	initializers.DB.Create(&task)
	return c.JSON(task)
}

// Удаление задачи
func DeleteTask(c *fiber.Ctx) error {
	id := c.Params("taskId")
	initializers.DB.Delete(&models.Task{}, "id = ?", id)
	return c.JSON(fiber.Map{"message": "Task deleted"})
}
