package cli

import (
	"strings"
)

// ParsedArgs represents parsed command line arguments
type ParsedArgs struct {
	Text  string
	Value string
	Flags map[string]bool
}

// ParseArgs parses command line arguments
func ParseArgs(args []string) ParsedArgs {
	flags := make(map[string]bool)
	var positional []string
	var value string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "-") {
			flag := strings.TrimLeft(arg, "-")
			flags[flag] = true

			// Check if next arg is a value for this flag
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				switch flag {
				case "c", "d", "e", "f", "fs", "n", "t":
					value = args[i+1]
					i++
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}

	// Join all positional args as the task text
	text := strings.Join(positional, " ")

	return ParsedArgs{
		Text:  text,
		Value: value,
		Flags: flags,
	}
}

// HasFlag checks if a flag is set
func (p *ParsedArgs) HasFlag(names ...string) bool {
	for _, name := range names {
		if p.Flags[name] {
			return true
		}
	}
	return false
}
