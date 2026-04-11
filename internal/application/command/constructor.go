package command

func NewCommandService(chats ChatsRepository, messages MessagesRepository, users UsersRepository, notifier Notifier, generator IDGenerator) *CommandService {
	return &CommandService{
		chats:     chats,
		messages:  messages,
		users:     users,
		notifier:  notifier,
		generator: generator,
	}
}
