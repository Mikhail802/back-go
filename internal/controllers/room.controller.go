package controllers

import (
    "go_back/internal/initializers"
    "go_back/internal/models"
    "log"
    "strconv"
    "strings"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
)

// Получение списка комнат с пагинацией
func GetRooms(c *fiber.Ctx) error {
	page := c.Query("page", "1")
	limit := c.Query("limit", "10")
	ownerId := c.Query("ownerId")
	userId := c.Query("userId")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var rooms []models.Room

	if userId != "" {
		var roomIDs []uuid.UUID
		var ownedRooms, memberRooms []models.Room

		// Владелец
		if err := initializers.DB.Preload("Members").
			Where("owner_id = ?", userId).
			Find(&ownedRooms).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении своих комнат"})
		}

		// Участник
		if err := initializers.DB.Model(&models.RoomMember{}).
			Select("room_id").
			Where("user_id = ?", userId).
			Pluck("room_id", &roomIDs).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении участий"})
		}

		if len(roomIDs) > 0 {
			if err := initializers.DB.Preload("Members").
				Where("id IN ?", roomIDs).
				Find(&memberRooms).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении участий"})
			}
		}

		roomMap := map[string]models.Room{}
		for _, r := range ownedRooms {
			roomMap[r.ID.String()] = r
		}
		for _, r := range memberRooms {
			roomMap[r.ID.String()] = r
		}
		for _, r := range roomMap {
			rooms = append(rooms, r)
		}
	} else if ownerId != "" {
		initializers.DB.Preload("Members").
			Where("owner_id = ?", ownerId).
			Limit(intLimit).
			Offset(offset).
			Find(&rooms)
	} else {
		initializers.DB.Preload("Members").
			Limit(intLimit).
			Offset(offset).
			Find(&rooms)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"results": len(rooms),
		"rooms":   rooms,
	})
}


// Создание комнаты
func CreateRoom(c *fiber.Ctx) error {
    var payload *models.Room

    if err := c.BodyParser(&payload); err != nil {
        log.Printf("CreateRoom: BodyParser error: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "fail",
            "message": err.Error(),
        })
    }

    if payload.Name == "" || payload.Theme == "" || payload.OwnerID == uuid.Nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  "fail",
            "message": "Поля name, theme и owner_id обязательны",
        })
    }

    payload.ID = uuid.New()

    // Сохраняем комнату
    if err := initializers.DB.Create(&payload).Error; err != nil {
        if strings.Contains(err.Error(), "(SQLSTATE 23505)") {
            return c.Status(fiber.StatusConflict).JSON(fiber.Map{
                "status":  "fail",
                "message": "Комната с таким названием уже существует",
            })
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": err.Error(),
        })
    }

    // 🔥 Добавляем владельца как участника с ролью owner
    member := models.RoomMember{
        ID:     uuid.New(),
        RoomID: payload.ID,
        UserID: payload.OwnerID,
        Role:   "owner",
    }

    if err := initializers.DB.Create(&member).Error; err != nil {
        log.Printf("CreateRoom: Не удалось добавить владельца как участника: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status":  "error",
            "message": "Комната создана, но не удалось добавить владельца как участника",
        })
    }

    log.Printf("CreateRoom: Room и членство владельца успешно созданы")
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "status": "success",
        "data": fiber.Map{
            "room": payload,
        },
    })
}



// Удаление комнаты
func DeleteRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	result := initializers.DB.Delete(&models.Room{}, "id = ?", roomId)
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Комната не найдена"})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Получение комнаты по ID
func GetRoomById(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	var room models.Room
	err := initializers.DB.
		Preload("Members.User").
		First(&room, "id = ?", roomId).Error

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": "Комната не найдена",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   room,
	})
}


func InviteToRoom(c *fiber.Ctx) error {
	var payload struct {
		RoomID     uuid.UUID `json:"roomId"`
		ToUsername string    `json:"toUsername"`
		Role       string    `json:"role"` // "member" по умолчанию
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	if strings.TrimSpace(payload.ToUsername) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Имя пользователя не может быть пустым"})
	}

	var toUser models.User
	if err := initializers.DB.First(&toUser, "username = ?", payload.ToUsername).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Пользователь не найден"})
	}

	// Проверка: уже в комнате?
	var existingMember models.RoomMember
	if err := initializers.DB.
		Where("room_id = ? AND user_id = ?", payload.RoomID, toUser.ID).
		First(&existingMember).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Пользователь уже в комнате"})
	}

	// Проверка: уже есть приглашение?
	var existingInvite models.RoomInvite
	if err := initializers.DB.
		Where("room_id = ? AND to_user_id = ? AND status = ?", payload.RoomID, toUser.ID, "pending").
		First(&existingInvite).Error; err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Пользователь уже приглашён"})
	}

	// Получение отправителя из middleware
	fromUserRaw := c.Locals("user")
	fromUser, ok := fromUserRaw.(*models.User)
	if !ok || fromUser == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Неавторизован"})
	}

	// Создание приглашения
	invite := models.RoomInvite{
		ID:         uuid.New(),
		RoomID:     payload.RoomID,
		FromUserID: fromUser.ID,
		ToUserID:   toUser.ID,
		Status:     "pending",
	}

	if err := initializers.DB.Create(&invite).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при создании приглашения"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "ok"})
}



