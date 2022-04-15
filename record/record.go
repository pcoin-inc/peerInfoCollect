package record

import (
	"context"
	"peerInfoCollect/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RecordInfo struct {
	BlockNum   uint64  `bson:"blocknum"`//区块高度
	BlockHash  string  `bson:"blockhash"`//区块hash
	PeerId     string  `bson:"peerid"`//节点id
	PeerAddress string  `bson:"peeraddr"`//节点ip地址
}

var MgoCli *mongo.Client
var MgoCnn *mongo.Collection

func initEngine() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// 连接到MongoDB
	MgoCli, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		println("connect err",err.Error())
	}
	// 检查连接
	err = MgoCli.Ping(context.TODO(), nil)
	if err != nil {
		println("ping err",err.Error())
	}
}

func GetMgoCli() *mongo.Client {
	if MgoCli == nil {
		initEngine()
	}
	return MgoCli
}

func NewConnectionWithDBName(dbname,tab string)  {
	db := MgoCli.Database(dbname)
	//选择表 my_collection
	collection := db.Collection(tab)
	MgoCnn = collection
}

func InsertInfo(c *mongo.Collection,r *RecordInfo) error {
	insertRes,err := c.InsertOne(context.TODO(),&r)
	if err != nil {
		log.Error("insert info err","err info",err.Error())
		return err
	}
	log.Info("insert single document","id",insertRes.InsertedID)
	return nil
}

func FindInfoWithNumber(c *mongo.Collection,num uint64) (*RecordInfo,error) {
	filter := bson.D{{"blocknum",num}}
	var r RecordInfo
	err := c.FindOne(context.TODO(),filter).Decode(&r)
	if err != nil {
		log.Error("find info err","num",num)
		return nil, err
	}
	return &r,nil
}
