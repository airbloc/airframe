package database

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"golang.org/x/crypto/sha3"
)

func IsOwner(obj *Object, sig []byte) bool {
	hash := GetObjectHash(obj.Type, obj.ID, obj.Data)

	pub, err := crypto.SigToPub(hash[:], sig)
	if err != nil {
		return false
	}
	recovered := crypto.CompressPubkey(pub)
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
	rawData := new(bytes.Buffer)
	gob.NewEncoder(rawData).Encode(data)
	typAndId := fmt.Sprintf("%s/%s", typ, id)
	return sha3.Sum256(append([]byte(typAndId), rawData.Bytes()...))
}
