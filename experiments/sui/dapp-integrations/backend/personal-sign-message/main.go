package personalsignmessage

import (
	"fmt"
	"os"
	"strings"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
)

func MessageVerify() bool {

	mnemonic := os.Getenv("MNEMONIC")
	signerAccount, err := signer.NewSignertWithMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	message := "Sign this message to login:\nAddress: 0xac463879634ef8be8c0db3c718a2ac9d08a1db45d48ceed3021eff656d1ce8ee\nNonce: fab4396b-70b2-48c8-be9d-d8527933e48c"
	signature, err := signerAccount.SignPersonalMessageV1(message)

	if err != nil {
		fmt.Printf("%s", err.Error())
		return false
	}

	signer, isValid, err := models.VerifyPersonalMessage(message, signature.Signature)
	if err != nil {
		fmt.Printf("%s", err.Error())
		return false
	}

	signer = "0xAc463879634Ef8be8c0db3c718a2ac9d08a1db45d48ceed3021eff656d1ce8ee"
	if !AddressEqual(signer, signerAccount.Address) {
		fmt.Printf("both strings are not equal")
		return false
	}

	fmt.Println("Signature", signature.Signature)
	fmt.Println("Message", message)
	fmt.Println("Address: ", signer)
	fmt.Println("isValid: ", isValid)

	return true

}

func AddressEqual(a, b string) bool {
	return strings.EqualFold(a, b)
}
