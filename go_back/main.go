package main

import (
    "log"
    "os"
    "flag"
    "fmt"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/cors"

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
        AllowOrigins: "*", // Разрешаем все источники (лучше указать конкретные)
        AllowMethods: "GET,POST,PUT,DELETE",
        AllowHeaders: "Origin, Content-Type, Accept",
    }))

    api := app.Group("/api")

    api.Route("/users", func(router fiber.Router) {
        router.Post("/login", controllers.LoginUser)
        router.Post("/register", controllers.CreateUser)
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

    addr := fmt.Sprintf("%s:%s", *host, *port)
    log.Printf("Server is running on %s\n", addr)
    log.Fatal(app.Listen(addr))
}