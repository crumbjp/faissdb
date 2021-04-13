// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: rpcreplica/rpcreplica.proto

package rpcreplica

import (
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

type GetLastKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetLastKeyRequest) Reset() {
	*x = GetLastKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetLastKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLastKeyRequest) ProtoMessage() {}

func (x *GetLastKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLastKeyRequest.ProtoReflect.Descriptor instead.
func (*GetLastKeyRequest) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{0}
}

type GetLastKeyReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Lastkey string `protobuf:"bytes,1,opt,name=lastkey,proto3" json:"lastkey,omitempty"`
}

func (x *GetLastKeyReply) Reset() {
	*x = GetLastKeyReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetLastKeyReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetLastKeyReply) ProtoMessage() {}

func (x *GetLastKeyReply) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetLastKeyReply.ProtoReflect.Descriptor instead.
func (*GetLastKeyReply) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{1}
}

func (x *GetLastKeyReply) GetLastkey() string {
	if x != nil {
		return x.Lastkey
	}
	return ""
}

type GetTrainedRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetTrainedRequest) Reset() {
	*x = GetTrainedRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTrainedRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTrainedRequest) ProtoMessage() {}

func (x *GetTrainedRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTrainedRequest.ProtoReflect.Descriptor instead.
func (*GetTrainedRequest) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{2}
}

type GetTrainedReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *GetTrainedReply) Reset() {
	*x = GetTrainedReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTrainedReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTrainedReply) ProtoMessage() {}

func (x *GetTrainedReply) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTrainedReply.ProtoReflect.Descriptor instead.
func (*GetTrainedReply) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{3}
}

func (x *GetTrainedReply) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type GetDataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Startkey string `protobuf:"bytes,1,opt,name=startkey,proto3" json:"startkey,omitempty"`
	Length   int32  `protobuf:"varint,2,opt,name=length,proto3" json:"length,omitempty"`
}

func (x *GetDataRequest) Reset() {
	*x = GetDataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataRequest) ProtoMessage() {}

func (x *GetDataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataRequest.ProtoReflect.Descriptor instead.
func (*GetDataRequest) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{4}
}

func (x *GetDataRequest) GetStartkey() string {
	if x != nil {
		return x.Startkey
	}
	return ""
}

func (x *GetDataRequest) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type GetDataReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Nextkey string   `protobuf:"bytes,1,opt,name=nextkey,proto3" json:"nextkey,omitempty"`
	Keys    []string `protobuf:"bytes,2,rep,name=keys,proto3" json:"keys,omitempty"`
	Values  [][]byte `protobuf:"bytes,3,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *GetDataReply) Reset() {
	*x = GetDataReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDataReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataReply) ProtoMessage() {}

func (x *GetDataReply) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataReply.ProtoReflect.Descriptor instead.
func (*GetDataReply) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{5}
}

func (x *GetDataReply) GetNextkey() string {
	if x != nil {
		return x.Nextkey
	}
	return ""
}

func (x *GetDataReply) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *GetDataReply) GetValues() [][]byte {
	if x != nil {
		return x.Values
	}
	return nil
}

type GetCurrentOplogRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Startkey string `protobuf:"bytes,1,opt,name=startkey,proto3" json:"startkey,omitempty"`
	Length   int32  `protobuf:"varint,2,opt,name=length,proto3" json:"length,omitempty"`
}

func (x *GetCurrentOplogRequest) Reset() {
	*x = GetCurrentOplogRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCurrentOplogRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCurrentOplogRequest) ProtoMessage() {}

func (x *GetCurrentOplogRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCurrentOplogRequest.ProtoReflect.Descriptor instead.
func (*GetCurrentOplogRequest) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{6}
}

func (x *GetCurrentOplogRequest) GetStartkey() string {
	if x != nil {
		return x.Startkey
	}
	return ""
}

func (x *GetCurrentOplogRequest) GetLength() int32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type GetCurrentOplogReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys   []string `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	Values [][]byte `protobuf:"bytes,2,rep,name=values,proto3" json:"values,omitempty"`
}

func (x *GetCurrentOplogReply) Reset() {
	*x = GetCurrentOplogReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rpcreplica_rpcreplica_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetCurrentOplogReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetCurrentOplogReply) ProtoMessage() {}

