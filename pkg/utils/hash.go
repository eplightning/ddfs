package utils

import (
	"encoding/hex"
)

type BlockHash struct {
	Bytes []byte

	stringRepr string
}

func NewBlockHash(bytes []byte) *BlockHash {
	return &BlockHash{
		Bytes: bytes,
	}
}

func (h *BlockHash) String() string {
	if h.stringRepr != "" {
		return h.stringRepr
	}
	if len(h.Bytes) == 0 {
		return ""
	}

	h.stringRepr = hex.EncodeToString(h.Bytes)
	return h.stringRepr
}
