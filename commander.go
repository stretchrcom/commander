package commander

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

// Default is used to register a default command that will be run when no
// arguments are given.
//
// The handler for the default command will be passed a nil map as no arguments
// are present.
const DefaultCommand = ""

// Commander provides methods and functionality to create a command line
// interface quickly and easily.
type Commander struct {
	// commands contains all the mapped commands
	commands []*command

	// defaultRegistered stores whether a default has been registered or not
	defaultRegistered bool
}

// initOnce is used to guarantee that the sharedCommander is initialized only once.
var initOnce sync.Once

// sharedCommander is the shared instance of the Commander type
var sharedCommander *Commander

// incomingArgs is the array of arguments to be analyzed. This exists to facilitate
// testing.
var incomingArgs []string

// commandMap builds a map of indentifier,value to be passed to the handler
func commandMap(cmd *command, args []string) map[string]interface{} {
	argMap := make(map[string]interface{})
	for i, a := range cmd.arguments {
		if len(args) <= i {
			break
		}
		if !a.isLiteral() {
			if !a.isVariable() {
				argMap[a.identifier] = args[i]
			} else {
				if len(cmd.arguments) == len(args) {
					argMap[a.identifier] = args[i]
				} else {
					argMap[a.identifier] = args[i:]
				}
			}
		}
	}
	return argMap
}

// printUsage prints the usage of the program
func printUsage(cmd *command) {

	appName := path.Base(os.Args[0])
	if extension := path.Ext(os.Args[0]); extension != "" {
		appName = strings.Replace(appName, extension, "", 1)
	}

	if cmd == nil {
		fmt.Printf("usage: %s <command> [arguments]\n\n", appName)
		for _, cmd := range sharedCommander.commands {
			fmt.Printf("\t %s\n", cmd.definition)
		}
	} else {
		fmt.Printf("Not enough arguments to command \"%s\". Usage:\n", cmd.arguments[0].literal)
		fmt.Printf("\t %s\n", cmd.definition)
		//TODO: this
		//fmt.Printf("\t %s\n", cmd.description)
	}

}

// Initialize sets up various internal fields to ready the system. If this is not
// called, Commander will not function.
func Initialize() {
	initOnce.Do(func() {
		sharedCommander = new(Commander)
		Map("help [arg=(string)]", func(map[string]interface{}) {
			printUsage(nil)
		})
	})
}

// Map is used to map a definition string to a handler function. If the arguments
// given on the command line are represented by the definition string, the
// handler function will be called.
func Map(definition string, handler Handler) {

	if sharedCommander == nil {
		panic("Initialize must be called before Map")
	}

	if definition == DefaultCommand {
		if sharedCommander.defaultRegistered {
			panic("Only one default command can be registered.")
		} else {
			sharedCommander.defaultRegistered = true
		}
	}

	newCommand := makeCommand(definition, handler)

	for _, cmd := range sharedCommander.commands {
		if cmd.isEqualTo(newCommand) {
			panic("Each command must have a unique signature.")
		}
	}

	sharedCommander.commands = append(sharedCommander.commands, newCommand)

}

// TODO: add matchCount or similar to determine which command has the closest match, then print the usage for just that command

// Execute analyzes the arguments given to the program and executes the
// appropriate command handler function
func Execute() {

	executeDefault := false
	executed := false
	closestMatchCount := 0
	var closestMatch *command

	if len(incomingArgs) == 0 {
		incomingArgs = os.Args
		if len(incomingArgs) == 1 {
			executeDefault = true
		}
	}

	if executeDefault {
		for _, cmd := range sharedCommander.commands {
			if cmd.isDefaultCommand() {
				cmd.handler(nil)
				executed = true
			}
		}
	} else {
		for _, cmd := range sharedCommander.commands {
			if represents, matchCount := cmd.represents(incomingArgs); represents {
				args := commandMap(cmd, incomingArgs)
				cmd.handler(args)
				executed = true
			} else {
				if matchCount > closestMatchCount {
					closestMatchCount = matchCount
					closestMatch = cmd
				}
			}
		}
	}
	if !executed {
		printUsage(closestMatch)
	}

}