func (x *GetCurrentOplogReply) ProtoReflect() protoreflect.Message {
	mi := &file_rpcreplica_rpcreplica_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetCurrentOplogReply.ProtoReflect.Descriptor instead.
func (*GetCurrentOplogReply) Descriptor() ([]byte, []int) {
	return file_rpcreplica_rpcreplica_proto_rawDescGZIP(), []int{7}
}

func (x *GetCurrentOplogReply) GetKeys() []string {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *GetCurrentOplogReply) GetValues() [][]byte {
	if x != nil {
		return x.Values
	}
	return nil
}

var File_rpcreplica_rpcreplica_proto protoreflect.FileDescriptor

var file_rpcreplica_rpcreplica_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2f, 0x72, 0x70, 0x63,
	0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x72,
	0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x22, 0x13, 0x0a, 0x11, 0x47, 0x65, 0x74,
	0x4c, 0x61, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x2b,
	0x0a, 0x0f, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x70, 0x6c,
	0x79, 0x12, 0x18, 0x0a, 0x07, 0x6c, 0x61, 0x73, 0x74, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x6c, 0x61, 0x73, 0x74, 0x6b, 0x65, 0x79, 0x22, 0x13, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x25, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x64, 0x52, 0x65,
	0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x44, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x22, 0x54, 0x0a,
	0x0c, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x18, 0x0a,
	0x07, 0x6e, 0x65, 0x78, 0x74, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6e, 0x65, 0x78, 0x74, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18,
	0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x73, 0x22, 0x4c, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e,
	0x74, 0x4f, 0x70, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x6b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e,
	0x67, 0x74, 0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x22, 0x42, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x4f,
	0x70, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x73, 0x32, 0xbf, 0x02, 0x0a, 0x07, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x12, 0x4a, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x4c, 0x61, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x12,
	0x1d, 0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65, 0x74,
	0x4c, 0x61, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b,
	0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65, 0x74, 0x4c,
	0x61, 0x73, 0x74, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x4a, 0x0a,
	0x0a, 0x47, 0x65, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6e, 0x65, 0x64, 0x12, 0x1d, 0x2e, 0x72, 0x70,
	0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x72, 0x61, 0x69,
	0x6e, 0x65, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x72, 0x70, 0x63,
	0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65, 0x74, 0x54, 0x72, 0x61, 0x69, 0x6e,
	0x65, 0x64, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x41, 0x0a, 0x07, 0x47, 0x65, 0x74,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x1a, 0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63,
	0x61, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x18, 0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65,
	0x74, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x59, 0x0a, 0x0f,
	0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x4f, 0x70, 0x6c, 0x6f, 0x67, 0x12,
	0x22, 0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x2e, 0x47, 0x65, 0x74,
	0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x4f, 0x70, 0x6c, 0x6f, 0x67, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x20, 0x2e, 0x72, 0x70, 0x63, 0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x2e, 0x47, 0x65, 0x74, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x4f, 0x70, 0x6c, 0x6f, 0x67,
	0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x42, 0x2e, 0x5a, 0x2c, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x72, 0x75, 0x6d, 0x62, 0x6a, 0x70, 0x2f, 0x66, 0x61,
	0x69, 0x73, 0x73, 0x64, 0x62, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x72, 0x70, 0x63,
	0x72, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rpcreplica_rpcreplica_proto_rawDescOnce sync.Once
	file_rpcreplica_rpcreplica_proto_rawDescData = file_rpcreplica_rpcreplica_proto_rawDesc
)

func file_rpcreplica_rpcreplica_proto_rawDescGZIP() []byte {
	file_rpcreplica_rpcreplica_proto_rawDescOnce.Do(func() {
		file_rpcreplica_rpcreplica_proto_rawDescData = protoimpl.X.CompressGZIP(file_rpcreplica_rpcreplica_proto_rawDescData)
	})
	return file_rpcreplica_rpcreplica_proto_rawDescData
}

var file_rpcreplica_rpcreplica_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_rpcreplica_rpcreplica_proto_goTypes = []interface{}{
	(*GetLastKeyRequest)(nil),      // 0: rpcreplica.GetLastKeyRequest
	(*GetLastKeyReply)(nil),        // 1: rpcreplica.GetLastKeyReply
	(*GetTrainedRequest)(nil),      // 2: rpcreplica.GetTrainedRequest
	(*GetTrainedReply)(nil),        // 3: rpcreplica.GetTrainedReply
	(*GetDataRequest)(nil),         // 4: rpcreplica.GetDataRequest
	(*GetDataReply)(nil),           // 5: rpcreplica.GetDataReply
	(*GetCurrentOplogRequest)(nil), // 6: rpcreplica.GetCurrentOplogRequest
	(*GetCurrentOplogReply)(nil),   // 7: rpcreplica.GetCurrentOplogReply
}
var file_rpcreplica_rpcreplica_proto_depIdxs = []int32{
	0, // 0: rpcreplica.Replica.GetLastKey:input_type -> rpcreplica.GetLastKeyRequest
	2, // 1: rpcreplica.Replica.GetTrained:input_type -> rpcreplica.GetTrainedRequest
	4, // 2: rpcreplica.Replica.GetData:input_type -> rpcreplica.GetDataRequest
	6, // 3: rpcreplica.Replica.GetCurrentOplog:input_type -> rpcreplica.GetCurrentOplogRequest
	1, // 4: rpcreplica.Replica.GetLastKey:output_type -> rpcreplica.GetLastKeyReply
	3, // 5: rpcreplica.Replica.GetTrained:output_type -> rpcreplica.GetTrainedReply
	5, // 6: rpcreplica.Replica.GetData:output_type -> rpcreplica.GetDataReply
	7, // 7: rpcreplica.Replica.GetCurrentOplog:output_type -> rpcreplica.GetCurrentOplogReply
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rpcreplica_rpcreplica_proto_init() }
func file_rpcreplica_rpcreplica_proto_init() {
	if File_rpcreplica_rpcreplica_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rpcreplica_rpcreplica_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetLastKeyRequest); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetLastKeyReply); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTrainedRequest); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTrainedReply); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDataRequest); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDataReply); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCurrentOplogRequest); i {
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
		file_rpcreplica_rpcreplica_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetCurrentOplogReply); i {
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
			RawDescriptor: file_rpcreplica_rpcreplica_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rpcreplica_rpcreplica_proto_goTypes,
		DependencyIndexes: file_rpcreplica_rpcreplica_proto_depIdxs,
		MessageInfos:      file_rpcreplica_rpcreplica_proto_msgTypes,
	}.Build()
	File_rpcreplica_rpcreplica_proto = out.File
	file_rpcreplica_rpcreplica_proto_rawDesc = nil
	file_rpcreplica_rpcreplica_proto_goTypes = nil
	file_rpcreplica_rpcreplica_proto_depIdxs = nil
}
