package main

import (
	"fmt"
	"log"

	"github.com/iqhater/pkg/generate"
)

// Random token generate example
func main() {

	tokenLength := uint(32)

	token, err := generate.GenerateRandomToken(tokenLength)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Generated Random Token:", token)
}
