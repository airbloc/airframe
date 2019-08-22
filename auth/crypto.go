package auth

import (
	"fmt"
	"github.com/json-iterator/go"
	"github.com/klaytn/klaytn/crypto"
	"golang.org/x/crypto/sha3"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

// PublicKey is 33-byte compressed ECDSA public key.
type PublicKey [33]byte

// GetOwnerFromSignature returns 33-byte PublicKey from given signature.
func GetSigner(typ, id string, data interface{}, sig []byte) (PublicKey, error) {
	hash := GetObjectHash(typ, id, data)
	pubkey, err := crypto.SigToPub(hash[:], sig)
	if err != nil {
		return PublicKey{}, err
	}
	signerPub := crypto.CompressPubkey(pubkey)
	var p PublicKey
	copy(p[:], signerPub[:])
	return p, nil
}

func GetObjectHash(typ, id string, data interface{}) [32]byte {
	rawData, _ := json.MarshalToString(data)
	preimage := fmt.Sprintf("%s/%s/%s", typ, id, rawData)
	return sha3.Sum256([]byte(preimage))
}
