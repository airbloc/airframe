package database

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	testData1 = Payload{"foo": "bar"}
	testData2 = Payload{"foo": "baz"}
)

func getSignature(priv *ecdsa.PrivateKey, typ, id string, data Payload) []byte {
	// generate signature
	hash := GetObjectHash(typ, id, data)
	sig, _ := crypto.Sign(hash[:], priv)
	return sig
}

func TestInMemoryDatabase_Put(t *testing.T) {
	ctx := context.TODO()
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()

	// test creation
	sig := getSignature(priv, "testdata", "1", testData1)
	result, err := imdb.Put(ctx, "testdata", "1", testData1, sig)
	require.NoError(t, err)
	require.Equal(t, true, result.Created)

	// test update
	newSig := getSignature(priv, "testdata", "1", testData2)
	result, err = imdb.Put(ctx, "testdata", "1", testData2, newSig)
	require.NoError(t, err)
	require.Equal(t, false, result.Created)
}

func TestInMemoryDatabase_Get(t *testing.T) {
	ctx := context.TODO()
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()
	sig := getSignature(priv, "testdata", "1", testData1)

	_, err := imdb.Put(ctx, "testdata", "1", testData1, sig[:])
	require.NoError(t, err)

	obj, err := imdb.Get(ctx, "testdata", "1")
	require.NoError(t, err)
	require.Equal(t, testData1, obj.Data)
}

func TestInMemoryDatabase_Query(t *testing.T) {
	ctx := context.TODO()
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()

	sig := getSignature(priv, "testdata", "1", testData1)
	_, err := imdb.Put(ctx, "testdata", "1", testData1, sig)
	require.NoError(t, err)
	sig = getSignature(priv, "testdata", "2", testData2)
	_, err = imdb.Put(ctx, "testdata", "2", testData2, sig)
	require.NoError(t, err)

	// test equals
	q, err := QueryFromJson(`{"foo": "bar"}`)
	require.NoError(t, err)
	results, err := imdb.Query(ctx, "testdata", q, 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))

	// test contains
	q, err = QueryFromJson(`{"foo": {"contains": "b"}}`)
	require.NoError(t, err)
	results, err = imdb.Query(ctx, "testdata", q, 0, 0)
	require.NoError(t, err)
	require.Equal(t, 2, len(results))

	// test skip
	q, err = QueryFromJson(`{"foo": {"contains": "b"}}`)
	require.NoError(t, err)
	results, err = imdb.Query(ctx, "testdata", q, 1, 0)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))

	// test limit
	q, err = QueryFromJson(`{"foo": {"contains": "b"}}`)
	require.NoError(t, err)
	results, err = imdb.Query(ctx, "testdata", q, 0, 1)
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
}

func TestInMemoryDatabase_Exists(t *testing.T) {
	ctx := context.TODO()
	imdb, _ := NewInMemoryDatabase()
	priv, _ := crypto.GenerateKey()
	sig := getSignature(priv, "testdata", "1", testData1)

	_, err := imdb.Put(ctx, "testdata", "1", testData1, sig[:])
	require.NoError(t, err)

	exists, err := imdb.Exists(ctx, "testdata", "1")
	require.NoError(t, err)
	require.Equal(t, true, exists)
}
