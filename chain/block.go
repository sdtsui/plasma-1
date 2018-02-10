package chain

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/keybase/go-codec/codec"
	"github.com/kyokan/minimum-viable-plasma/util"
)

type BlockHeader struct {
	MerkleRoot util.Hash
	PrevHash   util.Hash
	Number     uint64
}

type Block struct {
	Header    *BlockHeader
	BlockHash util.Hash
}

func BlockFromCbor(data []byte) (*Block, error) {
	hdl := new(codec.CborHandle)
	dec := codec.NewDecoderBytes(data, hdl)
	ptr := &Block{}
	err := dec.Decode(ptr)

	if err != nil {
		return nil, err
	}

	return ptr, nil
}

func (blk Block) ToCbor() ([]byte, error) {
	buf := new(bytes.Buffer)
	bw := bufio.NewWriter(buf)
	hdl := new(codec.CborHandle)
	enc := codec.NewEncoder(bw, hdl)
	err := enc.Encode(blk)

	if err != nil {
		return nil, err
	}

	bw.Flush()

	return buf.Bytes(), nil
}

func (head BlockHeader) Hash() util.Hash {
	buf := new(bytes.Buffer)
	buf.Write(head.MerkleRoot)
	buf.Write(head.PrevHash)
	binary.Write(buf, binary.BigEndian, head.Number)
	digest := sha3.Sum256(buf.Bytes())
	return digest[:]
}
