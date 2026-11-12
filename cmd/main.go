package main

import (
	"errors"
	"io/fs"
	"log"
	"net/http"

	appcmd "github.com/ImaSerix/go-message-service/internal/application/command"
	appqry "github.com/ImaSerix/go-message-service/internal/application/query"
	"github.com/ImaSerix/go-message-service/internal/asyncrunner"
	"github.com/ImaSerix/go-message-service/internal/config"
	"github.com/ImaSerix/go-message-service/internal/generator"
	"github.com/ImaSerix/go-message-service/internal/interface/handler/auth"
	chatHandler "github.com/ImaSerix/go-message-service/internal/interface/handler/chat"
	messageHandler "github.com/ImaSerix/go-message-service/internal/interface/handler/message"
	userHandler "github.com/ImaSerix/go-message-service/internal/interface/handler/user"
	wsHandler "github.com/ImaSerix/go-message-service/internal/interface/handler/ws"
	wsHub "github.com/ImaSerix/go-message-service/internal/interface/ws"
	"github.com/ImaSerix/go-message-service/internal/repository/postgres/command"
	"github.com/ImaSerix/go-message-service/internal/repository/postgres/query"
	"github.com/ImaSerix/go-message-service/internal/storage"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Fatal("Error loading .env file")
	}

	config := config.Load()
	db, err := storage.NewDBConnection(config.DB.FormatDSN())
	if err != nil {
		log.Fatalf("error while connecting to db: %v", err)
	}

	usersRepository := command.NewUsersRepository(db)
	chatsRepository := command.NewChatsRepository(db)
	messagesRepository := command.NewMessagesRepository(db)

	hub := wsHub.New(config.WSHub.ReadLimit, config.WSHub.PongWait, config.WSHub.PingPerion, config.WSHub.WriteWait)
	idGenerator := generator.NewSnowflakeIDGenerator(config.IDGenerator.NodeID, config.IDGenerator.Epoch)
	runner := asyncrunner.New()

	commandService := appcmd.NewCommandService(
		chatsRepository,
		messagesRepository,
		usersRepository,
		hub,
		idGenerator,
		runner,
	)
	queryService := appqry.NewQueryService(query.NewChatsQry(db), query.NewMessageQry(db), config.MaxMessagePageSize)

	autheMiddleware := auth.AutheMiddleware(usersRepository)
	authoMiddleware := messageHandler.AuthoMiddleware(chatsRepository)
	urlParamMiddleware := chatHandler.UrlParamMiddleware()
	paramMiddleware := messageHandler.GetParamMiddleware()

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

		r.Route("/{chatID}", func(r chi.Router) {
			r.Use(urlParamMiddleware)
			r.Use(authoMiddleware)

			r.Get("/", chatsHandler.GetChat)

			r.Route("/messages", func(r chi.Router) {
				r.Use(paramMiddleware)
				r.Get("/", messagesHandler.Get)
				r.Post("/", messagesHandler.Send)
			})

			r.Post("/read", messagesHandler.Read)
		})
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
