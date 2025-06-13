package main

import (
	"log"

	"github.com/joho/godotenv"
	mv "github.com/saiSunkari19/blockchain-projects-monorepo/experiments/sui/dapp-integrations/backend/twitter-apis"
	// mv "github.com/saiSunkari19/blockchain-projects-monorepo/experiments/sui/dapp-integrations/backend/move-interactions"
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

	// et.HandleSUISubscribeEvents()

	// mv.UpdateAccessList()
	mv.RunTwitterAPI()

	// mv.HandleMoveCall()
	// mv.EnableMintFee()
	// mv.HandleMoveMintNFT()
	// mv.HandleAirdrop()
	// mv.HandleMergeCoins()
	// mv.QueryObject()

}
