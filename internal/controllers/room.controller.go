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
    var page = c.Query("page", "1")
    var limit = c.Query("limit", "10")

    intPage, _ := strconv.Atoi(page)
    intLimit, _ := strconv.Atoi(limit)
    offset := (intPage - 1) * intLimit

    var rooms []models.Room
    results := initializers.DB.Limit(intLimit).Offset(offset).Find(&rooms)
    if results.Error != nil {
        log.Printf("GetRooms: Database error: %v", results.Error)
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": results.Error})
    }

    log.Printf("GetRooms: Retrieved %d rooms", len(rooms))
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "results": len(rooms), "rooms": rooms})
}

// Создание комнаты
func CreateRoom(c *fiber.Ctx) error {
    var payload *models.Room

    if err := c.BodyParser(&payload); err != nil {
        log.Printf("CreateRoom: BodyParser error: %v", err)
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
    }

    payload.ID = uuid.New()
    result := initializers.DB.Create(&payload)

    if result.Error != nil {
        if strings.Contains(result.Error.Error(), "(SQLSTATE 23505)") {
            log.Printf("CreateRoom: Room already exists: %s", payload.Name)
            return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "Room already exists"})
        }
        log.Printf("CreateRoom: Database error: %v", result.Error)
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error.Error()})
    }

    log.Printf("CreateRoom: Room created successfully: %v", payload)
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"room": payload}})
}

// Удаление комнаты
func DeleteRoom(c *fiber.Ctx) error {
    roomId := c.Params("roomId")

    result := initializers.DB.Delete(&models.Room{}, "id = ?", roomId)

    if result.RowsAffected == 0 {
        log.Printf("DeleteRoom: No room found with ID: %s", roomId)
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "No room with that Id exists"})
    } else if result.Error != nil {
        log.Printf("DeleteRoom: Database error: %v", result.Error)
        return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "error", "message": result.Error})
    }

    log.Printf("DeleteRoom: Room deleted successfully with ID: %s", roomId)
    return c.SendStatus(fiber.StatusNoContent)
}

// Получение комнаты по ID
func GetRoomById(c *fiber.Ctx) error {
    roomId := c.Params("roomId")

    var room models.Room
    result := initializers.DB.First(&room, "id = ?", roomId)
    if result.Error != nil {
        log.Printf("GetRoomById: Room not found with ID: %s", roomId)
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Room not found"})
    }

    log.Printf("GetRoomById: Room found with ID: %s", roomId)
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": room})
}