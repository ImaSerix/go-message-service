package command_test

import "github.com/ImaSerix/go-message-service/internal/application/command"

func NewTestService() (*command.CommandService, *fakeChatsRepo, *fakeMessagesRepo, *fakeUsersRepo, *fakeNotifier, *fakeIDGenerator) {
	fChats := fakeChatsRepo{}
	fMessages := fakeMessagesRepo{}
	fUsers := fakeUsersRepo{}
	fNotifier := fakeNotifier{}
	fGenerator := fakeIDGenerator{}
	runner := syncRunner{}

	cmdServ := command.NewCommandService(&fChats, &fMessages, &fUsers, &fNotifier, &fGenerator, &runner)
	return cmdServ, &fChats, &fMessages, &fUsers, &fNotifier, &fGenerator
}
