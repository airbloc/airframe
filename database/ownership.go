package database

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/airbloc/logger"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/json-iterator/go"
	"golang.org/x/crypto/sha3"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func IsOwner(obj *Object, sig []byte) bool {
	hash := GetObjectHash(obj.Type, obj.ID, obj.Data)

	pub, err := crypto.SigToPub(hash[:], sig)
	if err != nil {
		return false
	}
	recovered := crypto.CompressPubkey(pub)

	logger.New("database").Debug("IsOwner({type}, {id}, {owner})", logger.Attrs{
		"type": obj.Type,
		"id":   obj.ID,
		"hash": hash,

		"recovered": hex.EncodeToString(recovered),
		"owner":     hex.EncodeToString(obj.Owner[:]),
	})
	return bytes.Equal(recovered, obj.Owner[:])
}

// GetOwnerFromSignature returns 33-byte PublicKey from given signature.
func GetOwnerFromSignature(obj *Object, sig []byte) (PublicKey, error) {
	hash := GetObjectHash(obj.Type, obj.ID, obj.Data)
	pubkey, err := crypto.SigToPub(hash[:], sig)
	if err != nil {
		return PublicKey{}, err
	}
	pub := crypto.CompressPubkey(pubkey)
	var p PublicKey
	copy(p[:], pub[:])
	return p, nil
}

func GetObjectHash(typ, id string, data Payload) [32]byte {
	rawData, _ := json.MarshalToString(data)
	preimage := fmt.Sprintf("%s/%s/%s", typ, id, rawData)
	return sha3.Sum256([]byte(preimage))
}
