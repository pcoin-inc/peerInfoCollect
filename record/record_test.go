package record

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
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

func TestFindInfoWithNumber(t *testing.T) {
	client, err := ethclient.DialContext(context.TODO(), "http://65.108.120.57:30303")
	if err != nil {
		println("new client with peer err", err.Error())
	}

	id,err := client.ChainID(context.TODO())
	if err != nil {
		println("get chain id",err.Error())
	}

	if id != nil{
		println("id",id.Uint64())
	}

	//hash := common.HexToHash("0x52bbecea52b15dc492147db62caa5b02e4d91c43949f7f8b34aa25f93588d894")
	//
	//block, err := client.BlockByHash(context.TODO(),hash)
	//if err != nil {
	//	println("get block by hash err", err.Error())
	//}
	//if block != nil{
	//	println("block num",block.NumberU64())
	//}

}
