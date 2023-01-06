// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: pkg/service/rpc/ingress.proto

package rpc

import (
	livekit "github.com/livekit/protocol/livekit"
	_ "github.com/livekit/psrpc/protoc-gen-psrpc/options"
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

type Ignored struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Ignored) Reset() {
	*x = Ignored{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_service_rpc_ingress_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Ignored) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ignored) ProtoMessage() {}

func (x *Ignored) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_rpc_ingress_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ignored.ProtoReflect.Descriptor instead.
func (*Ignored) Descriptor() ([]byte, []int) {
	return file_pkg_service_rpc_ingress_proto_rawDescGZIP(), []int{0}
}

type ListActiveIngressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ListActiveIngressRequest) Reset() {
	*x = ListActiveIngressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_service_rpc_ingress_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListActiveIngressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListActiveIngressRequest) ProtoMessage() {}

func (x *ListActiveIngressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_rpc_ingress_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListActiveIngressRequest.ProtoReflect.Descriptor instead.
func (*ListActiveIngressRequest) Descriptor() ([]byte, []int) {
	return file_pkg_service_rpc_ingress_proto_rawDescGZIP(), []int{1}
}

type ListActiveIngressResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IngressIds []string `protobuf:"bytes,1,rep,name=ingress_ids,json=ingressIds,proto3" json:"ingress_ids,omitempty"`
}

func (x *ListActiveIngressResponse) Reset() {
	*x = ListActiveIngressResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_service_rpc_ingress_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListActiveIngressResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListActiveIngressResponse) ProtoMessage() {}

func (x *ListActiveIngressResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_rpc_ingress_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListActiveIngressResponse.ProtoReflect.Descriptor instead.
func (*ListActiveIngressResponse) Descriptor() ([]byte, []int) {
	return file_pkg_service_rpc_ingress_proto_rawDescGZIP(), []int{2}
}

func (x *ListActiveIngressResponse) GetIngressIds() []string {
	if x != nil {
		return x.IngressIds
	}
	return nil
}

type HangUpIngressRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HangUpIngressRequest) Reset() {
	*x = HangUpIngressRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_service_rpc_ingress_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HangUpIngressRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HangUpIngressRequest) ProtoMessage() {}

func (x *HangUpIngressRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_rpc_ingress_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HangUpIngressRequest.ProtoReflect.Descriptor instead.
func (*HangUpIngressRequest) Descriptor() ([]byte, []int) {
	return file_pkg_service_rpc_ingress_proto_rawDescGZIP(), []int{3}
}

type HangUpIngressResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *HangUpIngressResponse) Reset() {
	*x = HangUpIngressResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_service_rpc_ingress_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *HangUpIngressResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HangUpIngressResponse) ProtoMessage() {}

func (x *HangUpIngressResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_service_rpc_ingress_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HangUpIngressResponse.ProtoReflect.Descriptor instead.
func (*HangUpIngressResponse) Descriptor() ([]byte, []int) {
	return file_pkg_service_rpc_ingress_proto_rawDescGZIP(), []int{4}
}

var File_pkg_service_rpc_ingress_proto protoreflect.FileDescriptor

var file_pkg_service_rpc_ingress_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x72, 0x70,
	0x63, 0x2f, 0x69, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x03, 0x72, 0x70, 0x63, 0x1a, 0x0d, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x1a, 0x1a, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x5f, 0x72, 0x70, 0x63,
	0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x15, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x5f, 0x69, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x09, 0x0a, 0x07, 0x49, 0x67, 0x6e, 0x6f, 0x72, 0x65,
	0x64, 0x22, 0x1a, 0x0a, 0x18, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x49,
	0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x3c, 0x0a,
	0x19, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x49, 0x6e, 0x67, 0x72, 0x65,
	0x73, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e,
	0x67, 0x72, 0x65, 0x73, 0x73, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0a, 0x69, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x64, 0x73, 0x22, 0x16, 0x0a, 0x14, 0x48,
	0x61, 0x6e, 0x67, 0x55, 0x70, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x17, 0x0a, 0x15, 0x48, 0x61, 0x6e, 0x67, 0x55, 0x70, 0x49, 0x6e, 0x67,
	0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x6d, 0x0a, 0x0f,
	0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x12,
	0x5a, 0x0a, 0x11, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x63, 0x74, 0x69, 0x76, 0x65, 0x49, 0x6e, 0x67,
	0x72, 0x65, 0x73, 0x73, 0x12, 0x1d, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41,
	0x63, 0x74, 0x69, 0x76, 0x65, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x41, 0x63,
	0x74, 0x69, 0x76, 0x65, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x06, 0xb2, 0x89, 0x01, 0x02, 0x08, 0x01, 0x32, 0x62, 0x0a, 0x0d, 0x49,
	0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x45, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x12, 0x51, 0x0a, 0x0e,
	0x47, 0x65, 0x74, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1e,
	0x2e, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f,
	0x2e, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x67, 0x72,
	0x65, 0x73, 0x73, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32,
	0x60, 0x0a, 0x0e, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x48, 0x61, 0x6e, 0x64, 0x6c, 0x65,
	0x72, 0x12, 0x4e, 0x0a, 0x0d, 0x48, 0x61, 0x6e, 0x67, 0x55, 0x70, 0x49, 0x6e, 0x67, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x19, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x48, 0x61, 0x6e, 0x67, 0x55, 0x70, 0x49,
	0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1a, 0x2e,
	0x72, 0x70, 0x63, 0x2e, 0x48, 0x61, 0x6e, 0x67, 0x55, 0x70, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x06, 0xb2, 0x89, 0x01, 0x02, 0x18,
	0x01, 0x32, 0x58, 0x0a, 0x0d, 0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x12, 0x47, 0x0a, 0x0b, 0x53, 0x74, 0x61, 0x74, 0x65, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x12, 0x0c, 0x2e, 0x72, 0x70, 0x63, 0x2e, 0x49, 0x67, 0x6e, 0x6f, 0x72, 0x65, 0x64, 0x1a,
	0x22, 0x2e, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x49, 0x6e, 0x67, 0x72, 0x65, 0x73, 0x73, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x06, 0xb2, 0x89, 0x01, 0x02, 0x10, 0x01, 0x42, 0x2c, 0x5a, 0x2a, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69,
	0x74, 0x2f, 0x6c, 0x69, 0x76, 0x65, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_pkg_service_rpc_ingress_proto_rawDescOnce sync.Once
	file_pkg_service_rpc_ingress_proto_rawDescData = file_pkg_service_rpc_ingress_proto_rawDesc
)

