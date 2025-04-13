package middleware

import (
	"go_back/internal/initializers"
	"go_back/internal/models"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RoomRoleGuard(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRaw := c.Locals("user")
		user, ok := userRaw.(*models.User)
		if !ok || user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неавторизован"})
		}

		// 1. Пробуем взять из query
		roomIdStr := c.Query("roomId")
		if roomIdStr == "" {
			roomIdStr = c.Params("roomId")
		}

		// 2. Если всё ещё пусто — пробуем взять из JSON
		if roomIdStr == "" {
			var body struct {
				RoomID string `json:"room_id"`
			}
			if err := c.BodyParser(&body); err == nil && body.RoomID != "" {
				roomIdStr = body.RoomID
			}
		}

		// 3. Попытка найти roomId по columnId (если roomId всё ещё не найден)
		if roomIdStr == "" {
			columnIdStr := c.Params("columnId")
			if columnIdStr != "" {
				columnID, err := uuid.Parse(columnIdStr)
				if err == nil {
					var column models.Column
					if err := initializers.DB.First(&column, "id = ?", columnID).Error; err == nil {
						roomIdStr = column.RoomID.String()
					}
				}
			}
		}
		
		// 4. Если всё равно нет — ошибка
		if roomIdStr == "" {
			log.Println("❌ RoomRoleGuard: roomId отсутствует в запросе и не найден по columnId")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "roomId не передан"})
		}


		roomID, err := uuid.Parse(roomIdStr)
		if err != nil {
			log.Printf("❌ RoomRoleGuard: Некорректный roomId %s", roomIdStr)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Некорректный roomId"})
		}

		var membership models.RoomMember
		err = initializers.DB.
			Where("room_id = ? AND user_id = ?", roomID, user.ID).
			First(&membership).Error
		if err != nil {
			log.Printf("❌ RoomRoleGuard: не найден участник комнаты %v", err)
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Нет доступа к комнате"})
		}

		userRole := strings.ToLower(membership.Role)
		for _, r := range requiredRoles {
			if userRole == r {
				return c.Next()
			}
		}

		log.Printf("❌ RoomRoleGuard: роль %s не входит в %v", userRole, requiredRoles)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Недостаточно прав"})
	}
}

