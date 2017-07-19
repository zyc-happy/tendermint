package mongodb

import (
	"fmt"

	"github.com/tendermint/tendermint/state/txindex"
	"github.com/tendermint/tendermint/types"
	mgo "gopkg.in/mgo.v2"
)

type MongoDB struct {
	session *mgo.Session
}

type doc struct {
	types.TxResult
	hash []byte `json:"hash"`
}

func NewMongoDB(url string) *MongoDB {
	session, err := mgo.Dial(url)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
	}
	session.SetMode(mgo.Monotonic, true)
	return &MongoDB{session}
}

func (db *MongoDB) Get(hash []byte) (*types.TxResult, error) {
	if len(hash) == 0 {
		return nil, txindex.ErrorEmptyHash
	}

	session := db.session.Copy()
	defer session.Close()

	c := session.DB("store").C("transactions")

	var d doc
	err := c.Find(bson.M{"hash": hash}).One(&d)
	if err != nil {
		return nil, fmt.Errorf("Database error: %v", err)
	}

	return d.TxResult, nil
}

func (db *MongoDB) AddBatch(b *txindex.Batch) error {
	session := db.session.Copy()
	defer session.Close()

	c := session.DB("store").C("transactions")

	docs := make([]doc, len(b.Ops))
	for _, result := range b.Ops {
		docs[i] = doc{hash: result.Tx.Hash(), TxResult: result}
	}

	bulk := c.Bulk()
	bulk.Unordered().Insert(docs)
	_, err := bulk.Run()
	if err != nil {
		return fmt.Errorf("Database error: %v", err)
	}

	return nil
}
