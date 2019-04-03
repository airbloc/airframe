// package dynamodb implements DynamoDB backend interface.
package dynamodatabase

import (
	"context"
	"github.com/airbloc/airframe/database"
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/guregu/dynamo"
	"github.com/pkg/errors"
	"time"
)

var (
	reservedFields = []string{"ID", "Data", "Owner", "CreatedAt", "LastUpdatedAt"}

	dynamoOperators = map[database.OperatorType]string{
		database.OpEquals:             "$ = ?",
		database.OpGreaterThan:        "$ > ?",
		database.OpGreaterThanOrEqual: "$ >= ?",
		database.OpLessThan:           "$ < ?",
		database.OpLessThanOrEqual:    "$ <= ?",
		database.OpContains:           "contains($, ?)",
	}
)

type DynamoDatabase struct {
	svc *dynamo.DB

	tablePrefix string
}

func New(session awsclient.ConfigProvider) *DynamoDatabase {
	return &DynamoDatabase{
		svc: dynamo.New(session),

		tablePrefix: "airbloc_",
	}
}

func (db *DynamoDatabase) Get(ctx context.Context, typ, id string) (*database.Object, error) {
	table := db.svc.Table(db.tablePrefix + typ)

	items := make(map[string]*dynamodb.AttributeValue)
	if err := table.Get("ID", id).OneWithContext(ctx, &items); err != nil {
		if err == dynamo.ErrNotFound {
			return nil, database.ErrNotExists
		}
		return nil, errors.Wrap(err, "failed to get item from DynamoDB")
	}

	// trick: copy the data attrs into Data object, since the result is flattened
	items["Data"] = &dynamodb.AttributeValue{M: make(map[string]*dynamodb.AttributeValue)}
CopyData:
	for key, value := range items {
		for _, reserved := range reservedFields {
			if key == reserved {
				continue CopyData
			}
		}
		items["Data"].M[key] = value
	}

	// now we can unmarshal it xD
	obj := database.Object{}
	if err := dynamo.UnmarshalItem(items, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (db *DynamoDatabase) Exists(ctx context.Context, typ, id string) (bool, error) {
	table := db.svc.Table(db.tablePrefix + typ)
	count, err := table.Get("ID", id).Count()
	if err != nil {
		return false, errors.Wrap(err, "failed to get item from DynamoDB")
	}
	return count > 0, nil
}

func (db *DynamoDatabase) Query(ctx context.Context, typ string, query *database.Query, skip, limit int) ([]*database.Object, error) {
	table := db.svc.Table(db.tablePrefix + typ)

	// TODO: Use Query instead of Scan.
	q := table.Scan()
	for _, op := range query.Conditions {
		filter := dynamoOperators[op.Type]
		q.Filter(filter, op.Field, op.Operand)
	}
	q.Limit(int64(skip + limit))

	var items []map[string]*dynamodb.AttributeValue
	if err := q.AllWithContext(ctx, &items); err != nil {
		return nil, errors.Wrap(err, "failed to scan item from DynamoDB")
	}
	items = items[skip:]

	results := make([]*database.Object, len(items))
	for i := 0; i < len(items); i++ {
		item := items[i]

		// trick: copy the data attrs into Data object, since the result is flattened
		item["Data"] = &dynamodb.AttributeValue{M: make(map[string]*dynamodb.AttributeValue)}
	CopyData:
		for key, value := range item {
			for _, reserved := range reservedFields {
				if key == reserved {
					continue CopyData
				}
			}
			item["Data"].M[key] = value
		}

		// now we can unmarshal it xD
		obj := database.Object{}
		if err := dynamo.UnmarshalItem(item, &obj); err != nil {
			return nil, err
		}
		results[i] = &obj
	}
	return results, nil
}

func (db *DynamoDatabase) Put(ctx context.Context, typ, id string, data database.Payload, signature []byte) (*database.PutResult, error) {
	created := false
	obj, err := db.Get(ctx, typ, id)
	if err == database.ErrNotExists {
		// create new
		obj = &database.Object{
			ID:   id,
			Type: typ,

			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
		}
		if obj.Owner, err = database.GetOwnerFromSignature(obj, signature); err != nil {
			return nil, errors.Wrap(err, "invalid signature")
		}
		created = true

	} else if err == nil {
		// update object
		obj.Data = data
		if !database.IsOwner(obj, signature) {
			return nil, database.ErrNotAuthorized
		}
		obj.LastUpdatedAt = time.Now()

	} else {
		return nil, errors.Wrap(err, "error while checking existence")
	}

	item, err := dynamo.MarshalItem(obj)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal data")
	}

	// flatten the obj.Data
	for k, v := range item["Data"].M {
		item[k] = v
	}
	delete(item, "Data")

	table := db.svc.Table(db.tablePrefix + typ)
	if err := table.Put(item).RunWithContext(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to write to DynamoDB")
	}
	return &database.PutResult{
		FeeUsed: 0,
		Created: created,
	}, nil
}
