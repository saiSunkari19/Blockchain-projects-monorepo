package eventstokentransfer

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"

	"context"
)

func HandleSUISubscribeEvents() {

	ctx := context.Background()

	if ctx.Err() != nil {
		fmt.Println("Clean shutdown", ctx.Err().Error())
	}

	var nextCursor string = ""
	// Pooling Interval
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
			fmt.Println("Pooling for SUI Transfers")
			nextCursor = pollIncomingSUITransfers(ctx, nextCursor)

		}

	}

}

func pollIncomingSUITransfers(ctx context.Context, cursor string) string {
	cli := sui.NewSuiClient(os.Getenv("HTTP_URL"))
	batchSize := uint(50)

	options := models.SuiTransactionBlockOptions{
		ShowEffects:        true,
		ShowBalanceChanges: true,
		ShowInput:          true,
	}

	nextCursor := cursor
	for {

		req := models.SuiXQueryTransactionBlocksRequest{
			SuiTransactionBlockResponseQuery: models.SuiTransactionBlockResponseQuery{
				Options: options,
				TransactionFilter: map[string]interface{}{
					"ToAddress": "",
				},
			},
			Limit:           uint64(batchSize),
			Cursor:          nil,
			DescendingOrder: true,
		}

		resp, err := cli.SuiXQueryTransactionBlocks(ctx, req)
		if err != nil {
			fmt.Println("Error Querying transaction Blocks", err)
			break
		}

		if len(resp.Data) == 0 {
			break
		}

		for _, tx := range resp.Data {

			sender := tx.Transaction.Data.Sender
			if tx.BalanceChanges == nil {
				continue
			}
			for _, bc := range tx.BalanceChanges {
				data := bc.GetBalanceChangeOwner()
				amount, err := strconv.ParseInt(bc.Amount, 10, 64)

				if err != nil {
					fmt.Println("error while parsing error", err.Error())
				}

				if strings.EqualFold(data, "") && bc.CoinType == "0x2::sui::SUI" && amount > 0 {

					fmt.Printf("✅ Sender: %s, Agent wallet received %d SUI in tx %s\n", sender, amount, tx.Digest)

				}

			}
		}

		if resp.HasNextPage && resp.NextCursor != "" {
			nextCursor = resp.NextCursor
		} else {
			break
		}

	}
	return nextCursor

}
