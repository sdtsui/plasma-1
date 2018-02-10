package plasma

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/minimum-viable-plasma/db"
	"github.com/urfave/cli"
	"log"
)

func PrintUTXOs(c *cli.Context) {
	level, err := db.CreateLevelDatabase(c.GlobalString("db"))

	if err != nil {
		log.Panic("Failed to establish connection with database:", err)
	}

	addrStr := c.String("addr")

	if addrStr == "" {
		log.Panic("Addr is required.")
	}

	addr := common.HexToAddress(c.String("addr"))
	utxos, err := level.AddressDao.GetTransactionsWithUTXOs(addr)

	if err != nil {
		log.Panic("Failed to get UTXOs: ", err)
	}

	if len(utxos) == 0 {
		log.Printf("No UTXOs found for address %s.", addrStr)
	}

	for _, utxo := range utxos {
		log.Printf("--------------------")
		log.Printf("Hash: %s", common.ToHex(utxo.Hash()))
		log.Printf("Amount: %s", utxo.GetUTXO(&addr).Amount.String())
		log.Printf("In Block: %d", utxo.BlkNum)
		log.Printf("At Index: %d", utxo.TxIdx)
	}
}
