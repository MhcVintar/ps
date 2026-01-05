package main

import (
	"fmt"
	"razpravljalnica/internal/client" 	
)

func main() {
	fmt.Println("bootstrap")

	if x := client.Bootstrap(); x != nil{
		fmt.Println("bootstrap")

	}
}

