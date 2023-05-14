// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0-devel
// 	protoc        v3.21.12
// source: geometry.proto

package proto

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

type PolylineProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Points       []*PointProto `protobuf:"bytes,1,rep,name=points,proto3" json:"points,omitempty"`
	LengthMeters float64       `protobuf:"fixed64,2,opt,name=length_meters,json=lengthMeters,proto3" json:"length_meters,omitempty"`
}

func (x *PolylineProto) Reset() {
	*x = PolylineProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geometry_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PolylineProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolylineProto) ProtoMessage() {}

func (x *PolylineProto) ProtoReflect() protoreflect.Message {
	mi := &file_geometry_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolylineProto.ProtoReflect.Descriptor instead.
func (*PolylineProto) Descriptor() ([]byte, []int) {
	return file_geometry_proto_rawDescGZIP(), []int{0}
}

func (x *PolylineProto) GetPoints() []*PointProto {
	if x != nil {
		return x.Points
	}
	return nil
}

func (x *PolylineProto) GetLengthMeters() float64 {
	if x != nil {
		return x.LengthMeters
	}
	return 0
}

type MultiPolygonProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Polygons []*PolygonProto `protobuf:"bytes,1,rep,name=polygons,proto3" json:"polygons,omitempty"`
}

func (x *MultiPolygonProto) Reset() {
	*x = MultiPolygonProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geometry_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MultiPolygonProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MultiPolygonProto) ProtoMessage() {}

func (x *MultiPolygonProto) ProtoReflect() protoreflect.Message {
	mi := &file_geometry_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MultiPolygonProto.ProtoReflect.Descriptor instead.
func (*MultiPolygonProto) Descriptor() ([]byte, []int) {
	return file_geometry_proto_rawDescGZIP(), []int{1}
}

func (x *MultiPolygonProto) GetPolygons() []*PolygonProto {
	if x != nil {
		return x.Polygons
	}
	return nil
}

type PolygonProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// All loops are ordered counter-clockwise, and a point is defined to be
	// inside the polygon if it's enclosed by an odd number of loops.
	Loops []*LoopProto `protobuf:"bytes,1,rep,name=loops,proto3" json:"loops,omitempty"`
}

func (x *PolygonProto) Reset() {
	*x = PolygonProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geometry_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PolygonProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PolygonProto) ProtoMessage() {}

func (x *PolygonProto) ProtoReflect() protoreflect.Message {
	mi := &file_geometry_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PolygonProto.ProtoReflect.Descriptor instead.
func (*PolygonProto) Descriptor() ([]byte, []int) {
	return file_geometry_proto_rawDescGZIP(), []int{2}
}

func (x *PolygonProto) GetLoops() []*LoopProto {
	if x != nil {
		return x.Loops
	}
	return nil
}

type LoopProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Points []*PointProto `protobuf:"bytes,1,rep,name=points,proto3" json:"points,omitempty"`
}

func (x *LoopProto) Reset() {
	*x = LoopProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geometry_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LoopProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LoopProto) ProtoMessage() {}

func (x *LoopProto) ProtoReflect() protoreflect.Message {
	mi := &file_geometry_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LoopProto.ProtoReflect.Descriptor instead.
func (*LoopProto) Descriptor() ([]byte, []int) {
	return file_geometry_proto_rawDescGZIP(), []int{3}
}

func (x *LoopProto) GetPoints() []*PointProto {
	if x != nil {
		return x.Points
	}
	return nil
}

type PointProto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LatE7 int32 `protobuf:"varint,1,opt,name=lat_e7,json=latE7,proto3" json:"lat_e7,omitempty"`
	LngE7 int32 `protobuf:"varint,2,opt,name=lng_e7,json=lngE7,proto3" json:"lng_e7,omitempty"`
}

func (x *PointProto) Reset() {
	*x = PointProto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_geometry_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PointProto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PointProto) ProtoMessage() {}

