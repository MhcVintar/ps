package main

import (
	"fmt"
	"razpravljalnica/internal/client" 	
)

func main() {
	fmt.Println("bootstrap")

	if x := client.RunGUI(); x != nil{
		fmt.Println("ERROR: An error has occured while running the TUI.")
	}
}