func file_pkg_service_rpc_ingress_proto_rawDescGZIP() []byte {
	file_pkg_service_rpc_ingress_proto_rawDescOnce.Do(func() {
		file_pkg_service_rpc_ingress_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_service_rpc_ingress_proto_rawDescData)
	})
	return file_pkg_service_rpc_ingress_proto_rawDescData
}

var file_pkg_service_rpc_ingress_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_service_rpc_ingress_proto_goTypes = []interface{}{
	(*Ignored)(nil),                           // 0: rpc.Ignored
	(*ListActiveIngressRequest)(nil),          // 1: rpc.ListActiveIngressRequest
	(*ListActiveIngressResponse)(nil),         // 2: rpc.ListActiveIngressResponse
	(*HangUpIngressRequest)(nil),              // 3: rpc.HangUpIngressRequest
	(*HangUpIngressResponse)(nil),             // 4: rpc.HangUpIngressResponse
	(*livekit.GetIngressInfoRequest)(nil),     // 5: livekit.GetIngressInfoRequest
	(*livekit.GetIngressInfoResponse)(nil),    // 6: livekit.GetIngressInfoResponse
	(*livekit.UpdateIngressStateRequest)(nil), // 7: livekit.UpdateIngressStateRequest
}
var file_pkg_service_rpc_ingress_proto_depIdxs = []int32{
	1, // 0: rpc.IngressInternal.ListActiveIngress:input_type -> rpc.ListActiveIngressRequest
	5, // 1: rpc.IngressEntity.GetIngressInfo:input_type -> livekit.GetIngressInfoRequest
	3, // 2: rpc.IngressHandler.HangUpIngress:input_type -> rpc.HangUpIngressRequest
	0, // 3: rpc.IngressUpdate.StateUpdate:input_type -> rpc.Ignored
	2, // 4: rpc.IngressInternal.ListActiveIngress:output_type -> rpc.ListActiveIngressResponse
	6, // 5: rpc.IngressEntity.GetIngressInfo:output_type -> livekit.GetIngressInfoResponse
	4, // 6: rpc.IngressHandler.HangUpIngress:output_type -> rpc.HangUpIngressResponse
	7, // 7: rpc.IngressUpdate.StateUpdate:output_type -> livekit.UpdateIngressStateRequest
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_service_rpc_ingress_proto_init() }
func file_pkg_service_rpc_ingress_proto_init() {
	if File_pkg_service_rpc_ingress_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_service_rpc_ingress_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Ignored); i {
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
		file_pkg_service_rpc_ingress_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListActiveIngressRequest); i {
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
		file_pkg_service_rpc_ingress_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListActiveIngressResponse); i {
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
		file_pkg_service_rpc_ingress_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HangUpIngressRequest); i {
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
		file_pkg_service_rpc_ingress_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*HangUpIngressResponse); i {
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
			RawDescriptor: file_pkg_service_rpc_ingress_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   4,
		},
		GoTypes:           file_pkg_service_rpc_ingress_proto_goTypes,
		DependencyIndexes: file_pkg_service_rpc_ingress_proto_depIdxs,
		MessageInfos:      file_pkg_service_rpc_ingress_proto_msgTypes,
	}.Build()
	File_pkg_service_rpc_ingress_proto = out.File
	file_pkg_service_rpc_ingress_proto_rawDesc = nil
	file_pkg_service_rpc_ingress_proto_goTypes = nil
	file_pkg_service_rpc_ingress_proto_depIdxs = nil
}
