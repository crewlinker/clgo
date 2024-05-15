// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: clconnect/v1/rpc.proto

package clconnectv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

// the health check endpoint allows for inducing errors.
type InducedError int32

const (
	// when no induced error is specified
	InducedError_INDUCED_ERROR_UNSPECIFIED InducedError = 0
	// induces a unknown error
	InducedError_INDUCED_ERROR_UNKNOWN InducedError = 1
	// induces a panic
	InducedError_INDUCED_ERROR_PANIC InducedError = 2
)

// Enum value maps for InducedError.
var (
	InducedError_name = map[int32]string{
		0: "INDUCED_ERROR_UNSPECIFIED",
		1: "INDUCED_ERROR_UNKNOWN",
		2: "INDUCED_ERROR_PANIC",
	}
	InducedError_value = map[string]int32{
		"INDUCED_ERROR_UNSPECIFIED": 0,
		"INDUCED_ERROR_UNKNOWN":     1,
		"INDUCED_ERROR_PANIC":       2,
	}
)

func (x InducedError) Enum() *InducedError {
	p := new(InducedError)
	*p = x
	return p
}

func (x InducedError) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InducedError) Descriptor() protoreflect.EnumDescriptor {
	return file_clconnect_v1_rpc_proto_enumTypes[0].Descriptor()
}

func (InducedError) Type() protoreflect.EnumType {
	return &file_clconnect_v1_rpc_proto_enumTypes[0]
}

func (x InducedError) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InducedError.Descriptor instead.
func (InducedError) EnumDescriptor() ([]byte, []int) {
	return file_clconnect_v1_rpc_proto_rawDescGZIP(), []int{0}
}

// Request for checking health
type CheckHealthRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// echo will cause the respose to include the message
	Echo string `protobuf:"bytes,1,opt,name=echo,proto3" json:"echo,omitempty"`
	// induce error
	InduceError InducedError `protobuf:"varint,2,opt,name=induce_error,json=induceError,proto3,enum=clconnect.v1.InducedError" json:"induce_error,omitempty"`
}

func (x *CheckHealthRequest) Reset() {
	*x = CheckHealthRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clconnect_v1_rpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckHealthRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckHealthRequest) ProtoMessage() {}

func (x *CheckHealthRequest) ProtoReflect() protoreflect.Message {
	mi := &file_clconnect_v1_rpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckHealthRequest.ProtoReflect.Descriptor instead.
func (*CheckHealthRequest) Descriptor() ([]byte, []int) {
	return file_clconnect_v1_rpc_proto_rawDescGZIP(), []int{0}
}

func (x *CheckHealthRequest) GetEcho() string {
	if x != nil {
		return x.Echo
	}
	return ""
}

func (x *CheckHealthRequest) GetInduceError() InducedError {
	if x != nil {
		return x.InduceError
	}
	return InducedError_INDUCED_ERROR_UNSPECIFIED
}

// Reponse for for cheking ehalth
type CheckHealthResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// echo response
	Echo string `protobuf:"bytes,1,opt,name=echo,proto3" json:"echo,omitempty"`
}

func (x *CheckHealthResponse) Reset() {
	*x = CheckHealthResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clconnect_v1_rpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CheckHealthResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckHealthResponse) ProtoMessage() {}

func (x *CheckHealthResponse) ProtoReflect() protoreflect.Message {
	mi := &file_clconnect_v1_rpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CheckHealthResponse.ProtoReflect.Descriptor instead.
func (*CheckHealthResponse) Descriptor() ([]byte, []int) {
	return file_clconnect_v1_rpc_proto_rawDescGZIP(), []int{1}
}

func (x *CheckHealthResponse) GetEcho() string {
	if x != nil {
		return x.Echo
	}
	return ""
}

// Simple test request
type FooRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *FooRequest) Reset() {
	*x = FooRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clconnect_v1_rpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FooRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FooRequest) ProtoMessage() {}

func (x *FooRequest) ProtoReflect() protoreflect.Message {
	mi := &file_clconnect_v1_rpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FooRequest.ProtoReflect.Descriptor instead.
func (*FooRequest) Descriptor() ([]byte, []int) {
	return file_clconnect_v1_rpc_proto_rawDescGZIP(), []int{2}
}

// Simple test response
type FooResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// some response data
	Bar string `protobuf:"bytes,1,opt,name=bar,proto3" json:"bar,omitempty"`
}

func (x *FooResponse) Reset() {
	*x = FooResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_clconnect_v1_rpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FooResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FooResponse) ProtoMessage() {}

