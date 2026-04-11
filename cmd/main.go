package main

import (
	"log"
	"net/http"

	appcmd "github.com/ImaSerix/go-message-service/internal/application/command"
	appqry "github.com/ImaSerix/go-message-service/internal/application/query"
	"github.com/ImaSerix/go-message-service/internal/config"
	"github.com/ImaSerix/go-message-service/internal/generator"
	"github.com/ImaSerix/go-message-service/internal/handler/auth"
	chatHandler "github.com/ImaSerix/go-message-service/internal/handler/chat"
	messageHandler "github.com/ImaSerix/go-message-service/internal/handler/message"
	userHandler "github.com/ImaSerix/go-message-service/internal/handler/user"
	wsHandler "github.com/ImaSerix/go-message-service/internal/handler/ws"
	"github.com/ImaSerix/go-message-service/internal/notifier"
	"github.com/ImaSerix/go-message-service/internal/repository/postgres"
	"github.com/ImaSerix/go-message-service/internal/storage"
	wsHub "github.com/ImaSerix/go-message-service/internal/ws"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	config := config.Load()
	db, err := storage.NewDBConnection(config.DB.FormatDSN())
	if err != nil {
		log.Fatalf("error while connecting to db: %v", err)
	}

	usersRepository := postgres.NewUsersRepository(db)
	chatsRepository := postgres.NewChatsRepository(db)

	hub := wsHub.New(config.WSHub.ReadLimit, config.WSHub.PongWait, config.WSHub.PingPerion, config.WSHub.WriteWait)
	notifier := notifier.New(chatsRepository, hub)
	idGenerator := generator.NewSnowflakeIDGenerator(config.IDGenerator.NodeID, config.IDGenerator.Epoch)

	commandService := appcmd.NewCommandService(
		appcmd.NewPGChatsRepository(db),
		appcmd.NewPGMessagesRepository(db),
		appcmd.NewPGUsersRepository(db),
		notifier,
		idGenerator,
	)
	queryService := appqry.NewQueryService(db, config.MaxMessagePageSize)

	autheMiddleware := auth.AutheMiddleware(usersRepository)
	authoMiddleware := messageHandler.AuthoMiddleware(chatsRepository)
	paramMiddleware := messageHandler.ParamMiddleware()

	usersHandler := userHandler.New(commandService)
	chatsHandler := chatHandler.New(commandService, queryService)
	messagesHandler := messageHandler.New(commandService, queryService)
	wshandler := wsHandler.New(hub)

	r := chi.NewRouter()

	r.Post("/internal/users", usersHandler.SaveLocal)
	r.Route("/ws", func(r chi.Router) {
		r.Use(autheMiddleware)
		r.Post("/", wshandler.ServeHTTP)
	})

	r.Route("/chats", func(r chi.Router) {
		r.Use(autheMiddleware)
		r.Get("/", chatsHandler.GetChats)
		r.Post("/", chatsHandler.CreateChat)

		r.Route("/{chatID}/messages", func(r chi.Router) {
			r.Use(paramMiddleware)
			r.Use(authoMiddleware)
			r.Get("/", messagesHandler.Get)
			r.Post("/", messagesHandler.Send)
		})

		r.Route("/{chatID}/read", func(r chi.Router) {
			r.Use(paramMiddleware)
			r.Use(authoMiddleware)
			r.Post("/", messagesHandler.Read)
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
