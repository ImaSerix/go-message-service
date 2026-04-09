package main

import (
	"log"
	"net/http"

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
	chatUseCase "github.com/ImaSerix/go-message-service/internal/usecase/chat"
	messageUseCase "github.com/ImaSerix/go-message-service/internal/usecase/message"
	userUseCase "github.com/ImaSerix/go-message-service/internal/usecase/user"
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
	messagesRepository := postgres.NewMessagesRepository(db)

	wsHub := wsHub.New(config.WSHub.ReadLimit, config.WSHub.PongWait, config.WSHub.PingPerion, config.WSHub.WriteWait)
	notifier := notifier.New(chatsRepository, wsHub)

	idGenerator := generator.NewSnowflakeIDGenerator(config.IDGenerator.NodeID, config.IDGenerator.Epoch)

	usersUseCase := userUseCase.New(usersRepository)
	chatsUseCase := chatUseCase.New(chatsRepository, messagesRepository, usersRepository, notifier)
	messagesUseCase := messageUseCase.New(messagesRepository, config.MaxMessagePageSize, chatsRepository, notifier, idGenerator)

	autheMiddleware := auth.AutheMiddleware(usersRepository)
	authoMiddleware := messageHandler.AuthoMiddleware(chatsRepository)
	paramMiddleware := messageHandler.ParamMiddleware()

	usersHandler := userHandler.New(usersUseCase)
	chatsHandler := chatHandler.New(chatsUseCase)
	messagesHandler := messageHandler.New(messagesUseCase)
	wshandler := wsHandler.New(wsHub)

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

		r.Route("{chatID}/messages", func(r chi.Router) {
			r.Use(paramMiddleware)
			r.Use(authoMiddleware)

			r.Get("/", messagesHandler.Get)
			r.Post("/", messagesHandler.Send)
		})

		r.Route("{chatID}/read", func(r chi.Router) {
			r.Use(paramMiddleware)
			r.Use(authoMiddleware)

			r.Post("/", messagesHandler.Read)
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
