package chain

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/keybase/go-codec/codec"
	"github.com/kyokan/minimum-viable-plasma/util"
	"math/big"
)

type Transaction struct {
	Input0  *Input
	Input1  *Input
	Sig0    []byte
	Sig1    []byte
	Output0 *Output
	Output1 *Output
	Fee     *big.Int
	BlkNum  uint64
	TxIdx   uint32
}

func TransactionFromCbor(data []byte) (*Transaction, error) {
	hdl := util.PatchedCBORHandle()
	dec := codec.NewDecoderBytes(data, hdl)
	ptr := &Transaction{}
	err := dec.Decode(ptr)

	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func (tx *Transaction) IsDeposit() bool {
	return tx.Input0.IsZeroInput() &&
		tx.Input1.IsZeroInput() &&
		!tx.Output0.IsZeroOutput() &&
		tx.Output1.IsZeroOutput()
}

func (tx *Transaction) GetUTXO(addr *common.Address) *Output {
	if tx.IsDeposit() {
		return tx.Output0
	}

	if tx.Output0.IsUTXO {
		return tx.Output0
	}

	if tx.Output1.IsUTXO {
		return tx.Output1
	}

	panic("expected to find a UTXO")
}

func (tx *Transaction) GetUTXOIndex(addr *common.Address) uint8 {
	if tx.IsDeposit() {
		return 0
	}

	if tx.Output0.IsUTXO {
		return 0
	}

	if tx.Output1.IsUTXO {
		return 1
	}

	panic("expected to find a UTXO")
}

func (tx *Transaction) ToCbor() ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)
	hdl := util.PatchedCBORHandle()
	enc := codec.NewEncoder(bw, hdl)
	err := enc.Encode(tx)

	if err != nil {
		return nil, err
	}

	bw.Flush()

	return buf.Bytes(), nil
}

func (tx *Transaction) Hash() util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Sig0,
		tx.Input1.Hash(),
		tx.Sig1,
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
		tx.BlkNum,
		tx.TxIdx,
	}

	return doHash(values)
}

func (tx *Transaction) SignatureHash() util.Hash {
	values := []interface{}{
		tx.Input0.Hash(),
		tx.Input1.Hash(),
		tx.Output0.Hash(),
		tx.Output1.Hash(),
		tx.Fee,
	}

	return doHash(values)
}

func doHash(values []interface{}) util.Hash {
	buf := new(bytes.Buffer)

	for _, component := range values {
		var err error
		switch t := component.(type) {
		case util.Hash:
			_, err = buf.Write(t)
		case []byte:
			_, err = buf.Write(t)
		case *big.Int:
			_, err = buf.Write(t.Bytes())
		case uint64, uint32:
			err = binary.Write(buf, binary.BigEndian, t)
		default:
			err = errors.New("invalid component type")
		}

		if err != nil {
			panic(err)
		}
	}

	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
