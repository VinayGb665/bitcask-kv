package main

import (
	"bufio"
	"regexp"
	"strconv"
	"time"

	// BStorage "github.com/vinaygb665/bitcask-kv/bitcask_storage"
	"fmt"
	"os"
	"strings"

	Utils "github.com/vinaygb665/bitcask-kv/utils"

	Bitcask "github.com/vinaygb665/bitcask-kv/bitcask"

	"github.com/abiosoft/ishell"
)

const MAX_RESPONSES_PER_PAGE = 30

func CommandHandler() {
	s := Bitcask.Storage{}
	s.Init("/tmp/bitcask", false, 1024*1024*1024)
	// Do nothing
	fmt.Println("Welcome to Bitcask-KV")
	var response string
	for {
		inputCommand := Prompt()
		start := time.Now()
		allowedCommands := []string{"get", "set", "exit", "scan"}

		if len(inputCommand) > 0 && Utils.Contains(inputCommand[0], allowedCommands) {
			switch inputCommand[0] {
			case "get":
				response = GetCommandHandler(inputCommand, &s)
			case "set":
				response = SetCommandHandler(inputCommand, &s)
			case "scan":
				response = ScanCommandHandler(inputCommand, &s)
			case "exit":
				fmt.Println("Bye")
				return
			}
		} else {
			response = "Invalid command"
		}
		fmt.Println(response)
		fmt.Println("Time taken: ", time.Since(start))
	}

}
func Prompt() []string {
	// Do nothing
	fmt.Print("> ")
	in := bufio.NewReader(os.Stdin)

	line, _ := in.ReadString('\n')
	resp := strings.Fields(line)
	// fmt.Println(resp)
	return resp
}

func GetCommandHandler(args []string, client *Bitcask.Storage) string {
	// Check if the command is valid, len(args) == 2
	// Check if the key is valid, len(args[1]) > 0
	if len(args) != 1 || len(args[0]) == 0 {
		return "Usage get <key>"
	}
	// Do nothing
	readVal, success := client.Read(args[0])
	if success {
		return string(readVal)
	}
	return "nil"

}

func SetCommandHandler(args []string, client *Bitcask.Storage) string {
	// Check if the command is valid, len(args) == 3
	// Check if the key is valid, len(args[1]) > 0
	// Check if the value is valid, len(args[2]) > 0
	if len(args) != 2 || len(args[0]) == 0 || len(args[1]) == 0 {
		return "Usage set <key> <value>"
	}
	// Do nothing
	err := client.Write(args[0], []byte(args[1]))
	if err != nil {
		return "Error writing to storage"
	}
	return "Ok"
}

func ScanCommandHandler(args []string, client *Bitcask.Storage) string {
	// Check if the command is valid, len(inputCommand) == 2
	// Check if the key is valid, len(inputCommand[1]) > 0

	if len(args) != 1 || len(args[0]) == 0 {
		return "Usage scan <key/pattern>"
	}
	var response string
	count := 0

	// Do nothing
	if args[0] == "*" {
		for key, _ := range client.Keymap {
			response += key + "\n"
			count++
			if count == MAX_RESPONSES_PER_PAGE {
				break
			}
		}
		fmt.Println("Total keys: %d", len(client.Keymap))
		response += "Total keys: " + strconv.Itoa(len(client.Keymap))
		return response
	} else {
		regexp, err := regexp.Compile(args[0])
		if err != nil {
			return "Invalid regex"
		}
		for key, _ := range client.Keymap {

			if regexp.MatchString(key) {
				response += key + "\n"
			}
			count++
			if count == MAX_RESPONSES_PER_PAGE {
				break
			}
		}
		return response
	}

}

func main() {
	// Do nothing
	shell := ishell.New()
	storage := Bitcask.Storage{}
	storage.Init("/tmp/bitcask", false, 1024*1024*1024)
	shell.Println("Sample Interactive Shell")
	// CommandHandler()
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "get <key>",
		Func: func(c *ishell.Context) {
			// c.Println("Hello", strings.Join(c.Args, " "))
			response := GetCommandHandler(c.Args, &storage)
			c.Println(response)

		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set <key> <value>",
		Func: func(c *ishell.Context) {
			// c.Println("Hello", strings.Join(c.Args, " "))
			response := SetCommandHandler(c.Args, &storage)
			c.Println(response)

		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "scan",
		Help: "scan <key/pattern>",
		Func: func(c *ishell.Context) {
			// c.Println("Hello", strings.Join(c.Args, " "))
			response := ScanCommandHandler(c.Args, &storage)
			c.Println(response)

		},
	})
	shell.Run()
}
