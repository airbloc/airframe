package apiserver

import (
	"github.com/airbloc/airframe/database"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

var (
	msgInvalidSigLength = "should be 65-byte ECDSA signature with [R || S || V] format"
)

type PutRequest struct {
	Data      database.Payload `json:"data" binding:"required"`
	Signature string           `json:"signature" binding:"required"`
}

func RegisterV1API(r *gin.Engine, db database.Database) {
	route := r.Group("/v1")
	route.GET("/object/:type/:id", handleGetObject(db))
	route.GET("/object/:type", handleQuery(db))
	route.POST("/object/:type/:id", handlePutObject(db))

	// health check
	route.GET("/", func(c *gin.Context) {
		priv, _ := crypto.GenerateKey()
		hash := database.GetObjectHash("testdata", "deadbeef", database.Payload{"foo": "bar"})
		sig, _ := crypto.Sign(hash[:], priv)

		if _, err := db.Put(c, "testdata", "deadbeef", database.Payload{"foo": "bar"}, sig); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "addr": crypto.PubkeyToAddress(priv.PublicKey).Hex()})
	})
}

func handleGetObject(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		obj, err := db.Get(c, c.Param("type"), c.Param("id"))
		if err != nil {
			if err == database.ErrNotExists {
				c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objectToJson(obj))
	}
}

func handleQuery(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		q := c.Query("query")
		if q == "" {
			q = "{}"
		}
		limit, err := strconv.Atoi(c.Query("limit"))
		if err != nil {
			limit = 0
		}
		skip, err := strconv.Atoi(c.Query("skip"))
		if err != nil {
			skip = 0
		}
		query, err := database.QueryFromJson(q)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
			return
		}

		objects, err := db.Query(c, c.Param("type"), query, limit, skip)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]gin.H, len(objects))
		for i := 0; i < len(objects); i++ {
			results[i] = objectToJson(objects[i])
		}
		c.JSON(http.StatusOK, gin.H{"results": results})
	}
}

func handlePutObject(db database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PutRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		typ, id := c.Param("type"), c.Param("id")

		sig, err := hexutil.Decode(req.Signature)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature: " + err.Error()})
			return
		}
		if len(sig) != 65 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature: " + msgInvalidSigLength})
			return
		}

		result, err := db.Put(c, typ, id, req.Data, sig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"created": result.Created,
			"feeUsed": result.FeeUsed,
		})
	}
}

func objectToJson(obj *database.Object) gin.H {
	pub, _ := crypto.DecompressPubkey(obj.Owner[:])
	ownerAddr := crypto.PubkeyToAddress(*pub)

	return gin.H{
		"data":          obj.Data,
		"owner":         ownerAddr.Hex(),
		"createdAt":     obj.CreatedAt,
		"lastUpdatedAt": obj.LastUpdatedAt,
	}
}
