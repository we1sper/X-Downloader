package command

import (
	"fmt"
	"os"
)

type Command struct {
	arguments []Argument
}

func InitializeCommand() *Command {
	return &Command{
		arguments: []Argument{},
	}
}

func (cmd *Command) Register(argument Argument) *Command {
	for _, arg := range cmd.arguments {
		if arg.Full() == argument.Full() {
			panic(fmt.Sprintf("argument with full name '%s' already exists", arg.Full()))
		}
		if arg.Short() == argument.Short() {
			panic(fmt.Sprintf("argument with short name '%s' already exists", arg.Short()))
		}
	}
	cmd.arguments = append(cmd.arguments, argument)
	return cmd
}

func (cmd *Command) ReadLine(args []string) error {
	for _, argument := range cmd.arguments {
		if err := argument.ReadLine(args); err != nil {
			return err
		}
	}
	return nil
}

func (cmd *Command) Execute() error {
	for _, argument := range cmd.arguments {
		if err := argument.Execute(); err != nil {
			return fmt.Errorf("execute action of argument '%s/%s' failed: %v", argument.Full(), argument.Short(), err)
		}
	}
	return nil
}

func (cmd *Command) Pipeline() {
	if len(os.Args) < 2 {
		fmt.Println(cmd.usage())
		return
	}
	if err := cmd.ReadLine(os.Args); err != nil {
		cmd.errorHandler(err)
	}
	if err := cmd.Execute(); err != nil {
		cmd.errorHandler(err)
	}
}

func (cmd *Command) EnableHelp() *Command {
	help := NewMarkArgument("help", "h", "show usages").Action(func() error {
		fmt.Println(cmd.usage())
		return nil
	})
	return cmd.Register(help)
}

func (cmd *Command) usage() string {
	usage := "Usage:\n"
	for _, argument := range cmd.arguments {
		usage += fmt.Sprintf("    %s\n", argument)
	}
	return usage
}

func (cmd *Command) errorHandler(err error) {
	fmt.Printf("[ERROR] %v\n", err)
	fmt.Println(cmd.usage())
}
