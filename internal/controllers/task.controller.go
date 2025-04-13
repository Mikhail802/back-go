package controllers

import (
	"encoding/json"
    "gorm.io/datatypes"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go_back/internal/initializers"
	"go_back/internal/models"
	"log"
)

// Получение задач по комнате с пользователями
func GetTasks(c *fiber.Ctx) error {
	roomID := c.Query("roomId")
	var tasks []models.Task

	if err := initializers.DB.Preload("AssignedUsers.User").Where("room_id = ?", roomID).Find(&tasks).Error; err != nil {
		log.Printf("GetTasks: Ошибка получения задач: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при получении задач"})
	}

	return c.JSON(tasks)
}

// Создание задачи — только для admin/owner
func CreateTask(c *fiber.Ctx) error {
	log.Println("🚨 [DEBUG] CreateTask запущен")
	user := c.Locals("user").(*models.User)

	var task models.Task
	if err := c.BodyParser(&task); err != nil {
		log.Printf("❌ BodyParser error: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	log.Printf("📥 Task payload: RoomID=%s, ColumnID=%s, Text=%s", task.RoomID, task.ColumnID, task.Text)

	// Проверка, существует ли колонка
	var column models.Column
	if err := initializers.DB.First(&column, "id = ?", task.ColumnID).Error; err != nil {
		log.Printf("❌ Column not found: %v", err)
		return c.Status(400).JSON(fiber.Map{"error": "Column not found"})
	}

	log.Printf("✅ Column найден: ID=%s, привязан к RoomID=%s", column.ID, column.RoomID)

	// Сопоставление room_id у колонки и переданного
	if column.RoomID != task.RoomID {
		log.Printf("❌ Column-room mismatch: column.RoomID = %s, task.RoomID = %s", column.RoomID, task.RoomID)
		return c.Status(400).JSON(fiber.Map{"error": "Column does not belong to room"})
	}

	// Проверка, есть ли права у пользователя
	var member models.RoomMember
	if err := initializers.DB.
		Where("room_id = ? AND user_id = ?", task.RoomID, user.ID).
		First(&member).Error; err != nil {
		log.Printf("❌ RoomMember not found: user.ID = %s, task.RoomID = %s", user.ID, task.RoomID)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к комнате"})
	}

	log.Printf("✅ RoomMember найден: роль = %s", member.Role)

	if member.Role != "admin" && member.Role != "owner" {
		log.Printf("❌ Недостаточно прав для создания задачи. Роль: %s", member.Role)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Недостаточно прав"})
	}

	// Создаём задачу
	task.ID = uuid.New()
	if err := initializers.DB.Create(&task).Error; err != nil {
		log.Printf("❌ Ошибка создания задачи: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при создании задачи"})
	}

	// Загружаем назначенных пользователей (сейчас будет пусто, но для единого формата)
	initializers.DB.Preload("AssignedUsers.User").First(&task, "id = ?", task.ID)

	log.Printf("✅ Задача успешно создана: %+v", task)
	return c.JSON(task)
}

// Удаление задачи — только для owner/admin или назначенного
func DeleteTask(c *fiber.Ctx) error {
	taskId := c.Params("taskId")
	user := c.Locals("user").(*models.User)

	var task models.Task
	if err := initializers.DB.First(&task, "id = ?", taskId).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Задача не найдена"})
	}

	// Проверим назначение задачи
	var assignment models.TaskAssignment
	initializers.DB.Where("task_id = ?", taskId).First(&assignment)

	// Проверим роль
	var member models.RoomMember
	initializers.DB.Where("room_id = ? AND user_id = ?", task.RoomID, user.ID).First(&member)

	// Условия: участник должен быть либо админом/владельцем, либо назначен на задачу
	if member.Role != "admin" && member.Role != "owner" && assignment.UserID != user.ID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к удалению задачи"})
	}

	if err := initializers.DB.Delete(&models.Task{}, "id = ?", taskId).Error; err != nil {
		log.Printf("DeleteTask: Ошибка удаления задачи: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Ошибка при удалении задачи"})
	}

	return c.JSON(fiber.Map{"message": "Задача удалена"})
}

// Получение задачи по ID с пользователями
func GetTaskById(c *fiber.Ctx) error {
	taskId := c.Params("taskId")

	var task models.Task
	err := initializers.DB.
		Preload("AssignedUsers.User").
		First(&task, "id = ?", taskId).
		Error

	if err != nil {
		log.Printf("Ошибка получения задачи: %v", err)
		return c.Status(404).JSON(fiber.Map{"error": "Задача не найдена"})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   task,
	})
}

func UpdateTask(c *fiber.Ctx) error {
	taskId := c.Params("taskId")

	var payload struct {
		Text        string                `json:"text"`
		Description string                `json:"description"`
		StartDate   *string               `json:"startDate"`
		EndDate     *string               `json:"endDate"`
		TaskLists   []models.TaskListData `json:"taskLists"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат запроса",
		})
	}

	var task models.Task
	if err := initializers.DB.First(&task, "id = ?", taskId).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Задача не найдена",
		})
	}

	jsonData, err := json.Marshal(payload.TaskLists)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Ошибка сериализации чеклистов",
		})
	}

	task.Text = payload.Text
	task.Description = payload.Description
	task.StartDate = payload.StartDate
	task.EndDate = payload.EndDate
	task.TaskLists = datatypes.JSON(jsonData)

	if err := initializers.DB.Save(&task).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при обновлении задачи",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"task":   task,
	})
}