func (x *PointProto) ProtoReflect() protoreflect.Message {
	mi := &file_geometry_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PointProto.ProtoReflect.Descriptor instead.
func (*PointProto) Descriptor() ([]byte, []int) {
	return file_geometry_proto_rawDescGZIP(), []int{4}
}

func (x *PointProto) GetLatE7() int32 {
	if x != nil {
		return x.LatE7
	}
	return 0
}

func (x *PointProto) GetLngE7() int32 {
	if x != nil {
		return x.LngE7
	}
	return 0
}

var File_geometry_proto protoreflect.FileDescriptor

var file_geometry_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x08, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x22, 0x62, 0x0a, 0x0d, 0x50, 0x6f,
	0x6c, 0x79, 0x6c, 0x69, 0x6e, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x2c, 0x0a, 0x06, 0x70,
	0x6f, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x67, 0x65,
	0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x74,
	0x6f, 0x52, 0x06, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x23, 0x0a, 0x0d, 0x6c, 0x65, 0x6e,
	0x67, 0x74, 0x68, 0x5f, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01,
	0x52, 0x0c, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x4d, 0x65, 0x74, 0x65, 0x72, 0x73, 0x22, 0x47,
	0x0a, 0x11, 0x4d, 0x75, 0x6c, 0x74, 0x69, 0x50, 0x6f, 0x6c, 0x79, 0x67, 0x6f, 0x6e, 0x50, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x32, 0x0a, 0x08, 0x70, 0x6f, 0x6c, 0x79, 0x67, 0x6f, 0x6e, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79,
	0x2e, 0x50, 0x6f, 0x6c, 0x79, 0x67, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x08, 0x70,
	0x6f, 0x6c, 0x79, 0x67, 0x6f, 0x6e, 0x73, 0x22, 0x39, 0x0a, 0x0c, 0x50, 0x6f, 0x6c, 0x79, 0x67,
	0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x29, 0x0a, 0x05, 0x6c, 0x6f, 0x6f, 0x70, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72,
	0x79, 0x2e, 0x4c, 0x6f, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x05, 0x6c, 0x6f, 0x6f,
	0x70, 0x73, 0x22, 0x39, 0x0a, 0x09, 0x4c, 0x6f, 0x6f, 0x70, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x2c, 0x0a, 0x06, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x14, 0x2e, 0x67, 0x65, 0x6f, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x50, 0x6f, 0x69, 0x6e, 0x74,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x06, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x22, 0x3a, 0x0a,
	0x0a, 0x50, 0x6f, 0x69, 0x6e, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x0a, 0x06, 0x6c,
	0x61, 0x74, 0x5f, 0x65, 0x37, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x6c, 0x61, 0x74,
	0x45, 0x37, 0x12, 0x15, 0x0a, 0x06, 0x6c, 0x6e, 0x67, 0x5f, 0x65, 0x37, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x05, 0x6c, 0x6e, 0x67, 0x45, 0x37, 0x42, 0x19, 0x5a, 0x17, 0x64, 0x69, 0x61,
	0x67, 0x6f, 0x6e, 0x61, 0x6c, 0x2e, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x2f, 0x62, 0x36, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_geometry_proto_rawDescOnce sync.Once
	file_geometry_proto_rawDescData = file_geometry_proto_rawDesc
)

func file_geometry_proto_rawDescGZIP() []byte {
	file_geometry_proto_rawDescOnce.Do(func() {
		file_geometry_proto_rawDescData = protoimpl.X.CompressGZIP(file_geometry_proto_rawDescData)
	})
	return file_geometry_proto_rawDescData
}

var file_geometry_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_geometry_proto_goTypes = []interface{}{
	(*PolylineProto)(nil),     // 0: geometry.PolylineProto
	(*MultiPolygonProto)(nil), // 1: geometry.MultiPolygonProto
	(*PolygonProto)(nil),      // 2: geometry.PolygonProto
	(*LoopProto)(nil),         // 3: geometry.LoopProto
	(*PointProto)(nil),        // 4: geometry.PointProto
}
var file_geometry_proto_depIdxs = []int32{
	4, // 0: geometry.PolylineProto.points:type_name -> geometry.PointProto
	2, // 1: geometry.MultiPolygonProto.polygons:type_name -> geometry.PolygonProto
	3, // 2: geometry.PolygonProto.loops:type_name -> geometry.LoopProto
	4, // 3: geometry.LoopProto.points:type_name -> geometry.PointProto
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_geometry_proto_init() }
func file_geometry_proto_init() {
	if File_geometry_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_geometry_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PolylineProto); i {
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
		file_geometry_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MultiPolygonProto); i {
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
		file_geometry_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PolygonProto); i {
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
		file_geometry_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LoopProto); i {
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
		file_geometry_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PointProto); i {
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
			RawDescriptor: file_geometry_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_geometry_proto_goTypes,
		DependencyIndexes: file_geometry_proto_depIdxs,
		MessageInfos:      file_geometry_proto_msgTypes,
	}.Build()
	File_geometry_proto = out.File
	file_geometry_proto_rawDesc = nil
	file_geometry_proto_goTypes = nil
	file_geometry_proto_depIdxs = nil
}