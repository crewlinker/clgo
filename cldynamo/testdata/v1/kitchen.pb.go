// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: cldynamo/testdata/v1/kitchen.proto

package testdatav1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Kind encodes the entity kind in the database.
type FridgeBrand int32

const (
	// when no fridge brand is specified
	FridgeBrand_FRIDGE_BRAND_UNSPECIFIED FridgeBrand = 0
	// siemens brand
	FridgeBrand_FRIDGE_BRAND_SIEMENS FridgeBrand = 1
)

// Enum value maps for FridgeBrand.
var (
	FridgeBrand_name = map[int32]string{
		0: "FRIDGE_BRAND_UNSPECIFIED",
		1: "FRIDGE_BRAND_SIEMENS",
	}
	FridgeBrand_value = map[string]int32{
		"FRIDGE_BRAND_UNSPECIFIED": 0,
		"FRIDGE_BRAND_SIEMENS":     1,
	}
)

func (x FridgeBrand) Enum() *FridgeBrand {
	p := new(FridgeBrand)
	*p = x
	return p
}

func (x FridgeBrand) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (FridgeBrand) Descriptor() protoreflect.EnumDescriptor {
	return file_cldynamo_testdata_v1_kitchen_proto_enumTypes[0].Descriptor()
}

func (FridgeBrand) Type() protoreflect.EnumType {
	return &file_cldynamo_testdata_v1_kitchen_proto_enumTypes[0]
}

func (x FridgeBrand) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use FridgeBrand.Descriptor instead.
func (FridgeBrand) EnumDescriptor() ([]byte, []int) {
	return file_cldynamo_testdata_v1_kitchen_proto_rawDescGZIP(), []int{0}
}

