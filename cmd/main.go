package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/UnendingLoop/EventBooker/internal/cleaner"
	"github.com/UnendingLoop/EventBooker/internal/mwauthlog"
	"github.com/UnendingLoop/EventBooker/internal/repository"
	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/dbpg"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	// инициализировать конфиг/ считать энвы
	appConfig := config.New()
	appConfig.EnableEnv("")
	if err := appConfig.LoadEnvFiles("./.env"); err != nil {
		log.Fatalf("Failed to load envs: %s\nExiting app...", err)
	}

	// стартуем логгер - переделать в JWT-processor
	zlog.InitConsole()
	err := zlog.SetLevel("info")
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	// готовим заранее слушатель прерываний - контекст для всего приложения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// подключитсья к базе
	dbConn := repository.ConnectWithRetries(appConfig, 5, 10*time.Second)
	// накатываем миграцию
	repository.MigrateWithRetries(dbConn.Master, "./migrations", 10, 15*time.Second)

	// repo
	// jwt
	// service
	// handlers
	// конфиг сервера
	mode := appConfig.GetString("GIN_MODE")
	engine := ginext.New(mode)
	engine.Use(
		mwauthlog.RequestIDMiddleware()) // вставка уникального UID в каждый реквест

	events := engine.Group("/events", mwauthlog.RequireAuth([]byte(appConfig.GetString("SECRET"))))
	books := engine.Group("/books", mwauthlog.RequireAuth([]byte(appConfig.GetString("SECRET"))))
	auth := engine.Group("/auth")

	engine.GET("/ping")
	engine.Static("", "./internal/web/base") // UI админа/юзера - функциональность и контент зависит от роли

	auth.POST("/signup")                            // регистрация пользователя
	auth.POST("/login")                             // авторизация
	auth.Static("/signin", "./internal/web/signin") // страница с формой авторизации/регистрации

	events.POST("", mwauthlog.RequireRole("admin"))       // создание ивента - только админ
	events.GET("")                                        // список всех ивентов
	events.DELETE("/:id", mwauthlog.RequireRole("admin")) // удаление ивента

	books.POST("")             // создание бронирования
	books.POST("/:id/confirm") // подтверждение бронирования
	books.GET("/my")           // все брони по одному пользователю
	books.DELETE("/:id")       // отмена брони

	srv := &http.Server{
		Addr:    ":" + appConfig.GetString("APP_PORT"),
		Handler: engine,
	}

	// запуск сервера
	go func() {
		log.Printf("Server running on http://localhost%s\n", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				log.Println("Server gracefully stopping...")
			default:
				log.Printf("Server stopped: %v", err)
				stop()
			}
		}
	}()

	// cleaner
	clb := cleaner.NewBookCleaner(nil)
	clb.StartBookCleaner(ctx, 30)

	// слушаем контекст прерываний для запуска Graceful Shutdown
	<-ctx.Done()
	shutdown(dbConn, srv)
}

func shutdown(dbConn *dbpg.DB, srv *http.Server) {
	log.Println("Interrupt received! Starting shutdown sequence...")

	// Closing Server
	if err := srv.Close(); err != nil {
		log.Println("Failed to close server correctly:", err)
	} else {
		log.Println("Server is closed.")
	}

	// Closing DB connection
	if err := dbConn.Master.Close(); err != nil {
		log.Println("Failed to close DB-conn correctly:", err)
	} else {
		log.Println("DBconn is closed")
	}
}
