package main

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var subCommands = map[string]func(c *Client){
	"login":  LoginCommand,
	"course": CourseCommand,
	"submit": SubmitCommand,
}

func main() {
	if len(os.Args) < 2 {
		prog := filepath.Base(os.Args[0])
		listCmds := strings.Join(slices.Collect(maps.Keys(subCommands)), "|")
		fmt.Println("Usage:", prog, listCmds)
		return
	}

	subCmd := strings.ToLower(os.Args[1])
	executor, ok := subCommands[subCmd]
	if !ok {
		fmt.Println("Unknown command")
		return
	}

	client, err := NewClient(&FileStore{path: "session.json"})
	if err != nil {
		fmt.Println("Failed to initialize client:", err)
		return
	}

	if subCmd != "login" && !client.LoggedIn() {
		fmt.Println("Please log in first")
		return
	}

	if client.SignKey() != nil {
		if err := client.Refresh(); err != nil {
			fmt.Println("Failed to refresh:", err)
			fmt.Println("Please try logging in again")
			return
		}
	}

	executor(client)
}
