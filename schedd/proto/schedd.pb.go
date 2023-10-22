// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.24.3
// source: schedd/proto/schedd.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	proc "sigmaos/proc"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ForceRunRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ProcProto       *proc.ProcProto `protobuf:"bytes,1,opt,name=procProto,proto3" json:"procProto,omitempty"`
	MemAccountedFor bool            `protobuf:"varint,2,opt,name=memAccountedFor,proto3" json:"memAccountedFor,omitempty"`
}

func (x *ForceRunRequest) Reset() {
	*x = ForceRunRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForceRunRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForceRunRequest) ProtoMessage() {}

func (x *ForceRunRequest) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForceRunRequest.ProtoReflect.Descriptor instead.
func (*ForceRunRequest) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{0}
}

func (x *ForceRunRequest) GetProcProto() *proc.ProcProto {
	if x != nil {
		return x.ProcProto
	}
	return nil
}

func (x *ForceRunRequest) GetMemAccountedFor() bool {
	if x != nil {
		return x.MemAccountedFor
	}
	return false
}

type ForceRunResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ForceRunResponse) Reset() {
	*x = ForceRunResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ForceRunResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ForceRunResponse) ProtoMessage() {}

func (x *ForceRunResponse) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ForceRunResponse.ProtoReflect.Descriptor instead.
func (*ForceRunResponse) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{1}
}

type WaitRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PidStr string `protobuf:"bytes,1,opt,name=pidStr,proto3" json:"pidStr,omitempty"`
}

func (x *WaitRequest) Reset() {
	*x = WaitRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WaitRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WaitRequest) ProtoMessage() {}

func (x *WaitRequest) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WaitRequest.ProtoReflect.Descriptor instead.
func (*WaitRequest) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{2}
}

func (x *WaitRequest) GetPidStr() string {
	if x != nil {
		return x.PidStr
	}
	return ""
}

type WaitResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status []byte `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *WaitResponse) Reset() {
	*x = WaitResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WaitResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WaitResponse) ProtoMessage() {}

func (x *WaitResponse) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WaitResponse.ProtoReflect.Descriptor instead.
func (*WaitResponse) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{3}
}

func (x *WaitResponse) GetStatus() []byte {
	if x != nil {
		return x.Status
	}
	return nil
}

type NotifyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PidStr string `protobuf:"bytes,1,opt,name=pidStr,proto3" json:"pidStr,omitempty"`
	Status []byte `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *NotifyRequest) Reset() {
	*x = NotifyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyRequest) ProtoMessage() {}

func (x *NotifyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyRequest.ProtoReflect.Descriptor instead.
func (*NotifyRequest) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{4}
}

func (x *NotifyRequest) GetPidStr() string {
	if x != nil {
		return x.PidStr
	}
	return ""
}

func (x *NotifyRequest) GetStatus() []byte {
	if x != nil {
		return x.Status
	}
	return nil
}

type NotifyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *NotifyResponse) Reset() {
	*x = NotifyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyResponse) ProtoMessage() {}

func (x *NotifyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyResponse.ProtoReflect.Descriptor instead.
func (*NotifyResponse) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{5}
}

type GetCPUSharesRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetCPUSharesRequest) Reset() {
	*x = GetCPUSharesRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUSharesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUSharesRequest) ProtoMessage() {}

func (x *GetCPUSharesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUSharesRequest.ProtoReflect.Descriptor instead.
func (*GetCPUSharesRequest) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{6}
}

type GetCPUSharesResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Shares map[string]int64 `protobuf:"bytes,1,rep,name=shares,proto3" json:"shares,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"varint,2,opt,name=value,proto3"`
}

func (x *GetCPUSharesResponse) Reset() {
	*x = GetCPUSharesResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUSharesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUSharesResponse) ProtoMessage() {}

func (x *GetCPUSharesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUSharesResponse.ProtoReflect.Descriptor instead.
func (*GetCPUSharesResponse) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{7}
}

func (x *GetCPUSharesResponse) GetShares() map[string]int64 {
	if x != nil {
		return x.Shares
	}
	return nil
}

type GetCPUUtilRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RealmStr string `protobuf:"bytes,1,opt,name=realmStr,proto3" json:"realmStr,omitempty"`
}

func (x *GetCPUUtilRequest) Reset() {
	*x = GetCPUUtilRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUUtilRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUUtilRequest) ProtoMessage() {}

func (x *GetCPUUtilRequest) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUUtilRequest.ProtoReflect.Descriptor instead.
func (*GetCPUUtilRequest) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{8}
}

func (x *GetCPUUtilRequest) GetRealmStr() string {
	if x != nil {
		return x.RealmStr
	}
	return ""
}

type GetCPUUtilResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Util float64 `protobuf:"fixed64,1,opt,name=util,proto3" json:"util,omitempty"`
}

func (x *GetCPUUtilResponse) Reset() {
	*x = GetCPUUtilResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_schedd_proto_schedd_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCPUUtilResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCPUUtilResponse) ProtoMessage() {}

func (x *GetCPUUtilResponse) ProtoReflect() protoreflect.Message {
	mi := &file_schedd_proto_schedd_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCPUUtilResponse.ProtoReflect.Descriptor instead.
func (*GetCPUUtilResponse) Descriptor() ([]byte, []int) {
	return file_schedd_proto_schedd_proto_rawDescGZIP(), []int{9}
}

func (x *GetCPUUtilResponse) GetUtil() float64 {
	if x != nil {
		return x.Util
	}
	return 0
}

var File_schedd_proto_schedd_proto protoreflect.FileDescriptor

var file_schedd_proto_schedd_proto_rawDesc = []byte{
	0x0a, 0x19, 0x73, 0x63, 0x68, 0x65, 0x64, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73,
	0x63, 0x68, 0x65, 0x64, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x0f, 0x70, 0x72, 0x6f,
	0x63, 0x2f, 0x70, 0x72, 0x6f, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x65, 0x0a, 0x0f,
	0x46, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x28, 0x0a, 0x09, 0x70, 0x72, 0x6f, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x50, 0x72, 0x6f, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x09,
	0x70, 0x72, 0x6f, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x28, 0x0a, 0x0f, 0x6d, 0x65, 0x6d,
	0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x64, 0x46, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x0f, 0x6d, 0x65, 0x6d, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x65, 0x64,
	0x46, 0x6f, 0x72, 0x22, 0x12, 0x0a, 0x10, 0x46, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x75, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x25, 0x0a, 0x0b, 0x57, 0x61, 0x69, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74, 0x72, 0x22, 0x26,
	0x0a, 0x0c, 0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x3f, 0x0a, 0x0d, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x70, 0x69, 0x64, 0x53, 0x74, 0x72, 0x12,
	0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x10, 0x0a, 0x0e, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x15, 0x0a, 0x13, 0x47, 0x65, 0x74,
	0x43, 0x50, 0x55, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x8c, 0x01, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x53, 0x68, 0x61, 0x72, 0x65,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x39, 0x0a, 0x06, 0x73, 0x68, 0x61,
	0x72, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x47, 0x65, 0x74, 0x43,
	0x50, 0x55, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x2e, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x73, 0x68,
	0x61, 0x72, 0x65, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x22,
	0x2f, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x55, 0x74, 0x69, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x53, 0x74, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x61, 0x6c, 0x6d, 0x53, 0x74, 0x72,
	0x22, 0x28, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x55, 0x74, 0x69, 0x6c, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x74, 0x69, 0x6c, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x75, 0x74, 0x69, 0x6c, 0x32, 0xf4, 0x02, 0x0a, 0x06, 0x53,
	0x63, 0x68, 0x65, 0x64, 0x64, 0x12, 0x2f, 0x0a, 0x08, 0x46, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x75,
	0x6e, 0x12, 0x10, 0x2e, 0x46, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x75, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x46, 0x6f, 0x72, 0x63, 0x65, 0x52, 0x75, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x09, 0x57, 0x61, 0x69, 0x74, 0x53, 0x74,
	0x61, 0x72, 0x74, 0x12, 0x0c, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x0d, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x2a, 0x0a, 0x07, 0x53, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64, 0x12, 0x0e, 0x2e, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x27, 0x0a, 0x08,
	0x57, 0x61, 0x69, 0x74, 0x45, 0x78, 0x69, 0x74, 0x12, 0x0c, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x29, 0x0a, 0x06, 0x45, 0x78, 0x69, 0x74, 0x65, 0x64, 0x12,
	0x0e, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x0f, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x12, 0x28, 0x0a, 0x09, 0x57, 0x61, 0x69, 0x74, 0x45, 0x76, 0x69, 0x63, 0x74, 0x12, 0x0c, 0x2e,
	0x57, 0x61, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x57, 0x61,
	0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x45, 0x76,
	0x69, 0x63, 0x74, 0x12, 0x0e, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x0f, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3b, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x53, 0x68,
	0x61, 0x72, 0x65, 0x73, 0x12, 0x14, 0x2e, 0x47, 0x65, 0x74, 0x43, 0x50, 0x55, 0x53, 0x68, 0x61,
	0x72, 0x65, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x50, 0x55, 0x53, 0x68, 0x61, 0x72, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x16, 0x5a, 0x14, 0x73, 0x69, 0x67, 0x6d, 0x61, 0x6f, 0x73, 0x2f, 0x73, 0x63, 0x68,
	0x65, 0x64, 0x64, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_schedd_proto_schedd_proto_rawDescOnce sync.Once
	file_schedd_proto_schedd_proto_rawDescData = file_schedd_proto_schedd_proto_rawDesc
)

func file_schedd_proto_schedd_proto_rawDescGZIP() []byte {
	file_schedd_proto_schedd_proto_rawDescOnce.Do(func() {
		file_schedd_proto_schedd_proto_rawDescData = protoimpl.X.CompressGZIP(file_schedd_proto_schedd_proto_rawDescData)
	})
	return file_schedd_proto_schedd_proto_rawDescData
}

var file_schedd_proto_schedd_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_schedd_proto_schedd_proto_goTypes = []interface{}{
	(*ForceRunRequest)(nil),      // 0: ForceRunRequest
	(*ForceRunResponse)(nil),     // 1: ForceRunResponse
	(*WaitRequest)(nil),          // 2: WaitRequest
	(*WaitResponse)(nil),         // 3: WaitResponse
	(*NotifyRequest)(nil),        // 4: NotifyRequest
	(*NotifyResponse)(nil),       // 5: NotifyResponse
	(*GetCPUSharesRequest)(nil),  // 6: GetCPUSharesRequest
	(*GetCPUSharesResponse)(nil), // 7: GetCPUSharesResponse
	(*GetCPUUtilRequest)(nil),    // 8: GetCPUUtilRequest
	(*GetCPUUtilResponse)(nil),   // 9: GetCPUUtilResponse
	nil,                          // 10: GetCPUSharesResponse.SharesEntry
	(*proc.ProcProto)(nil),       // 11: ProcProto
}
var file_schedd_proto_schedd_proto_depIdxs = []int32{
	11, // 0: ForceRunRequest.procProto:type_name -> ProcProto
	10, // 1: GetCPUSharesResponse.shares:type_name -> GetCPUSharesResponse.SharesEntry
	0,  // 2: Schedd.ForceRun:input_type -> ForceRunRequest
	2,  // 3: Schedd.WaitStart:input_type -> WaitRequest
	4,  // 4: Schedd.Started:input_type -> NotifyRequest
	2,  // 5: Schedd.WaitExit:input_type -> WaitRequest
	4,  // 6: Schedd.Exited:input_type -> NotifyRequest
	2,  // 7: Schedd.WaitEvict:input_type -> WaitRequest
	4,  // 8: Schedd.Evict:input_type -> NotifyRequest
	6,  // 9: Schedd.GetCPUShares:input_type -> GetCPUSharesRequest
	1,  // 10: Schedd.ForceRun:output_type -> ForceRunResponse
	3,  // 11: Schedd.WaitStart:output_type -> WaitResponse
	5,  // 12: Schedd.Started:output_type -> NotifyResponse
	3,  // 13: Schedd.WaitExit:output_type -> WaitResponse
	5,  // 14: Schedd.Exited:output_type -> NotifyResponse
	3,  // 15: Schedd.WaitEvict:output_type -> WaitResponse
	5,  // 16: Schedd.Evict:output_type -> NotifyResponse
	7,  // 17: Schedd.GetCPUShares:output_type -> GetCPUSharesResponse
	10, // [10:18] is the sub-list for method output_type
	2,  // [2:10] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}

func init() { file_schedd_proto_schedd_proto_init() }
func file_schedd_proto_schedd_proto_init() {
	if File_schedd_proto_schedd_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_schedd_proto_schedd_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForceRunRequest); i {
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
		file_schedd_proto_schedd_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ForceRunResponse); i {
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
		file_schedd_proto_schedd_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WaitRequest); i {
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
		file_schedd_proto_schedd_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WaitResponse); i {
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
		file_schedd_proto_schedd_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyRequest); i {
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
		file_schedd_proto_schedd_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyResponse); i {
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
		file_schedd_proto_schedd_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUSharesRequest); i {
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
		file_schedd_proto_schedd_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUSharesResponse); i {
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
		file_schedd_proto_schedd_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUUtilRequest); i {
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
		file_schedd_proto_schedd_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCPUUtilResponse); i {
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
			RawDescriptor: file_schedd_proto_schedd_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_schedd_proto_schedd_proto_goTypes,
		DependencyIndexes: file_schedd_proto_schedd_proto_depIdxs,
		MessageInfos:      file_schedd_proto_schedd_proto_msgTypes,
	}.Build()
	File_schedd_proto_schedd_proto = out.File
	file_schedd_proto_schedd_proto_rawDesc = nil
	file_schedd_proto_schedd_proto_goTypes = nil
	file_schedd_proto_schedd_proto_depIdxs = nil
}
