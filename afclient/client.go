// package afclient implements Go SDK of Airframe.
package afclient

import (
	"context"
	"crypto/ecdsa"
	"github.com/airbloc/airframe/database"
	"github.com/airbloc/airframe/rpcserver"
	"github.com/airbloc/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"time"
)

var (
	DefaultTimeout = 10 * time.Second
	json           = jsoniter.ConfigCompatibleWithStandardLibrary
)

type M map[string]interface{}

type Object struct {
	Data  M
	Owner common.Address

	// timestamps
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

type PutResult struct {
	FeeUsed uint64
	Created bool
}

type Client struct {
	api rpcserver.APIClient
	key *ecdsa.PrivateKey

	Timeout time.Duration

	log logger.Logger
}

func Dial(addr string, key *ecdsa.PrivateKey) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect gRPC server")
	}
	return &Client{
		key:     key,
		api:     rpcserver.NewAPIClient(conn),
		Timeout: DefaultTimeout,
	}, nil
}

func (c *Client) Get(uri string) (*Object, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	res, err := c.api.GetObject(ctx, &rpcserver.GetRequest{Uri: uri})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call RPC")
	}

	obj := &Object{
		Owner:         common.HexToAddress(res.GetOwner()),
		CreatedAt:     time.Unix(0, int64(res.GetCreatedAt())),
		LastUpdatedAt: time.Unix(0, int64(res.GetLastUpdatedAt())),
	}
	if err := json.UnmarshalFromString(res.GetData(), &obj.Data); err != nil {
		return nil, errors.Wrap(err, "error on unmarshalling data")
	}
	return obj, nil
}

func (c *Client) Put(typ, id string, data M) (*PutResult, error) {
	hash := database.GetObjectHash(typ, id, database.Payload(data))

	c.log.Debug("Put({type}, {id}) by {owner}", logger.Attrs{
		"type":  typ,
		"id":    id,
		"hash":  hash,
		"owner": crypto.PubkeyToAddress(c.key.PublicKey).Hex(),
	})

	sig, err := crypto.Sign(hash[:], c.key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign data")
	}

	marshalledData, err := json.MarshalToString(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal data into JSON")
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	res, err := c.api.PutObject(ctx, &rpcserver.PutRequest{
		Type:      typ,
		Id:        id,
		Data:      marshalledData,
		Signature: sig,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call RPC")
	}
	return &PutResult{
		FeeUsed: res.GetFeeUsed(),
		Created: res.GetCreated(),
	}, nil
}
