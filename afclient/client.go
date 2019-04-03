// package afclient implements Go SDK of Airframe.
package afclient

import (
	"context"
	"crypto/ecdsa"
	"github.com/airbloc/airframe/auth"
	pb "github.com/airbloc/airframe/proto"
	"github.com/airbloc/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary

	// ErrNotExists is raised when request object is not found.
	ErrNotExists = errors.New("given object not exists.")

	// ErrNotAuthorized is raised when given signature mismatched with the object owner's one.
	ErrNotAuthorized = errors.New("you're not authorized to update the object.")
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
	api pb.APIClient
	key *ecdsa.PrivateKey

	log logger.Logger
}

func Dial(addr string, key *ecdsa.PrivateKey) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect gRPC server")
	}
	return &Client{
		key: key,
		api: pb.NewAPIClient(conn),
	}, nil
}

func (c *Client) Get(ctx context.Context, typ, id string) (*Object, error) {
	res, err := c.api.GetObject(ctx, &pb.GetRequest{
		Type: typ,
		Id:   id,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotExists
		}
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

func (c *Client) Query(ctx context.Context, query M, options ...QueryOption) ([]*Object, error) {
	opt := queryOptions{
		skip: 0,
		limit: 0,
	}
	for _, applyFunc := range options {
		applyFunc(&opt)
	}

	q, err := json.MarshalToString(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal query")
	}

	res, err := c.api.QueryObject(ctx, &pb.QueryRequest{
		Query: q,
		Skip:  uint64(opt.skip),
		Limit: uint64(opt.limit),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to call RPC")
	}

	results := res.GetResults()
	objects := make([]*Object, len(results))
	for i := 0; i < len(results); i++ {
		objects[i] = &Object{
			Owner:         common.HexToAddress(results[i].GetOwner()),
			CreatedAt:     time.Unix(0, int64(results[i].GetCreatedAt())),
			LastUpdatedAt: time.Unix(0, int64(results[i].GetLastUpdatedAt())),
		}
		if err := json.UnmarshalFromString(results[i].GetData(), &objects[i].Data); err != nil {
			return nil, errors.Wrap(err, "error on unmarshalling data")
		}
	}
	return objects, nil
}

func (c *Client) Put(ctx context.Context, typ, id string, data M) (*PutResult, error) {
	hash := auth.GetObjectHash(typ, id, data)

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

	res, err := c.api.PutObject(ctx, &pb.PutRequest{
		Type:      typ,
		Id:        id,
		Data:      marshalledData,
		Signature: sig,
	})
	if err != nil {
		if status.Code(err) == codes.Unauthenticated {
			return nil, ErrNotAuthorized
		}
		return nil, errors.Wrap(err, "failed to call RPC")
	}
	return &PutResult{
		FeeUsed: res.GetFeeUsed(),
		Created: res.GetCreated(),
	}, nil
}
