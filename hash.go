package main

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcutil/base58"
	"path"
)

func DecodeHash(hash string) string {
	return hex.EncodeToString(base58.Decode(hash))
}

func FormatRelativeZapFilePathFromHash(hash string) string {
	return path.Join(hash[:2], hash[2:4], hash[4:])
}
