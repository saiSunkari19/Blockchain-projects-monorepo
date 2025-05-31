package moveinteractions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/block-vision/sui-go-sdk/utils"
)

func HandleMoveCall() {

	var ctx = context.Background()
	var cli = sui.NewSuiClient(os.Getenv("TESTNET"))

	mnemonic := os.Getenv("MNEMONIC")
	signerAccount, err := signer.NewSignertWithMnemonic(mnemonic)
	if err != nil {
		fmt.Println(err.Error())
	}

	privKey := signerAccount.PriKey
	fmt.Printf("signerAccount.Address: %s\n", signerAccount.Address)

	gasObj := "0x4065a4fa52d47931ab4cb901db98a85abd52567392ca163470ad45d6e362a5d8"

	// // Get Balance
	// coin, err := cli.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
	// 	Owner:    signerAccount.Address,
	// 	CoinType: "0x02::sui::SUI",
	// })

	// if err != nil {
	// 	fmt.Println(err)
	// }

	// utils.PrettyPrint(coin)

	gasPrice, _ := cli.SuiXGetReferenceGasPrice(ctx)
	utils.PrettyPrint(gasPrice)

	rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAccount.Address,
		PackageObjectId: "0xa89c82b5ee25736b7494158924b971cfcc5db9c22045964c2d9d4ab20271e75b",
		Module:          "nft_simple",
		Function:        "mint_to_sender",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"Stable Difussion",
			"Description",
			"https://teal-perfect-tiglon-859.mypinata.cloud/ipfs/QmeX6KwHJjgLRqcQfqzpL26Abf9EuYaMUXm9EdDRgp49r1",
		},
		Gas:       &gasObj,
		GasBudget: "50000000000",
	})

	inspectResp, err := cli.SuiDryRunTransactionBlock(ctx, models.SuiDryRunTransactionBlockRequest{
		TxBytes: rsp.TxBytes,
	})

	utils.PrettyPrint(inspectResp)

	var gasInfo *GasBalanceInfo
	if err != nil {
		var dryRunError DryRunError
		err := json.Unmarshal([]byte(err.Error()), &dryRunError)
		if err != nil {
			fmt.Println("Error unmarshaling json: %v\n", err)
		}

		gasInfo, err = extractGasBalance(dryRunError.Message)
		if err != nil {
			fmt.Print("err", err)
		}
	}

	rsp, err = cli.MoveCall(ctx, models.MoveCallRequest{
		Signer:          signerAccount.Address,
		PackageObjectId: "0xa89c82b5ee25736b7494158924b971cfcc5db9c22045964c2d9d4ab20271e75b",
		Module:          "nft_simple",
		Function:        "mint_to_sender",
		TypeArguments:   []interface{}{},
		Arguments: []interface{}{
			"Sample New Test ",
			"ABCD",
			"https://123.com",
		},
		Gas:       &gasObj,
		GasBudget: strconv.Itoa(int(gasInfo.GasBalance)),
	})

	resp2, err := cli.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: rsp,
		PriKey:      privKey,

		Options: models.SuiTransactionBlockOptions{
			ShowInput:    true,
			ShowRawInput: true,
			ShowEffects:  true,
		},
		RequestType: "WaitForLocalExecution",
	})

	if err != nil {
		fmt.Println("2", err.Error())
	}

	utils.PrettyPrint(resp2)
}

// TODO : Improvise this
// > HACK: 	// errMsgJSON := `{"code":-32602,"message":"Error checking transaction input objects: GasBalanceTooLow { gas_balance: 58284480, needed_gas_amount: 50000000000 }"}`
type DryRunError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GasBalanceInfo struct {
	GasBalance      uint64
	NeededGasAmount uint64
}

func extractGasBalance(msg string) (*GasBalanceInfo, error) {
	// Regex to match the gas balance values
	re := regexp.MustCompile(`GasBalanceTooLow\s*\{\s*gas_balance:\s*(\d+),\s*needed_gas_amount:\s*(\d+)\s*\}`)
	matches := re.FindStringSubmatch(msg)

	if len(matches) != 3 {
		return nil, fmt.Errorf("failed to extract gas balance from message")
	}

	current, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid gas_balance value: %v", err)
	}

	needed, err := strconv.ParseUint(matches[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid needed_gas_amount value: %v", err)
	}

	return &GasBalanceInfo{
		GasBalance:      current,
		NeededGasAmount: needed,
	}, nil
}