// User describes a user on our platform
type Kitchen struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// id of the kitchen
	KitchenId string `protobuf:"bytes,1,opt,name=kitchen_id,json=kitchenId,proto3" json:"kitchen_id,omitempty"`
	// the brand of fridge
	FridgeBrand FridgeBrand `protobuf:"varint,2,opt,name=fridge_brand,json=fridgeBrand,proto3,enum=cldynamo.testdata.v1.FridgeBrand" json:"fridge_brand,omitempty"`
	// one of the tiling style
	//
	// Types that are assignable to TilingStyle:
	//
	//	*Kitchen_Terracotta
	//	*Kitchen_OtherStyle
	//	*Kitchen_YetAnother
	TilingStyle isKitchen_TilingStyle `protobuf_oneof:"tiling_style"`
	// another kitchen
	Another *Kitchen `protobuf:"bytes,4,opt,name=another,proto3" json:"another,omitempty"`
	// other nested kitchens
	Others []*Kitchen `protobuf:"bytes,5,rep,name=others,proto3" json:"others,omitempty"`
	// map with recursive type
	MapOthers map[int64]*Kitchen `protobuf:"bytes,9,rep,name=map_others,json=mapOthers,proto3" json:"map_others,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// field with imported type
	BuildAt *timestamppb.Timestamp `protobuf:"bytes,10,opt,name=build_at,json=buildAt,proto3" json:"build_at,omitempty"`
}

func (x *Kitchen) Reset() {
	*x = Kitchen{}
	if protoimpl.UnsafeEnabled {
		mi := &file_cldynamo_testdata_v1_kitchen_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Kitchen) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Kitchen) ProtoMessage() {}

func (x *Kitchen) ProtoReflect() protoreflect.Message {
	mi := &file_cldynamo_testdata_v1_kitchen_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Kitchen.ProtoReflect.Descriptor instead.
func (*Kitchen) Descriptor() ([]byte, []int) {
	return file_cldynamo_testdata_v1_kitchen_proto_rawDescGZIP(), []int{0}
}

func (x *Kitchen) GetKitchenId() string {
	if x != nil {
		return x.KitchenId
	}
	return ""
}

func (x *Kitchen) GetFridgeBrand() FridgeBrand {
	if x != nil {
		return x.FridgeBrand
	}
	return FridgeBrand_FRIDGE_BRAND_UNSPECIFIED
}

func (m *Kitchen) GetTilingStyle() isKitchen_TilingStyle {
	if m != nil {
		return m.TilingStyle
	}
	return nil
}

func (x *Kitchen) GetTerracotta() string {
	if x, ok := x.GetTilingStyle().(*Kitchen_Terracotta); ok {
		return x.Terracotta
	}
	return ""
}

func (x *Kitchen) GetOtherStyle() int64 {
	if x, ok := x.GetTilingStyle().(*Kitchen_OtherStyle); ok {
		return x.OtherStyle
	}
	return 0
}

func (x *Kitchen) GetYetAnother() *Kitchen {
	if x, ok := x.GetTilingStyle().(*Kitchen_YetAnother); ok {
		return x.YetAnother
	}
	return nil
}

func (x *Kitchen) GetAnother() *Kitchen {
	if x != nil {
		return x.Another
	}
	return nil
}

func (x *Kitchen) GetOthers() []*Kitchen {
	if x != nil {
		return x.Others
	}
	return nil
}

func (x *Kitchen) GetMapOthers() map[int64]*Kitchen {
	if x != nil {
		return x.MapOthers
	}
	return nil
}

func (x *Kitchen) GetBuildAt() *timestamppb.Timestamp {
	if x != nil {
		return x.BuildAt
	}
	return nil
}

type isKitchen_TilingStyle interface {
	isKitchen_TilingStyle()
}

type Kitchen_Terracotta struct {
	// terracotta style
	Terracotta string `protobuf:"bytes,6,opt,name=terracotta,proto3,oneof"`
}

type Kitchen_OtherStyle struct {
	// some other style
	OtherStyle int64 `protobuf:"varint,7,opt,name=other_style,json=otherStyle,proto3,oneof"`
}

type Kitchen_YetAnother struct {
	// nested message in oneof
	YetAnother *Kitchen `protobuf:"bytes,8,opt,name=yet_another,json=yetAnother,proto3,oneof"`
}

func (*Kitchen_Terracotta) isKitchen_TilingStyle() {}

func (*Kitchen_OtherStyle) isKitchen_TilingStyle() {}

func (*Kitchen_YetAnother) isKitchen_TilingStyle() {}

var File_cldynamo_testdata_v1_kitchen_proto protoreflect.FileDescriptor

var file_cldynamo_testdata_v1_kitchen_proto_rawDesc = []byte{
	0x0a, 0x22, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x64,
	0x61, 0x74, 0x61, 0x2f, 0x76, 0x31, 0x2f, 0x6b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x74,
	0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65,
	0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xd6, 0x04, 0x0a, 0x07,
	0x4b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x12, 0x1d, 0x0a, 0x0a, 0x6b, 0x69, 0x74, 0x63, 0x68,
	0x65, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x6b, 0x69, 0x74,
	0x63, 0x68, 0x65, 0x6e, 0x49, 0x64, 0x12, 0x44, 0x0a, 0x0c, 0x66, 0x72, 0x69, 0x64, 0x67, 0x65,
	0x5f, 0x62, 0x72, 0x61, 0x6e, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63,
	0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x76, 0x31, 0x2e, 0x46, 0x72, 0x69, 0x64, 0x67, 0x65, 0x42, 0x72, 0x61, 0x6e, 0x64, 0x52,
	0x0b, 0x66, 0x72, 0x69, 0x64, 0x67, 0x65, 0x42, 0x72, 0x61, 0x6e, 0x64, 0x12, 0x20, 0x0a, 0x0a,
	0x74, 0x65, 0x72, 0x72, 0x61, 0x63, 0x6f, 0x74, 0x74, 0x61, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x48, 0x00, 0x52, 0x0a, 0x74, 0x65, 0x72, 0x72, 0x61, 0x63, 0x6f, 0x74, 0x74, 0x61, 0x12, 0x21,
	0x0a, 0x0b, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x5f, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x03, 0x48, 0x00, 0x52, 0x0a, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x53, 0x74, 0x79, 0x6c,
	0x65, 0x12, 0x40, 0x0a, 0x0b, 0x79, 0x65, 0x74, 0x5f, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d,
	0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x69,
	0x74, 0x63, 0x68, 0x65, 0x6e, 0x48, 0x00, 0x52, 0x0a, 0x79, 0x65, 0x74, 0x41, 0x6e, 0x6f, 0x74,
	0x68, 0x65, 0x72, 0x12, 0x37, 0x0a, 0x07, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e,
	0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x69, 0x74, 0x63,
	0x68, 0x65, 0x6e, 0x52, 0x07, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x12, 0x35, 0x0a, 0x06,
	0x6f, 0x74, 0x68, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63,
	0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61,
	0x2e, 0x76, 0x31, 0x2e, 0x4b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x52, 0x06, 0x6f, 0x74, 0x68,
	0x65, 0x72, 0x73, 0x12, 0x4b, 0x0a, 0x0a, 0x6d, 0x61, 0x70, 0x5f, 0x6f, 0x74, 0x68, 0x65, 0x72,
	0x73, 0x18, 0x09, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x2c, 0x2e, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61,
	0x6d, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x2e, 0x4b,
	0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x2e, 0x4d, 0x61, 0x70, 0x4f, 0x74, 0x68, 0x65, 0x72, 0x73,
	0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x09, 0x6d, 0x61, 0x70, 0x4f, 0x74, 0x68, 0x65, 0x72, 0x73,
	0x12, 0x35, 0x0a, 0x08, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x61, 0x74, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x07,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x41, 0x74, 0x1a, 0x5b, 0x0a, 0x0e, 0x4d, 0x61, 0x70, 0x4f, 0x74,
	0x68, 0x65, 0x72, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x33, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x64,
	0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76,
	0x31, 0x2e, 0x4b, 0x69, 0x74, 0x63, 0x68, 0x65, 0x6e, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x3a, 0x02, 0x38, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x74, 0x69, 0x6c, 0x69, 0x6e, 0x67, 0x5f, 0x73,
	0x74, 0x79, 0x6c, 0x65, 0x2a, 0x45, 0x0a, 0x0b, 0x46, 0x72, 0x69, 0x64, 0x67, 0x65, 0x42, 0x72,
	0x61, 0x6e, 0x64, 0x12, 0x1c, 0x0a, 0x18, 0x46, 0x52, 0x49, 0x44, 0x47, 0x45, 0x5f, 0x42, 0x52,
	0x41, 0x4e, 0x44, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10,
	0x00, 0x12, 0x18, 0x0a, 0x14, 0x46, 0x52, 0x49, 0x44, 0x47, 0x45, 0x5f, 0x42, 0x52, 0x41, 0x4e,
	0x44, 0x5f, 0x53, 0x49, 0x45, 0x4d, 0x45, 0x4e, 0x53, 0x10, 0x01, 0x42, 0xd6, 0x01, 0x0a, 0x18,
	0x63, 0x6f, 0x6d, 0x2e, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x74, 0x65, 0x73,
	0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x76, 0x31, 0x42, 0x0c, 0x4b, 0x69, 0x74, 0x63, 0x68, 0x65,
	0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x72, 0x65, 0x77, 0x6c, 0x69, 0x6e, 0x6b, 0x65, 0x72, 0x2f,
	0x63, 0x6c, 0x67, 0x6f, 0x2f, 0x63, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x76, 0x31, 0x3b, 0x74, 0x65, 0x73, 0x74, 0x64, 0x61,
	0x74, 0x61, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x43, 0x54, 0x58, 0xaa, 0x02, 0x14, 0x43, 0x6c, 0x64,
	0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x2e, 0x54, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x14, 0x43, 0x6c, 0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x5c, 0x54, 0x65, 0x73,
	0x74, 0x64, 0x61, 0x74, 0x61, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x20, 0x43, 0x6c, 0x64, 0x79, 0x6e,
	0x61, 0x6d, 0x6f, 0x5c, 0x54, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x16, 0x43, 0x6c,
	0x64, 0x79, 0x6e, 0x61, 0x6d, 0x6f, 0x3a, 0x3a, 0x54, 0x65, 0x73, 0x74, 0x64, 0x61, 0x74, 0x61,
	0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_cldynamo_testdata_v1_kitchen_proto_rawDescOnce sync.Once
	file_cldynamo_testdata_v1_kitchen_proto_rawDescData = file_cldynamo_testdata_v1_kitchen_proto_rawDesc
)

func file_cldynamo_testdata_v1_kitchen_proto_rawDescGZIP() []byte {
	file_cldynamo_testdata_v1_kitchen_proto_rawDescOnce.Do(func() {
		file_cldynamo_testdata_v1_kitchen_proto_rawDescData = protoimpl.X.CompressGZIP(file_cldynamo_testdata_v1_kitchen_proto_rawDescData)
	})
	return file_cldynamo_testdata_v1_kitchen_proto_rawDescData
}

var file_cldynamo_testdata_v1_kitchen_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_cldynamo_testdata_v1_kitchen_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_cldynamo_testdata_v1_kitchen_proto_goTypes = []interface{}{
	(FridgeBrand)(0),              // 0: cldynamo.testdata.v1.FridgeBrand
	(*Kitchen)(nil),               // 1: cldynamo.testdata.v1.Kitchen
	nil,                           // 2: cldynamo.testdata.v1.Kitchen.MapOthersEntry
	(*timestamppb.Timestamp)(nil), // 3: google.protobuf.Timestamp
}
var file_cldynamo_testdata_v1_kitchen_proto_depIdxs = []int32{
	0, // 0: cldynamo.testdata.v1.Kitchen.fridge_brand:type_name -> cldynamo.testdata.v1.FridgeBrand
	1, // 1: cldynamo.testdata.v1.Kitchen.yet_another:type_name -> cldynamo.testdata.v1.Kitchen
	1, // 2: cldynamo.testdata.v1.Kitchen.another:type_name -> cldynamo.testdata.v1.Kitchen
	1, // 3: cldynamo.testdata.v1.Kitchen.others:type_name -> cldynamo.testdata.v1.Kitchen
	2, // 4: cldynamo.testdata.v1.Kitchen.map_others:type_name -> cldynamo.testdata.v1.Kitchen.MapOthersEntry
	3, // 5: cldynamo.testdata.v1.Kitchen.build_at:type_name -> google.protobuf.Timestamp
	1, // 6: cldynamo.testdata.v1.Kitchen.MapOthersEntry.value:type_name -> cldynamo.testdata.v1.Kitchen
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_cldynamo_testdata_v1_kitchen_proto_init() }
func file_cldynamo_testdata_v1_kitchen_proto_init() {
	if File_cldynamo_testdata_v1_kitchen_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_cldynamo_testdata_v1_kitchen_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Kitchen); i {
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
	file_cldynamo_testdata_v1_kitchen_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Kitchen_Terracotta)(nil),
		(*Kitchen_OtherStyle)(nil),
		(*Kitchen_YetAnother)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_cldynamo_testdata_v1_kitchen_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_cldynamo_testdata_v1_kitchen_proto_goTypes,
		DependencyIndexes: file_cldynamo_testdata_v1_kitchen_proto_depIdxs,
		EnumInfos:         file_cldynamo_testdata_v1_kitchen_proto_enumTypes,
		MessageInfos:      file_cldynamo_testdata_v1_kitchen_proto_msgTypes,
	}.Build()
	File_cldynamo_testdata_v1_kitchen_proto = out.File
	file_cldynamo_testdata_v1_kitchen_proto_rawDesc = nil
	file_cldynamo_testdata_v1_kitchen_proto_goTypes = nil
	file_cldynamo_testdata_v1_kitchen_proto_depIdxs = nil
}
