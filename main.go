package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"

	"go_back/internal/controllers"
	"go_back/internal/initializers"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalln("Failed to load environment variables! \n", err.Error())
	}

	initializers.ConnectDB(&config)
}

func main() {
	host := flag.String("host", "0.0.0.0", "host to listen on")
	port := flag.String("port", "8000", "port to listen on")
	flag.Parse()

	app := fiber.New()

	app.Use(logger.New(logger.Config{
		Format:     "[${time}] ${ip} ${method} ${status}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "https://kursovaya-ryzhov.ru, http://localhost:8081",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))
	

	api := app.Group("/api")

	api.Route("/users", func(router fiber.Router) {
		router.Post("/login", controllers.LoginUser)
		router.Post("/register", controllers.CreateUser)
		router.Post("/reset-password", controllers.ResetPassword)
		router.Delete("/:userId", controllers.DeleteUser)
		router.Get("/", controllers.FindUsers)
		router.Get("/:userId", controllers.FindUserById)
	})

	api.Route("/rooms", func(router fiber.Router) {
		router.Get("/", controllers.GetRooms)
		router.Post("/", controllers.CreateRoom)
		router.Delete("/:roomId", controllers.DeleteRoom)
		router.Get("/:roomId", controllers.GetRoomById)
	})

	api.Route("/tasks", func(router fiber.Router) {
		router.Get("/", controllers.GetTasks)
		router.Post("/", controllers.CreateTask)
		router.Delete("/:taskId", controllers.DeleteTask)
	})

	api.Route("/entries", func(router fiber.Router) {
		router.Get("/", controllers.GetEntries)
		router.Post("/", controllers.CreateEntry)
		router.Delete("/:entryId", controllers.DeleteEntry)
	})

	api.Route("/columns", func(router fiber.Router) {
		router.Get("/", controllers.GetColumns)
		router.Post("/", controllers.CreateColumn)
	})

	api.Route("/email", func(router fiber.Router) {
		router.Post("/send-code", controllers.SendVerificationCode)
		router.Post("/verify-code", controllers.VerifyCode)
	})

	api.Route("/friends", func(router fiber.Router) {
		router.Post("/request", controllers.SendFriendRequest)   // отправить заявку
		router.Post("/accept", controllers.AcceptFriendRequest) // принять заявку
		router.Get("/", controllers.GetFriendsList)             // список друзей
		router.Get("/requests", controllers.GetIncomingRequests) // входящие заявки
	})
	

	addr := fmt.Sprintf("%s:%s", *host, *port)
	log.Printf("Server is running on %s\n", addr)
	log.Fatal(app.Listen(addr))
}
