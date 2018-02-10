package rpc

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/minimum-viable-plasma/chain"
	"github.com/kyokan/minimum-viable-plasma/node"
	"log"
	"math/big"
	"net/http"
)

type SendArgs struct {
	From   string
	To     string
	Amount string
}

type SendResponse struct {
	Transaction *chain.Transaction
}

type TransactionService struct {
	TxChan chan<- chan node.TransactionRequest
}

func (t *TransactionService) Send(r *http.Request, args *SendArgs, reply *SendResponse) error {
	log.Printf("Received Transaction.Send request.")

	from := common.HexToAddress(args.From)
	to := common.HexToAddress(args.To)
	amount := new(big.Int)
	amount.SetString(args.Amount, 0)

	req := node.TransactionRequest{
		From:   from,
		To:     to,
		Amount: amount,
	}

	ch := make(chan node.TransactionRequest)
	t.TxChan <- ch
	ch <- req
	res := <-ch
	close(ch)

	if res.Response.Error != nil {
		return res.Response.Error
	}

	*reply = SendResponse{
		Transaction: res.Response.Transaction,
	}

	return nil
}
