
package controllers

import (
    "github.com/gofiber/fiber/v2"
    "go_back/internal/models"
    "go_back/internal/initializers"
	"github.com/google/uuid"
    "log"

	
)

func SendFriendRequest(c *fiber.Ctx) error {
    type Request struct {
		From       uuid.UUID `json:"from"`       
		ToUsername string    `json:"toUsername"` 
	}
	
	

    var body Request
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Неверный запрос"})
    }

    var toUser models.User
    if err := initializers.DB.Where("username = ?", body.ToUsername).First(&toUser).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Пользователь не найден"})
    }

    if body.From == toUser.ID {
        return c.Status(400).JSON(fiber.Map{"error": "Нельзя добавить себя в друзья"})
    }

    friendship := models.Friendship{
        UserID:   body.From,
        FriendID: toUser.ID,
        Status:   "pending",
    }

    if err := initializers.DB.Create(&friendship).Error; err != nil {
        log.Println("❌ Ошибка создания заявки:", err)
        return c.Status(500).JSON(fiber.Map{"error": "Ошибка при создании заявки"})
    }

    return c.JSON(fiber.Map{"message": "Заявка отправлена"})
}

func AcceptFriendRequest(c *fiber.Ctx) error {
    type AcceptRequest struct {
		From     uuid.UUID `json:"from"`
		FriendID uuid.UUID `json:"friendId"`
	}
	

    var body AcceptRequest
    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Неверный запрос"})
    }

    var friendship models.Friendship
    if err := initializers.DB.Where("user_id = ? AND friend_id = ? AND status = ?", body.FriendID, body.From, "pending").
        First(&friendship).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Заявка не найдена"})
    }

    friendship.Status = "accepted"
    if err := initializers.DB.Save(&friendship).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Ошибка при обновлении статуса"})
    }

    return c.JSON(fiber.Map{"message": "Заявка принята"})
}

func GetFriendsList(c *fiber.Ctx) error {
    userIDStr := c.Query("userId")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Неверный формат UUID",
        })
    }

    var friends []models.User

    err = initializers.DB.
        Table("users").
        Joins("JOIN friendships f ON (f.friend_id = users.id OR f.user_id = users.id)").
        Where("f.status = ?", "accepted").
        Where("f.user_id = ? OR f.friend_id = ?", userID, userID).
        Where("users.id != ?", userID).
        Find(&friends).Error // 🔥 вот ключевое отличие

    if err != nil {
        log.Println("❌ Ошибка получения друзей:", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Не удалось получить друзей",
        })
    }

    return c.JSON(friends)
}


func GetIncomingRequests(c *fiber.Ctx) error {
    userIDStr := c.Query("userId")

    // Преобразование строки в UUID
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Неверный формат userId (ожидается UUID)",
        })
    }

    var requests []models.Friendship
    if err := initializers.DB.
        Where("friend_id = ? AND status = ?", userID, "pending").
        Preload("User").
        Find(&requests).Error; err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Не удалось получить заявки",
        })
    }

    return c.JSON(requests)
}

