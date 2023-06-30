// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.20.0
// 	protoc        v3.12.4
// source: mongod/proto/mongo.proto

package proto

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type MongoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Db         string `protobuf:"bytes,1,opt,name=db,proto3" json:"db,omitempty"`
	Collection string `protobuf:"bytes,2,opt,name=collection,proto3" json:"collection,omitempty"`
	Query      []byte `protobuf:"bytes,3,opt,name=query,proto3" json:"query,omitempty"`
	Obj        []byte `protobuf:"bytes,4,opt,name=obj,proto3" json:"obj,omitempty"`
}

func (x *MongoRequest) Reset() {
	*x = MongoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mongod_proto_mongo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MongoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MongoRequest) ProtoMessage() {}

func (x *MongoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mongod_proto_mongo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MongoRequest.ProtoReflect.Descriptor instead.
func (*MongoRequest) Descriptor() ([]byte, []int) {
	return file_mongod_proto_mongo_proto_rawDescGZIP(), []int{0}
}

func (x *MongoRequest) GetDb() string {
	if x != nil {
		return x.Db
	}
	return ""
}

func (x *MongoRequest) GetCollection() string {
	if x != nil {
		return x.Collection
	}
	return ""
}

func (x *MongoRequest) GetQuery() []byte {
	if x != nil {
		return x.Query
	}
	return nil
}

func (x *MongoRequest) GetObj() []byte {
	if x != nil {
		return x.Obj
	}
	return nil
}

type MongoConfigRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Db         string   `protobuf:"bytes,1,opt,name=db,proto3" json:"db,omitempty"`
	Collection string   `protobuf:"bytes,2,opt,name=collection,proto3" json:"collection,omitempty"`
	Indexkeys  []string `protobuf:"bytes,3,rep,name=indexkeys,proto3" json:"indexkeys,omitempty"`
}

func (x *MongoConfigRequest) Reset() {
	*x = MongoConfigRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mongod_proto_mongo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MongoConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MongoConfigRequest) ProtoMessage() {}

