package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/kyokan/minimum-viable-plasma/chain"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"strconv"
	"github.com/kyokan/minimum-viable-plasma/util"
)

const txKeyPrefix = "tx"

type TransactionDao interface {
	Save(tx *chain.Transaction) error
	SaveMany(txs []chain.Transaction) error
	FindByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error)
}

type LevelTransactionDao struct {
	db *leveldb.DB
}

func (dao *LevelTransactionDao) Save(tx *chain.Transaction) error {
	return dao.SaveMany([]chain.Transaction{*tx})
}

func (dao *LevelTransactionDao) SaveMany(txs []chain.Transaction) error {
	batch := new(leveldb.Batch)

	for _, tx := range txs {
		err := dao.save(batch, &tx)

		if err != nil {
			return err
		}
	}

	return dao.db.Write(batch, nil)
}

func (dao *LevelTransactionDao) FindByBlockNumTxIdx(blkNum uint64, txIdx uint32) (*chain.Transaction, error) {
	key := blkNumTxIdxKey(blkNum, txIdx)
	exists, err := dao.db.Has(key, nil)

	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	gd := &GuardedDb{db: dao.db}
	data := gd.Get(key, nil)

	if gd.err != nil {
		return nil, gd.err
	}

	tx, err := chain.TransactionFromCbor(data)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (dao *LevelTransactionDao) save(batch *leveldb.Batch, tx *chain.Transaction) error {
	if tx.IsDeposit() {
		return saveDeposit(batch, tx)
	}

	cbor, err := tx.ToCbor()

	if err != nil {
		return err
	}

	inputTx1, err := dao.FindByBlockNumTxIdx(tx.Input0.BlkNum, tx.Input0.TxIdx)

	if err != nil {
		return err
	}

	prevOutput1 := getCorrectPreviousOutput(inputTx1, tx, 0)


	hash := tx.Hash()
	hexHash := common.ToHex(hash)
	hashKey := txPrefixKey("hash", hexHash)

	batch.Put(hashKey, cbor)
	batch.Put(addrPrefixKey(util.AddressToHex(&prevOutput1.NewOwner), "hash", hexHash), cbor)
	batch.Put(blkNumHashkey(tx.BlkNum, hexHash), cbor)
	batch.Put(blkNumTxIdxKey(tx.BlkNum, tx.TxIdx), cbor)

	if tx.Input1.IsZeroInput() {
		return nil
	}

	inputTx2, err := dao.FindByBlockNumTxIdx(tx.Input1.BlkNum, tx.Input1.TxIdx)

	if err != nil {
		return err
	}

	prevOutput2 := getCorrectPreviousOutput(inputTx2, tx, 1)
	batch.Put(addrPrefixKey(util.AddressToHex(&prevOutput2.NewOwner), "hash", hexHash), cbor)

	return nil
}

func saveDeposit(batch *leveldb.Batch, tx *chain.Transaction) error {
	cbor, err := tx.ToCbor()

	if err != nil {
		return err
	}

	hash := tx.Hash()
	hexHash := common.ToHex(hash)
	hashKey := txPrefixKey("hash", hexHash)

	batch.Put(hashKey, cbor)
	batch.Put(addrPrefixKey(util.AddressToHex(&tx.Output0.NewOwner), "hash", hexHash), cbor)
	batch.Put(blkNumHashkey(tx.BlkNum, hexHash), cbor)
	batch.Put(blkNumTxIdxKey(tx.BlkNum, tx.TxIdx), cbor)

	return nil
}

func getCorrectPreviousOutput(prevTx *chain.Transaction, thisTx *chain.Transaction, inputIdx uint8) *chain.Output {
	if inputIdx != 0 && inputIdx != 1 {
		log.Panicf("Invalid inputIdx: %d", inputIdx)
	}

	var outputIdx uint8

	if inputIdx == 0 {
		outputIdx = thisTx.Input0.OutIdx
	} else {
		outputIdx = thisTx.Input1.OutIdx
	}

	if outputIdx == 0 {
		return prevTx.Output0
	}

	return prevTx.Output1
}

func blkNumHashkey(blkNum uint64, hexHash string) []byte {
	return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "hash", hexHash)
}

func blkNumTxIdxKey(blkNum uint64, txIdx uint32) []byte {
	return txPrefixKey("blkNum", strconv.FormatUint(blkNum, 10), "txIdx", strconv.FormatUint(uint64(txIdx), 10))
}

func txPrefixKey(parts ...string) []byte {
	return prefixKey(txKeyPrefix, parts...)
}
