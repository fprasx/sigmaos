package mongod

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	sp "sigmaos/sigmap"
	dbg "sigmaos/debug"
	"sigmaos/fs"
	"sigmaos/protdevsrv"
	"sigmaos/mongod/proto"
	"time"
	"context"
	"fmt"
)

const (
	MONGO_NO = "No"
	MONGO_OK = "OK"
	DIAL_TIMEOUT_SEC = 1
	POOL_SIZE = 1000
	SOCKET_TIMEOUT_MIN = 5
	SYNC_TIMEOUT_SEC = 10
)

type Server struct {
	mclnt *mongo.Client
}

func makeServer(mongodUrl string) (*Server, error) {
	s := &Server{}
	uri := "mongodb://"+mongodUrl
	ctx, _ := context.WithTimeout(context.Background(), DIAL_TIMEOUT_SEC * time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetMaxPoolSize(POOL_SIZE))
	if err != nil {
		dbg.DFatalf("mongo dial err %v\n", err)
		return nil, err
	}
	s.mclnt = client
	if err = s.mclnt.Ping(context.TODO(), nil); err != nil {
		dbg.DFatalf("mongo ping err %v\n", err)
	}
	return s, nil
}

func RunMongod(mongodUrl string) error {
	dbg.DPrintf(dbg.MONGO, "Making mongo proxy server at %v", mongodUrl)
	s, err := makeServer(mongodUrl)
	if err != nil {
		return err
	}
	dbg.DPrintf(dbg.MONGO, "Starting mongo proxy server")
	pds, err := protdevsrv.MakeProtDevSrv(sp.MONGO, s)
	if err != nil {
		return err
	}
	return pds.RunServer()
}

func (s *Server) Insert(ctx fs.CtxI, req proto.MongoRequest, res *proto.MongoResponse) error {
	res.Ok = MONGO_NO
	var m bson.M
	if err := bson.Unmarshal(req.Obj, &m); err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot decode insert request: %v", err)
		return err
	}
	dbg.DPrintf(dbg.MONGO, "Received insert request: %v, %v, %v", req.Db, req.Collection, m)
	_, err := s.mclnt.Database(req.Db).Collection(req.Collection).InsertOne(context.TODO(), &m)
	if err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot insert: %v", err)
		return err
	}
	res.Ok = MONGO_OK
	return nil
}

func (s *Server) Update(ctx fs.CtxI, req proto.MongoRequest, res *proto.MongoResponse) error {
	return s.update(ctx, req, res, false)
}

func (s *Server) Upsert(ctx fs.CtxI, req proto.MongoRequest, res *proto.MongoResponse) error {
	return s.update(ctx, req, res, true)
}

func (s *Server) update(
		ctx fs.CtxI, req proto.MongoRequest, res *proto.MongoResponse, upsert bool) error {
	res.Ok = MONGO_NO
	rpcName := "update"
	if upsert {
		rpcName = "upsert"
	}
	var q, u bson.M
	if err := bson.Unmarshal(req.Query, &q); err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot decode query in %v request: %v", rpcName, err)
		return err
	}
	if err := bson.Unmarshal(req.Obj, &u); err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot decode object in %v request: %v", rpcName, err)
		return err
	}
	dbg.DPrintf(
		dbg.MONGO, "Received %v request: %v, %v, %v, %v", rpcName, req.Db, req.Collection, q, u)
	_, err := s.mclnt.Database(req.Db).Collection(req.Collection).UpdateOne(
		context.TODO(), &q, &u, options.Update().SetUpsert(upsert))
	if err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot %v: %v", rpcName, err)
		return err
	}
	res.Ok = MONGO_OK
	return nil
}

func (s *Server) Find(ctx fs.CtxI, req proto.MongoRequest, res *proto.MongoResponse) error {
	res.Ok = MONGO_NO
	var m bson.M
	if err := bson.Unmarshal(req.Query, &m); err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot decode find query request: %v", err)
		return err
	}
	dbg.DPrintf(dbg.MONGO, "Received Find request. %v, %v, %v", req.Db, req.Collection, m)
	var objs []bson.M
	mres, err1 := s.mclnt.Database(req.Db).Collection(req.Collection).Find(context.TODO(), &m)
	err2 := mres.All(context.TODO(), &objs)
	if err1 != nil || err2 != nil {
		err := fmt.Errorf("%w; %w", err1, err2)
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot find objects: %v. Err: %v", m, err)
		return err
	}
	res.Objs = make([][]byte, len(objs))
	for i, obj := range objs {
		res.Objs[i], _ = bson.Marshal(obj)
	}
	res.Ok = MONGO_OK
	return nil
}

func (s *Server) Drop(ctx fs.CtxI, req proto.MongoConfigRequest, res *proto.MongoResponse) error {
	dbg.DPrintf(dbg.MONGO, "Received drop request: %v", req)
	res.Ok = MONGO_NO
	err := s.mclnt.Database(req.Db).Collection(req.Collection).Drop(context.TODO())
	if err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot Drop collection  %v. Err: %v", req, err)
		return err
	}
	res.Ok = MONGO_OK
	return nil
}

func (s *Server) Remove(ctx fs.CtxI, req proto.MongoConfigRequest, res *proto.MongoResponse) error {
	dbg.DPrintf(dbg.MONGO, "Received remove request: %v", req)
	res.Ok = MONGO_NO
	_, err := s.mclnt.Database(req.Db).Collection(req.Collection).DeleteMany(
		context.TODO(), &bson.M{})
	if err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot delete %v. Err: %v", req, err)
		return err
	}
	res.Ok = MONGO_OK
	return nil
}

func (s *Server) Index(ctx fs.CtxI, req proto.MongoConfigRequest, res *proto.MongoResponse) error {
	dbg.DPrintf(dbg.MONGO, "Received index request: %v", req)
	res.Ok = MONGO_NO
	indexKeys := bson.D{}
	for _, key := range req.Indexkeys {
		indexKeys = append(indexKeys, bson.E{key, 1})
	}
	name, err := s.mclnt.Database(req.Db).Collection(req.Collection).Indexes().CreateOne(
		context.TODO(), mongo.IndexModel{Keys: indexKeys})
	dbg.DPrintf(dbg.MONGO, "Name of index created: %v", name)
	if err != nil {
		dbg.DPrintf(dbg.MONGO_ERR, "Cannot index %v. Err: %v", req, err)
		return err
	}
	res.Ok = MONGO_OK
	return nil
}
