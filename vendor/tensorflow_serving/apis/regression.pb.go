// Code generated by protoc-gen-go. DO NOT EDIT.
// source: tensorflow_serving/apis/regression.proto

package tensorflow_serving

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Regression result for a single item (tensorflow.Example).
type Regression struct {
	Value                float32  `protobuf:"fixed32,1,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Regression) Reset()         { *m = Regression{} }
func (m *Regression) String() string { return proto.CompactTextString(m) }
func (*Regression) ProtoMessage()    {}
func (*Regression) Descriptor() ([]byte, []int) {
	return fileDescriptor_regression_0f3e75a0dfe5eb5a, []int{0}
}
func (m *Regression) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Regression.Unmarshal(m, b)
}
func (m *Regression) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Regression.Marshal(b, m, deterministic)
}
func (dst *Regression) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Regression.Merge(dst, src)
}
func (m *Regression) XXX_Size() int {
	return xxx_messageInfo_Regression.Size(m)
}
func (m *Regression) XXX_DiscardUnknown() {
	xxx_messageInfo_Regression.DiscardUnknown(m)
}

var xxx_messageInfo_Regression proto.InternalMessageInfo

func (m *Regression) GetValue() float32 {
	if m != nil {
		return m.Value
	}
	return 0
}

// Contains one result per input example, in the same order as the input in
// RegressionRequest.
type RegressionResult struct {
	Regressions          []*Regression `protobuf:"bytes,1,rep,name=regressions,proto3" json:"regressions,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *RegressionResult) Reset()         { *m = RegressionResult{} }
func (m *RegressionResult) String() string { return proto.CompactTextString(m) }
func (*RegressionResult) ProtoMessage()    {}
func (*RegressionResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_regression_0f3e75a0dfe5eb5a, []int{1}
}
func (m *RegressionResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegressionResult.Unmarshal(m, b)
}
func (m *RegressionResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegressionResult.Marshal(b, m, deterministic)
}
func (dst *RegressionResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegressionResult.Merge(dst, src)
}
func (m *RegressionResult) XXX_Size() int {
	return xxx_messageInfo_RegressionResult.Size(m)
}
func (m *RegressionResult) XXX_DiscardUnknown() {
	xxx_messageInfo_RegressionResult.DiscardUnknown(m)
}

var xxx_messageInfo_RegressionResult proto.InternalMessageInfo

func (m *RegressionResult) GetRegressions() []*Regression {
	if m != nil {
		return m.Regressions
	}
	return nil
}

type RegressionRequest struct {
	// Model Specification. If version is not specified, will use the latest
	// (numerical) version.
	ModelSpec *ModelSpec `protobuf:"bytes,1,opt,name=model_spec,json=modelSpec,proto3" json:"model_spec,omitempty"`
	// Input data.
	Input                *Input   `protobuf:"bytes,2,opt,name=input,proto3" json:"input,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RegressionRequest) Reset()         { *m = RegressionRequest{} }
func (m *RegressionRequest) String() string { return proto.CompactTextString(m) }
func (*RegressionRequest) ProtoMessage()    {}
func (*RegressionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_regression_0f3e75a0dfe5eb5a, []int{2}
}
func (m *RegressionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegressionRequest.Unmarshal(m, b)
}
func (m *RegressionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegressionRequest.Marshal(b, m, deterministic)
}
func (dst *RegressionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegressionRequest.Merge(dst, src)
}
func (m *RegressionRequest) XXX_Size() int {
	return xxx_messageInfo_RegressionRequest.Size(m)
}
func (m *RegressionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_RegressionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_RegressionRequest proto.InternalMessageInfo

func (m *RegressionRequest) GetModelSpec() *ModelSpec {
	if m != nil {
		return m.ModelSpec
	}
	return nil
}

func (m *RegressionRequest) GetInput() *Input {
	if m != nil {
		return m.Input
	}
	return nil
}

type RegressionResponse struct {
	// Effective Model Specification used for regression.
	ModelSpec            *ModelSpec        `protobuf:"bytes,2,opt,name=model_spec,json=modelSpec,proto3" json:"model_spec,omitempty"`
	Result               *RegressionResult `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *RegressionResponse) Reset()         { *m = RegressionResponse{} }
func (m *RegressionResponse) String() string { return proto.CompactTextString(m) }
func (*RegressionResponse) ProtoMessage()    {}
func (*RegressionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_regression_0f3e75a0dfe5eb5a, []int{3}
}
func (m *RegressionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RegressionResponse.Unmarshal(m, b)
}
func (m *RegressionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RegressionResponse.Marshal(b, m, deterministic)
}
func (dst *RegressionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RegressionResponse.Merge(dst, src)
}
func (m *RegressionResponse) XXX_Size() int {
	return xxx_messageInfo_RegressionResponse.Size(m)
}
func (m *RegressionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_RegressionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_RegressionResponse proto.InternalMessageInfo

func (m *RegressionResponse) GetModelSpec() *ModelSpec {
	if m != nil {
		return m.ModelSpec
	}
	return nil
}

func (m *RegressionResponse) GetResult() *RegressionResult {
	if m != nil {
		return m.Result
	}
	return nil
}

func init() {
	proto.RegisterType((*Regression)(nil), "tensorflow.serving.Regression")
	proto.RegisterType((*RegressionResult)(nil), "tensorflow.serving.RegressionResult")
	proto.RegisterType((*RegressionRequest)(nil), "tensorflow.serving.RegressionRequest")
	proto.RegisterType((*RegressionResponse)(nil), "tensorflow.serving.RegressionResponse")
}

func init() {
	proto.RegisterFile("tensorflow_serving/apis/regression.proto", fileDescriptor_regression_0f3e75a0dfe5eb5a)
}

var fileDescriptor_regression_0f3e75a0dfe5eb5a = []byte{
	// 260 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x40, 0xe5, 0x54, 0xad, 0xc4, 0x65, 0x01, 0x8b, 0x21, 0x20, 0x81, 0x2a, 0xc3, 0x90, 0x29,
	0x91, 0xca, 0xda, 0x01, 0xb1, 0x31, 0xb0, 0x18, 0xf6, 0xaa, 0x94, 0xa3, 0x8a, 0xe4, 0xda, 0xc6,
	0x67, 0x97, 0x9d, 0x2f, 0xe0, 0x73, 0x19, 0x51, 0xed, 0x40, 0x02, 0xb4, 0x42, 0xdd, 0x12, 0xe9,
	0xbd, 0xf3, 0xf3, 0x19, 0x4a, 0x8f, 0x9a, 0x8c, 0x7b, 0x56, 0xe6, 0x75, 0x46, 0xe8, 0xd6, 0x8d,
	0x5e, 0xd6, 0x73, 0xdb, 0x50, 0xed, 0x70, 0xe9, 0x90, 0xa8, 0x31, 0xba, 0xb2, 0xce, 0x78, 0xc3,
	0x79, 0x47, 0x56, 0x2d, 0x79, 0x7a, 0xb1, 0xcb, 0x6e, 0xb4, 0x0d, 0x3e, 0x89, 0xbb, 0xa1, 0x95,
	0x79, 0x42, 0x95, 0x20, 0x21, 0x00, 0xe4, 0xf7, 0x89, 0xfc, 0x18, 0x86, 0xeb, 0xb9, 0x0a, 0x58,
	0xb0, 0x31, 0x2b, 0x33, 0x99, 0x7e, 0xc4, 0x03, 0x1c, 0x76, 0x8c, 0x44, 0x0a, 0xca, 0xf3, 0x6b,
	0xc8, 0xbb, 0x52, 0x2a, 0xd8, 0x78, 0x50, 0xe6, 0x93, 0xf3, 0xea, 0x6f, 0x6b, 0xd5, 0x53, 0xfb,
	0x8a, 0x78, 0x63, 0x70, 0xd4, 0x1f, 0xfb, 0x12, 0x90, 0x3c, 0x9f, 0x02, 0xc4, 0xbc, 0x19, 0x59,
	0x5c, 0xc4, 0x8c, 0x7c, 0x72, 0xb6, 0x6d, 0xec, 0xdd, 0x86, 0xba, 0xb7, 0xb8, 0x90, 0x07, 0xab,
	0xaf, 0x4f, 0x5e, 0xc3, 0x30, 0x6e, 0xa0, 0xc8, 0xa2, 0x78, 0xb2, 0x4d, 0xbc, 0xdd, 0x00, 0x32,
	0x71, 0xe2, 0x9d, 0x01, 0xff, 0x71, 0x37, 0x6b, 0x34, 0xe1, 0xaf, 0x8a, 0x6c, 0xcf, 0x8a, 0x29,
	0x8c, 0x5c, 0xdc, 0x52, 0xdb, 0x7f, 0xf9, 0xcf, 0x5a, 0x22, 0x2b, 0x5b, 0xe7, 0x66, 0xf0, 0xc1,
	0xd8, 0xe3, 0x28, 0xbe, 0xce, 0xd5, 0x67, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd2, 0xd1, 0x6c, 0x2e,
	0x27, 0x02, 0x00, 0x00,
}