func (x *MongoConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_mongod_proto_mongo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MongoConfigRequest.ProtoReflect.Descriptor instead.
func (*MongoConfigRequest) Descriptor() ([]byte, []int) {
	return file_mongod_proto_mongo_proto_rawDescGZIP(), []int{1}
}

func (x *MongoConfigRequest) GetDb() string {
	if x != nil {
		return x.Db
	}
	return ""
}

func (x *MongoConfigRequest) GetCollection() string {
	if x != nil {
		return x.Collection
	}
	return ""
}

func (x *MongoConfigRequest) GetIndexkeys() []string {
	if x != nil {
		return x.Indexkeys
	}
	return nil
}

type MongoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok   string   `protobuf:"bytes,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Objs [][]byte `protobuf:"bytes,2,rep,name=objs,proto3" json:"objs,omitempty"`
}

func (x *MongoResponse) Reset() {
	*x = MongoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_mongod_proto_mongo_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MongoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MongoResponse) ProtoMessage() {}

func (x *MongoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_mongod_proto_mongo_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MongoResponse.ProtoReflect.Descriptor instead.
func (*MongoResponse) Descriptor() ([]byte, []int) {
	return file_mongod_proto_mongo_proto_rawDescGZIP(), []int{2}
}

func (x *MongoResponse) GetOk() string {
	if x != nil {
		return x.Ok
	}
	return ""
}

func (x *MongoResponse) GetObjs() [][]byte {
	if x != nil {
		return x.Objs
	}
	return nil
}

var File_mongod_proto_mongo_proto protoreflect.FileDescriptor

var file_mongod_proto_mongo_proto_rawDesc = []byte{
	0x0a, 0x18, 0x6d, 0x6f, 0x6e, 0x67, 0x6f, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d,
	0x6f, 0x6e, 0x67, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x66, 0x0a, 0x0c, 0x4d, 0x6f,
	0x6e, 0x67, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x64, 0x62,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x64, 0x62, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a,
	0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x71, 0x75,
	0x65, 0x72, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x71, 0x75, 0x65, 0x72, 0x79,
	0x12, 0x10, 0x0a, 0x03, 0x6f, 0x62, 0x6a, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6f,
	0x62, 0x6a, 0x22, 0x62, 0x0a, 0x12, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x64, 0x62, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x64, 0x62, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x6f, 0x6c, 0x6c,
	0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x63, 0x6f,
	0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x69, 0x6e, 0x64,
	0x65, 0x78, 0x6b, 0x65, 0x79, 0x73, 0x22, 0x33, 0x0a, 0x0d, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x6f, 0x6b, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x62, 0x6a, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04, 0x6f, 0x62, 0x6a, 0x73, 0x32, 0x84, 0x02, 0x0a, 0x05,
	0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x12, 0x27, 0x0a, 0x06, 0x49, 0x6e, 0x73, 0x65, 0x72, 0x74, 0x12,
	0x0d, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e,
	0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27,
	0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x0d, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x06, 0x55, 0x70, 0x73, 0x65, 0x72,
	0x74, 0x12, 0x0d, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x0e, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x25, 0x0a, 0x04, 0x46, 0x69, 0x6e, 0x64, 0x12, 0x0d, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a, 0x04, 0x44, 0x72, 0x6f, 0x70, 0x12,
	0x13, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x05, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x13, 0x2e,
	0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x4d, 0x6f, 0x6e, 0x67, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x16, 0x5a, 0x14, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x6d, 0x6f,
	0x6e, 0x67, 0x6f, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_mongod_proto_mongo_proto_rawDescOnce sync.Once
	file_mongod_proto_mongo_proto_rawDescData = file_mongod_proto_mongo_proto_rawDesc
)

func file_mongod_proto_mongo_proto_rawDescGZIP() []byte {
	file_mongod_proto_mongo_proto_rawDescOnce.Do(func() {
		file_mongod_proto_mongo_proto_rawDescData = protoimpl.X.CompressGZIP(file_mongod_proto_mongo_proto_rawDescData)
	})
	return file_mongod_proto_mongo_proto_rawDescData
}

var file_mongod_proto_mongo_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_mongod_proto_mongo_proto_goTypes = []interface{}{
	(*MongoRequest)(nil),       // 0: MongoRequest
	(*MongoConfigRequest)(nil), // 1: MongoConfigRequest
	(*MongoResponse)(nil),      // 2: MongoResponse
}
var file_mongod_proto_mongo_proto_depIdxs = []int32{
	0, // 0: Mongo.Insert:input_type -> MongoRequest
	0, // 1: Mongo.Update:input_type -> MongoRequest
	0, // 2: Mongo.Upsert:input_type -> MongoRequest
	0, // 3: Mongo.Find:input_type -> MongoRequest
	1, // 4: Mongo.Drop:input_type -> MongoConfigRequest
	1, // 5: Mongo.Index:input_type -> MongoConfigRequest
	2, // 6: Mongo.Insert:output_type -> MongoResponse
	2, // 7: Mongo.Update:output_type -> MongoResponse
	2, // 8: Mongo.Upsert:output_type -> MongoResponse
	2, // 9: Mongo.Find:output_type -> MongoResponse
	2, // 10: Mongo.Drop:output_type -> MongoResponse
	2, // 11: Mongo.Index:output_type -> MongoResponse
	6, // [6:12] is the sub-list for method output_type
	0, // [0:6] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_mongod_proto_mongo_proto_init() }
func file_mongod_proto_mongo_proto_init() {
	if File_mongod_proto_mongo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_mongod_proto_mongo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MongoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mongod_proto_mongo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MongoConfigRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_mongod_proto_mongo_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MongoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_mongod_proto_mongo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_mongod_proto_mongo_proto_goTypes,
		DependencyIndexes: file_mongod_proto_mongo_proto_depIdxs,
		MessageInfos:      file_mongod_proto_mongo_proto_msgTypes,
	}.Build()
	File_mongod_proto_mongo_proto = out.File
	file_mongod_proto_mongo_proto_rawDesc = nil
	file_mongod_proto_mongo_proto_goTypes = nil
	file_mongod_proto_mongo_proto_depIdxs = nil
}