package util

import (
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

func AddressToHex(addr *common.Address) string {
	return strings.ToLower(addr.Hex())
}

