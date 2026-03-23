package command

import (
	"fmt"
	"testing"
)

func prepare() *Command {
	userArgument := NewValueArgument("user", "u", "specific user screen name").Action(func(values []string) error {
		fmt.Printf("user: %v\n", values)
		return nil
	})
	pathArgument := NewValueArgument("path", "p", "specific save path").Action(func(values []string) error {
		fmt.Printf("path: %v\n", values)
		return nil
	})
	deltaArgument := NewMarkArgument("delta", "t", "enable delta mode").Action(func() error {
		fmt.Println("delta mode enabled")
		return nil
	})
	downloadArgument := NewValueArgument("downloader", "d", "specific downloader number").Action(func(values []string) error {
		fmt.Printf("downloader: %v\n", values)
		return nil
	})

	return InitializeCommand().Register(userArgument).Register(pathArgument).Register(deltaArgument).Register(downloadArgument).EnableHelp()
}

func TestCommand_ReadLine(t *testing.T) {
	args := []string{"./app", "-u", "welsper", "-p", "/path/to/save_dir", "--delta", "--downloader", "6"}

	cmd := prepare()

	if err := cmd.ReadLine(args); err != nil {
		t.Fatalf("command read line error: %v", err)
	}
}

func TestCommand_Execute(t *testing.T) {
	args := []string{"./app", "-u", "welsper", "-p", "/path/to/save_dir", "--delta", "--downloader", "6"}

	cmd := prepare()

	if err := cmd.ReadLine(args); err != nil {
		t.Fatalf("command read line error: %v", err)
	}

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command execute error: %v", err)
	}
}
