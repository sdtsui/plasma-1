package db

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/minimum-viable-plasma/chain"
	"github.com/kyokan/minimum-viable-plasma/util"
	"github.com/syndtr/goleveldb/leveldb"
	levelutil "github.com/syndtr/goleveldb/leveldb/util"
	"math/big"
)

const addrKeyPrefix = "addr"

type AddressDao interface {
	GetBalance(addr common.Address) (*big.Int, error)
	GetTransactionsWithUTXOs(addr common.Address) ([]chain.Transaction, error)
}

type LevelAddressDao struct {
	db *leveldb.DB
}

func (dao *LevelAddressDao) GetBalance(addr common.Address) (*big.Int, error) {
	prefix := addrPrefixKey(util.AddressToHex(&addr), "hash")
	iter := dao.db.NewIterator(levelutil.BytesPrefix(prefix), nil)

	total := big.NewInt(0)

	for iter.Next() {
		tx, err := chain.TransactionFromCbor(iter.Value())

		if err != nil {
			return nil, err
		}

		total = total.Add(total, extractAmount(tx, addr))
	}

	return total, nil
}

func (dao *LevelAddressDao) GetTransactionsWithUTXOs(addr common.Address) ([]chain.Transaction, error) {
	prefix := addrPrefixKey(util.AddressToHex(&addr), "hash")
	iter := dao.db.NewIterator(levelutil.BytesPrefix(prefix), nil)

	txMap := make(map[string]*chain.Transaction)

	for iter.Next() {
		tx, err := chain.TransactionFromCbor(iter.Value())

		if err != nil {
			return nil, err
		}

		if tx.IsDeposit() {
			txMap[utxoMapKey(tx.BlkNum, tx.TxIdx, 0)] = tx
			continue
		}

		if !tx.Input0.IsZeroInput() {
			txMap[utxoMapKey(tx.BlkNum, tx.TxIdx, 0)] = tx
		}

		if !tx.Input1.IsZeroInput() {
			txMap[utxoMapKey(tx.BlkNum, tx.TxIdx, 1)] = tx
		}
	}

	for _, tx := range txMap {
		if tx.IsDeposit() {
			continue
		}

		checkKey := utxoMapKey(tx.Input0.BlkNum, tx.Input0.TxIdx, tx.Input0.OutIdx)
		checkTx, exists := txMap[checkKey]

		if exists && checkTx != tx {
			delete(txMap, checkKey)
		}

		checkKey = utxoMapKey(tx.Input1.BlkNum, tx.Input1.TxIdx, tx.Input1.OutIdx)
		checkTx, exists = txMap[checkKey]

		if exists && checkTx != tx {
			delete(txMap, checkKey)
		}
	}

	var ret []chain.Transaction
	uniq := make(map[*chain.Transaction]bool)

	for _, tx := range txMap {
		_, exists := uniq[tx]

		if exists {
			continue
		}

		ret = append(ret, *tx)
		uniq[tx] = true
	}

	return ret, nil
}

func utxoMapKey(blkNum uint64, txIdx uint32, outIdx uint8) string {
	return fmt.Sprintf("%s::%s::%s", blkNum, txIdx, outIdx)
}

func extractAmount(tx *chain.Transaction, addr common.Address) *big.Int {
	if tx.IsDeposit() {
		return tx.Output0.Amount
	}

	amount := tx.Output0.Amount

	if !bytes.Equal(tx.Output0.NewOwner.Bytes(), addr.Bytes()) {
		amount = amount.Neg(amount)
	}

	if tx.Output1.IsZeroOutput() {
		return amount
	}

	if bytes.Equal(tx.Output1.NewOwner.Bytes(), addr.Bytes()) {
		amount = amount.Add(amount, tx.Output1.Amount)
	} else {
		amount = amount.Sub(amount, tx.Output1.Amount)
	}

	return amount
}

func addrPrefixKey(parts ...string) []byte {
	return prefixKey(addrKeyPrefix, parts...)
}
