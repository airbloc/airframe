package database

import (
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

func getSignature(priv *ecdsa.PrivateKey, typ, id, data string) []byte {
	// generate signature
	hash := GetObjectHash(typ, id, data)
	sig, _ := crypto.Sign(hash[:], priv)
	return sig
}

func TestInMemoryDatabase_Put(t *testing.T) {
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()

	// test creation
	sig := getSignature(priv, "testdata", "1", "Hello World!")
	result, err := imdb.Put("testdata", "1", "Hello World!", sig)
	require.NoError(t, err)
	require.Equal(t, true, result.Created)

	// test update
	newSig := getSignature(priv, "testdata", "1", "Hello Eorld!")
	result, err = imdb.Put("testdata", "1", "Hello Eorld!", newSig)
	require.NoError(t, err)
	require.Equal(t, false, result.Created)
}

func TestInMemoryDatabase_Get(t *testing.T) {
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()
	sig := getSignature(priv, "testdata", "1", "Hello World!")

	_, err := imdb.Put("testdata", "1", "Hello World!", sig[:])
	require.NoError(t, err)

	obj, err := imdb.Get("/testdata/1")
	require.NoError(t, err)
	require.Equal(t, "Hello World!", obj.Data)
}

func TestInMemoryDatabase_Exists(t *testing.T) {
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()
	sig := getSignature(priv, "testdata", "1", "Hello World!")

	_, err := imdb.Put("testdata", "1", "Hello World!", sig[:])
	require.NoError(t, err)

	exists, err := imdb.Exists("testdata", "1")
	require.NoError(t, err)
	require.Equal(t, true, exists)
}
