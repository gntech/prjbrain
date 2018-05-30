package main

import "fmt"
import "github.com/icza/gowut/gwu"

func main() {
	fmt.Println("hello")
	win := gwu.NewWindow("main", "Test GUI Window")

	// Create and start a GUI server (omitting error check)
	server := gwu.NewServer("guitest", "localhost:8081")
	server.SetText("Test GUI App")
	server.AddWin(win)
	server.Start("") // Also opens windows list in browser

}
