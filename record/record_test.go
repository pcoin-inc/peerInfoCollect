package record

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
)

func TestCreateTable(t *testing.T) {
	var (
		client =   GetMgoCli()
		db         *mongo.Database
		collection *mongo.Collection
	)
	//2.选择数据库 my_db
	db = client.Database("my_db")

	//选择表 my_collection
	collection = db.Collection("my_collection")
	collection = collection

	defer client.Disconnect(context.TODO())

	r := &RecordInfo{
		BlockNum: 1,
		BlockHash: "22333",
		PeerId: "22",
		PeerAddress: "192.186.0.1",
	}

	err := InsertInfo(collection,r)
	if err != nil {
		println("insert err",err.Error())
	}


	s,err := FindInfoWithNumber(collection,1)
	if err != nil {
		println("find err",err.Error())
	}

	println("peer info",s.PeerAddress)
}
