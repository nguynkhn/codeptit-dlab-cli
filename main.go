package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	client := NewClient()
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

	if err := client.Refresh(); err != nil {
		fmt.Println("An error occurred:", err)
		return
	}
	fmt.Println("Refreshed successfully")
}
