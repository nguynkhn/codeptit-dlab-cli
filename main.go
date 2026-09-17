package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	client := NewClient()
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	fmt.Print("Enter your password: ")
	password, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	if err := client.Login(username, password); err != nil {
		fmt.Println("An error occurred:", err)
		return
	}
	fmt.Println("Logged in successfully")

	if err := client.Refresh(); err != nil {
		fmt.Println("An error occurred:", err)
		return
	}
	fmt.Println("Refreshed successfully")
}