func (x *FooResponse) ProtoReflect() protoreflect.Message {
	mi := &file_clconnect_v1_rpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FooResponse.ProtoReflect.Descriptor instead.
func (*FooResponse) Descriptor() ([]byte, []int) {
	return file_clconnect_v1_rpc_proto_rawDescGZIP(), []int{3}
}

func (x *FooResponse) GetBar() string {
	if x != nil {
		return x.Bar
	}
	return ""
}

var File_clconnect_v1_rpc_proto protoreflect.FileDescriptor

var file_clconnect_v1_rpc_proto_rawDesc = []byte{
	0x0a, 0x16, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x72,
	0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e,
	0x65, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x70, 0x0a, 0x12, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c,
	0x74, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1b, 0x0a, 0x04, 0x65, 0x63, 0x68,
	0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xba, 0x48, 0x04, 0x72, 0x02, 0x10, 0x01,
	0x52, 0x04, 0x65, 0x63, 0x68, 0x6f, 0x12, 0x3d, 0x0a, 0x0c, 0x69, 0x6e, 0x64, 0x75, 0x63, 0x65,
	0x5f, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x63,
	0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x6e, 0x64, 0x75,
	0x63, 0x65, 0x64, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x0b, 0x69, 0x6e, 0x64, 0x75, 0x63, 0x65,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x29, 0x0a, 0x13, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x65, 0x63, 0x68, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x65, 0x63, 0x68, 0x6f,
	0x22, 0x0c, 0x0a, 0x0a, 0x46, 0x6f, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x1f,
	0x0a, 0x0b, 0x46, 0x6f, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x62, 0x61, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x62, 0x61, 0x72, 0x2a,
	0x61, 0x0a, 0x0c, 0x49, 0x6e, 0x64, 0x75, 0x63, 0x65, 0x64, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12,
	0x1d, 0x0a, 0x19, 0x49, 0x4e, 0x44, 0x55, 0x43, 0x45, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x19,
	0x0a, 0x15, 0x49, 0x4e, 0x44, 0x55, 0x43, 0x45, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f,
	0x55, 0x4e, 0x4b, 0x4e, 0x4f, 0x57, 0x4e, 0x10, 0x01, 0x12, 0x17, 0x0a, 0x13, 0x49, 0x4e, 0x44,
	0x55, 0x43, 0x45, 0x44, 0x5f, 0x45, 0x52, 0x52, 0x4f, 0x52, 0x5f, 0x50, 0x41, 0x4e, 0x49, 0x43,
	0x10, 0x02, 0x32, 0x4d, 0x0a, 0x0f, 0x52, 0x65, 0x61, 0x64, 0x4f, 0x6e, 0x6c, 0x79, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3a, 0x0a, 0x03, 0x46, 0x6f, 0x6f, 0x12, 0x18, 0x2e, 0x63,
	0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x6f, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x19, 0x2e, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x46, 0x6f, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x32, 0x66, 0x0a, 0x10, 0x52, 0x65, 0x61, 0x64, 0x57, 0x72, 0x69, 0x74, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x52, 0x0a, 0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65,
	0x61, 0x6c, 0x74, 0x68, 0x12, 0x20, 0x2e, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65,
	0x63, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61, 0x6c, 0x74,
	0x68, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0xa2, 0x01, 0x0a, 0x10, 0x63, 0x6f,
	0x6d, 0x2e, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x2e, 0x76, 0x31, 0x42, 0x08,
	0x52, 0x70, 0x63, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x72, 0x65, 0x77, 0x6c, 0x69, 0x6e, 0x6b, 0x65,
	0x72, 0x2f, 0x63, 0x6c, 0x67, 0x6f, 0x2f, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x2f, 0x76, 0x31, 0x3b, 0x63, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x76, 0x31, 0xa2,
	0x02, 0x03, 0x43, 0x58, 0x58, 0xaa, 0x02, 0x0c, 0x43, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63,
	0x74, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0c, 0x43, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x5c, 0x56, 0x31, 0xe2, 0x02, 0x18, 0x43, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x5c,
	0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02,
	0x0d, 0x43, 0x6c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_clconnect_v1_rpc_proto_rawDescOnce sync.Once
	file_clconnect_v1_rpc_proto_rawDescData = file_clconnect_v1_rpc_proto_rawDesc
)

func file_clconnect_v1_rpc_proto_rawDescGZIP() []byte {
	file_clconnect_v1_rpc_proto_rawDescOnce.Do(func() {
		file_clconnect_v1_rpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_clconnect_v1_rpc_proto_rawDescData)
	})
	return file_clconnect_v1_rpc_proto_rawDescData
}

var file_clconnect_v1_rpc_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_clconnect_v1_rpc_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_clconnect_v1_rpc_proto_goTypes = []interface{}{
	(InducedError)(0),           // 0: clconnect.v1.InducedError
	(*CheckHealthRequest)(nil),  // 1: clconnect.v1.CheckHealthRequest
	(*CheckHealthResponse)(nil), // 2: clconnect.v1.CheckHealthResponse
	(*FooRequest)(nil),          // 3: clconnect.v1.FooRequest
	(*FooResponse)(nil),         // 4: clconnect.v1.FooResponse
}
var file_clconnect_v1_rpc_proto_depIdxs = []int32{
	0, // 0: clconnect.v1.CheckHealthRequest.induce_error:type_name -> clconnect.v1.InducedError
	3, // 1: clconnect.v1.ReadOnlyService.Foo:input_type -> clconnect.v1.FooRequest
	1, // 2: clconnect.v1.ReadWriteService.CheckHealth:input_type -> clconnect.v1.CheckHealthRequest
	4, // 3: clconnect.v1.ReadOnlyService.Foo:output_type -> clconnect.v1.FooResponse
	2, // 4: clconnect.v1.ReadWriteService.CheckHealth:output_type -> clconnect.v1.CheckHealthResponse
	3, // [3:5] is the sub-list for method output_type
	1, // [1:3] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_clconnect_v1_rpc_proto_init() }
func file_clconnect_v1_rpc_proto_init() {
	if File_clconnect_v1_rpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_clconnect_v1_rpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckHealthRequest); i {
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
		file_clconnect_v1_rpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CheckHealthResponse); i {
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
		file_clconnect_v1_rpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FooRequest); i {
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
		file_clconnect_v1_rpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FooResponse); i {
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
			RawDescriptor: file_clconnect_v1_rpc_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_clconnect_v1_rpc_proto_goTypes,
		DependencyIndexes: file_clconnect_v1_rpc_proto_depIdxs,
		EnumInfos:         file_clconnect_v1_rpc_proto_enumTypes,
		MessageInfos:      file_clconnect_v1_rpc_proto_msgTypes,
	}.Build()
	File_clconnect_v1_rpc_proto = out.File
	file_clconnect_v1_rpc_proto_rawDesc = nil
	file_clconnect_v1_rpc_proto_goTypes = nil
	file_clconnect_v1_rpc_proto_depIdxs = nil
}
