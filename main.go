package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	store := FileStore{path: "session.json"}
	client, err := NewClient(&store)
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	if !client.LoggedIn() {
		reader := bufio.NewScanner(os.Stdin)

		fmt.Print("Enter your username: ")
		reader.Scan()
		username := reader.Text()

		fmt.Print("Enter your password: ")
		reader.Scan()
		password := reader.Text()

		if err := client.Login(username, password); err != nil {
			fmt.Println("An error occurred:", err)
			return
		}
		fmt.Println("Logged in successfully")
	} else {
		if err := client.Refresh(); err != nil {
			fmt.Println("An error occurred:", err)
			return
		}
		fmt.Println("Refreshed successfully")
	}
}