func GetRoomMembers(c *fiber.Ctx) error {
	roomId := c.Params("roomId")

	var members []models.RoomMember
	if err := initializers.DB.Where("room_id = ?", roomId).Find(&members).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении участников"})
	}

	return c.JSON(fiber.Map{"members": members})
}

func AssignRoomRole(c *fiber.Ctx) error {
	var payload struct {
		RoomID uuid.UUID `json:"roomId"`
		UserID uuid.UUID `json:"userId"`
		Role   string    `json:"role"` // "admin" или "member"
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	if payload.Role != "admin" && payload.Role != "member" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Недопустимая роль"})
	}

	result := initializers.DB.Model(&models.RoomMember{}).
		Where("room_id = ? AND user_id = ?", payload.RoomID, payload.UserID).
		Update("role", payload.Role)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при обновлении роли"})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Участник не найден"})
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

// Контроллер для назначения участников на задачу

func UpdateTaskAssignment(c *fiber.Ctx) error {
	var payload struct {
		TaskID  uuid.UUID   `json:"taskId"`
		UserIDs []uuid.UUID `json:"userIds"` // Список пользователей, которых нужно оставить
	}

	// Получаем данные из тела запроса
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверные данные"})
	}

	// Проверка существования задачи
	var task models.Task
	if err := initializers.DB.First(&task, "id = ?", payload.TaskID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Задача не найдена"})
	}

	// Если список пользователей пуст, удаляем всех пользователей из task_assignments
	if len(payload.UserIDs) == 0 {
		if err := initializers.DB.Where("task_id = ?", payload.TaskID).Delete(&models.TaskAssignment{}).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при удалении всех пользователей"})
		}
		return c.JSON(fiber.Map{"status": "all users removed"})
	}

	// Удаляем пользователей, которых нет в новом списке
	if err := initializers.DB.Where("task_id = ? AND user_id NOT IN (?)", payload.TaskID, payload.UserIDs).Delete(&models.TaskAssignment{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при удалении участников"})
	}

	// Добавляем или обновляем пользователей, которых нужно оставить
	for _, userID := range payload.UserIDs {
		// Если пользователь еще не назначен на задачу, создаем новое назначение
		var existingAssignment models.TaskAssignment
		if err := initializers.DB.Where("task_id = ? AND user_id = ?", payload.TaskID, userID).First(&existingAssignment).Error; err != nil {
			assignment := models.TaskAssignment{
				ID:     uuid.New(),
				TaskID: payload.TaskID,
				UserID: userID,
			}
			if err := initializers.DB.Create(&assignment).Error; err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при назначении пользователя"})
			}
		}
	}

	return c.JSON(fiber.Map{"status": "users updated successfully"})
}






func GetRoomInvites(c *fiber.Ctx) error {
	userId := c.Query("userId")
	if userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "userId обязателен"})
	}

	var invites []models.RoomInvite
	if err := initializers.DB.Preload("Room").Preload("From").
		Where("to_user_id = ? AND status = ?", userId, "pending").
		Find(&invites).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при получении приглашений"})
	}

	var response []fiber.Map
	for _, inv := range invites {
		response = append(response, fiber.Map{
			"id":         inv.ID,
			"roomName":   inv.Room.Name,
			"inviterName": inv.From.Name,
		})
	}

	return c.JSON(fiber.Map{"invites": response})
}


func AcceptRoomInvite(c *fiber.Ctx) error {
	var body struct {
		InviteID uuid.UUID `json:"inviteId"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный формат"})
	}

	var invite models.RoomInvite
	if err := initializers.DB.First(&invite, "id = ?", body.InviteID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Приглашение не найдено"})
	}

	invite.Status = "accepted"
	if err := initializers.DB.Save(&invite).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при принятии приглашения"})
	}

	// Добавляем пользователя в комнату
	member := models.RoomMember{
		ID:     uuid.New(),
		RoomID: invite.RoomID,
		UserID: invite.ToUserID,
		Role:   "member",
	}
	initializers.DB.Create(&member)

	return c.JSON(fiber.Map{"status": "ok"})
}


func RejectRoomInvite(c *fiber.Ctx) error {
	var body struct {
		InviteID uuid.UUID `json:"inviteId"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Неверный формат"})
	}

	if err := initializers.DB.Model(&models.RoomInvite{}).
		Where("id = ?", body.InviteID).
		Update("status", "rejected").Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Ошибка при отклонении"})
	}

	return c.JSON(fiber.Map{"status": "отклонено"})
}

func RemoveUserFromRoom(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	userId := c.Params("userId")

	if roomId == "" || userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "roomId и userId обязательны",
		})
	}

	result := initializers.DB.Where("room_id = ? AND user_id = ?", roomId, userId).Delete(&models.RoomMember{})
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка при удалении пользователя из комнаты",
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Пользователь не найден в комнате",
		})
	}

	return c.JSON(fiber.Map{"status": "удалено"})
}
