// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.28.3
// source: idgeneration.proto

package gen

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

type IDGenerator struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Entity string `protobuf:"bytes,2,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *IDGenerator) Reset() {
	*x = IDGenerator{}
	mi := &file_idgeneration_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IDGenerator) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IDGenerator) ProtoMessage() {}

func (x *IDGenerator) ProtoReflect() protoreflect.Message {
	mi := &file_idgeneration_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IDGenerator.ProtoReflect.Descriptor instead.
func (*IDGenerator) Descriptor() ([]byte, []int) {
	return file_idgeneration_proto_rawDescGZIP(), []int{0}
}

func (x *IDGenerator) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *IDGenerator) GetEntity() string {
	if x != nil {
		return x.Entity
	}
	return ""
}

type IdGenerationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Entity string `protobuf:"bytes,1,opt,name=entity,proto3" json:"entity,omitempty"`
}

func (x *IdGenerationRequest) Reset() {
	*x = IdGenerationRequest{}
	mi := &file_idgeneration_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IdGenerationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdGenerationRequest) ProtoMessage() {}

func (x *IdGenerationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_idgeneration_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdGenerationRequest.ProtoReflect.Descriptor instead.
func (*IdGenerationRequest) Descriptor() ([]byte, []int) {
	return file_idgeneration_proto_rawDescGZIP(), []int{1}
}

func (x *IdGenerationRequest) GetEntity() string {
	if x != nil {
		return x.Entity
	}
	return ""
}

type IdGenerationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IdGenerator *IDGenerator `protobuf:"bytes,1,opt,name=idGenerator,proto3" json:"idGenerator,omitempty"`
}

func (x *IdGenerationResponse) Reset() {
	*x = IdGenerationResponse{}
	mi := &file_idgeneration_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *IdGenerationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IdGenerationResponse) ProtoMessage() {}

func (x *IdGenerationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_idgeneration_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IdGenerationResponse.ProtoReflect.Descriptor instead.
func (*IdGenerationResponse) Descriptor() ([]byte, []int) {
	return file_idgeneration_proto_rawDescGZIP(), []int{2}
}

func (x *IdGenerationResponse) GetIdGenerator() *IDGenerator {
	if x != nil {
		return x.IdGenerator
	}
	return nil
}

var File_idgeneration_proto protoreflect.FileDescriptor

var file_idgeneration_proto_rawDesc = []byte{
	0x0a, 0x12, 0x69, 0x64, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x35, 0x0a, 0x0b, 0x49, 0x44, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61,
	0x74, 0x6f, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x2d, 0x0a, 0x13, 0x49,
	0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x22, 0x46, 0x0a, 0x14, 0x49, 0x64,
	0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x2e, 0x0a, 0x0b, 0x69, 0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x6f,
	0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x49, 0x44, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x6f, 0x72, 0x52, 0x0b, 0x69, 0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74,
	0x6f, 0x72, 0x32, 0x50, 0x0a, 0x13, 0x49, 0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x47, 0x65, 0x6e,
	0x65, 0x72, 0x61, 0x74, 0x65, 0x49, 0x64, 0x12, 0x14, 0x2e, 0x49, 0x64, 0x47, 0x65, 0x6e, 0x65,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x49, 0x64, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x06, 0x5a, 0x04, 0x2f, 0x67, 0x65, 0x6e, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_idgeneration_proto_rawDescOnce sync.Once
	file_idgeneration_proto_rawDescData = file_idgeneration_proto_rawDesc
)

func file_idgeneration_proto_rawDescGZIP() []byte {
	file_idgeneration_proto_rawDescOnce.Do(func() {
		file_idgeneration_proto_rawDescData = protoimpl.X.CompressGZIP(file_idgeneration_proto_rawDescData)
	})
	return file_idgeneration_proto_rawDescData
}

var file_idgeneration_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_idgeneration_proto_goTypes = []any{
	(*IDGenerator)(nil),          // 0: IDGenerator
	(*IdGenerationRequest)(nil),  // 1: IdGenerationRequest
	(*IdGenerationResponse)(nil), // 2: IdGenerationResponse
}
var file_idgeneration_proto_depIdxs = []int32{
	0, // 0: IdGenerationResponse.idGenerator:type_name -> IDGenerator
	1, // 1: IdGenerationService.GenerateId:input_type -> IdGenerationRequest
	2, // 2: IdGenerationService.GenerateId:output_type -> IdGenerationResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_idgeneration_proto_init() }
func file_idgeneration_proto_init() {
	if File_idgeneration_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_idgeneration_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_idgeneration_proto_goTypes,
		DependencyIndexes: file_idgeneration_proto_depIdxs,
		MessageInfos:      file_idgeneration_proto_msgTypes,
	}.Build()
	File_idgeneration_proto = out.File
	file_idgeneration_proto_rawDesc = nil
	file_idgeneration_proto_goTypes = nil
	file_idgeneration_proto_depIdxs = nil
}
