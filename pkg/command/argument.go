package command

import "fmt"

type Argument interface {
	Full() string
	Short() string
	Execute() error
	ReadLine(args []string) error
	String() string
}

type ValueArgument struct {
	full     string
	short    string
	desc     string
	required bool
	values   []string
	action   func([]string) error
}

func NewValueArgument(full, short, desc string) *ValueArgument {
	return &ValueArgument{
		full:   "--" + full,
		short:  "-" + short,
		desc:   desc,
		values: []string{},
	}
}

func (valueCommand *ValueArgument) Full() string {
	return valueCommand.full
}

func (valueCommand *ValueArgument) Short() string {
	return valueCommand.short
}

func (valueCommand *ValueArgument) Execute() error {
	if valueCommand.action != nil && len(valueCommand.values) > 0 {
		return valueCommand.action(valueCommand.values)
	}
	return nil
}

func (valueCommand *ValueArgument) ReadLine(args []string) error {
	if len(args) < 2 && valueCommand.required {
		return fmt.Errorf("argument '%s/%s' is required", valueCommand.full, valueCommand.short)
	}

	find := false

	for pos := 0; pos < len(args); pos++ {
		if args[pos] == valueCommand.full || args[pos] == valueCommand.short {
			find = true
			if pos+1 < len(args) {
				pos++
				valueCommand.values = append(valueCommand.values, args[pos])
			}
		}
	}

	if valueCommand.required {
		if !find {
			return fmt.Errorf("argument '%s/%s' is required", valueCommand.full, valueCommand.short)
		} else if len(valueCommand.values) == 0 {
			return fmt.Errorf("value of argument '%s/%s' is missing", valueCommand.full, valueCommand.short)
		}
	}

	return nil
}

func (valueCommand *ValueArgument) String() string {
	if valueCommand.required {
		return fmt.Sprintf("[Required] %s/%s    %s", valueCommand.full, valueCommand.short, valueCommand.desc)
	}
	return fmt.Sprintf("[Optional] %s/%s    %s", valueCommand.full, valueCommand.short, valueCommand.desc)
}

func (valueCommand *ValueArgument) Required() *ValueArgument {
	valueCommand.required = true
	return valueCommand
}

func (valueCommand *ValueArgument) Action(action func([]string) error) *ValueArgument {
	valueCommand.action = action
	return valueCommand
}

type MarkArgument struct {
	full    string
	short   string
	desc    string
	present bool
	action  func() error
}

func NewMarkArgument(full, short, desc string) *MarkArgument {
	return &MarkArgument{
		full:  "--" + full,
		short: "-" + short,
		desc:  desc,
	}
}

func (markCommand *MarkArgument) Full() string {
	return markCommand.full
}

func (markCommand *MarkArgument) Short() string {
	return markCommand.short
}

func (markCommand *MarkArgument) Execute() error {
	if markCommand.action != nil && markCommand.present {
		return markCommand.action()
	}
	return nil
}

func (markCommand *MarkArgument) ReadLine(args []string) error {
	for pos := 0; pos < len(args); pos++ {
		if args[pos] == markCommand.full || args[pos] == markCommand.short {
			markCommand.present = true
			break
		}
	}
	return nil
}

func (markCommand *MarkArgument) String() string {
	return fmt.Sprintf("[Optional] %s/%s    %s", markCommand.full, markCommand.short, markCommand.desc)
}

func (markCommand *MarkArgument) Present() bool {
	return markCommand.present
}

func (markCommand *MarkArgument) Action(action func() error) *MarkArgument {
	markCommand.action = action
	return markCommand
}
