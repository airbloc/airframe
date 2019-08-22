package rpcserver

import (
	"context"
	"github.com/airbloc/airframe/database"
	pb "github.com/airbloc/airframe/proto"
	"github.com/json-iterator/go"
	"github.com/klaytn/klaytn/crypto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	json = jsoniter.ConfigCompatibleWithStandardLibrary
)

type API struct {
	db database.Database
}

func RegisterV1API(srv *grpc.Server, db database.Database) {
	api := API{db: db}
	pb.RegisterAPIServer(srv, &api)
}

func (api *API) GetObject(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	obj, err := api.db.Get(ctx, req.GetType(), req.GetId())
	if err != nil {
		if err == database.ErrNotExists {
			return nil, status.Error(codes.NotFound, "resource not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return objToGetResponse(obj), nil
}

func (api *API) QueryObject(ctx context.Context, req *pb.QueryRequest) (*pb.QueryResponse, error) {
	q := req.GetQuery()
	if q == "" {
		q = "{}"
	}
	query, err := database.QueryFromJson(q)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query")
	}
	objects, err := api.db.Query(ctx, req.GetType(), query, int(req.GetSkip()), int(req.GetLimit()))
	results := make([]*pb.GetResponse, len(objects))
	for i := 0; i < len(objects); i++ {
		results[i] = objToGetResponse(objects[i])
	}
	return &pb.QueryResponse{Results: results}, nil
}

func (api *API) PutObject(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	if len(req.Signature) != 65 {
		return nil, status.Errorf(codes.InvalidArgument, "invalid signature length: %d", len(req.Signature))
	}

	var data database.Payload
	if err := json.UnmarshalFromString(req.GetData(), &data); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid data: '%s'", req.GetData())
	}

	result, err := api.db.Put(ctx, req.GetType(), req.GetId(), data, req.Signature)
	if err != nil {
		if err == database.ErrNotAuthorized {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.PutResponse{
		Created: result.Created,
		FeeUsed: result.FeeUsed,
	}, nil
}

func objToGetResponse(obj *database.Object) *pb.GetResponse {
	pub, _ := crypto.DecompressPubkey(obj.Owner[:])
	ownerAddr := crypto.PubkeyToAddress(*pub)

	data, _ := json.MarshalToString(obj.Data)
	return &pb.GetResponse{
		Data:  data,
		Owner: ownerAddr.Hex(),

		CreatedAt:     uint64(obj.CreatedAt.UnixNano()),
		LastUpdatedAt: uint64(obj.LastUpdatedAt.UnixNano()),
	}
}
