package main

import (
	"log"

	"github.com/joho/godotenv"
	et "github.com/saiSunkari19/blockchain-projects-monorepo/experiments/sui/personal-signature-demo/backend/events-token-transfer"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Enable to Run Personal Message Sign
	// isValid := ab.MessageVerify()
	// fmt.Println(isValid)

	// Enable to Run Subscribe Events

	et.HandleSUISubscribeEvents()

}
