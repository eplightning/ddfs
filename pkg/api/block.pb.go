// Code generated by protoc-gen-go. DO NOT EDIT.
// source: block.proto

package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type BlockGetRequest struct {
	Hashes               [][]byte `protobuf:"bytes,1,rep,name=hashes,proto3" json:"hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockGetRequest) Reset()         { *m = BlockGetRequest{} }
func (m *BlockGetRequest) String() string { return proto.CompactTextString(m) }
func (*BlockGetRequest) ProtoMessage()    {}
func (*BlockGetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{0}
}
func (m *BlockGetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockGetRequest.Unmarshal(m, b)
}
func (m *BlockGetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockGetRequest.Marshal(b, m, deterministic)
}
func (dst *BlockGetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockGetRequest.Merge(dst, src)
}
func (m *BlockGetRequest) XXX_Size() int {
	return xxx_messageInfo_BlockGetRequest.Size(m)
}
func (m *BlockGetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockGetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockGetRequest proto.InternalMessageInfo

func (m *BlockGetRequest) GetHashes() [][]byte {
	if m != nil {
		return m.Hashes
	}
	return nil
}

type BlockGetResponse struct {
	// Types that are valid to be assigned to Msg:
	//	*BlockGetResponse_Info_
	//	*BlockGetResponse_Data_
	Msg                  isBlockGetResponse_Msg `protobuf_oneof:"msg"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *BlockGetResponse) Reset()         { *m = BlockGetResponse{} }
func (m *BlockGetResponse) String() string { return proto.CompactTextString(m) }
func (*BlockGetResponse) ProtoMessage()    {}
func (*BlockGetResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{1}
}
func (m *BlockGetResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockGetResponse.Unmarshal(m, b)
}
func (m *BlockGetResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockGetResponse.Marshal(b, m, deterministic)
}
func (dst *BlockGetResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockGetResponse.Merge(dst, src)
}
func (m *BlockGetResponse) XXX_Size() int {
	return xxx_messageInfo_BlockGetResponse.Size(m)
}
func (m *BlockGetResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockGetResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockGetResponse proto.InternalMessageInfo

type isBlockGetResponse_Msg interface {
	isBlockGetResponse_Msg()
}

type BlockGetResponse_Info_ struct {
	Info *BlockGetResponse_Info `protobuf:"bytes,1,opt,name=info,proto3,oneof"`
}

type BlockGetResponse_Data_ struct {
	Data *BlockGetResponse_Data `protobuf:"bytes,2,opt,name=data,proto3,oneof"`
}

func (*BlockGetResponse_Info_) isBlockGetResponse_Msg() {}

func (*BlockGetResponse_Data_) isBlockGetResponse_Msg() {}

func (m *BlockGetResponse) GetMsg() isBlockGetResponse_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *BlockGetResponse) GetInfo() *BlockGetResponse_Info {
	if x, ok := m.GetMsg().(*BlockGetResponse_Info_); ok {
		return x.Info
	}
	return nil
}

func (m *BlockGetResponse) GetData() *BlockGetResponse_Data {
	if x, ok := m.GetMsg().(*BlockGetResponse_Data_); ok {
		return x.Data
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*BlockGetResponse) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _BlockGetResponse_OneofMarshaler, _BlockGetResponse_OneofUnmarshaler, _BlockGetResponse_OneofSizer, []interface{}{
		(*BlockGetResponse_Info_)(nil),
		(*BlockGetResponse_Data_)(nil),
	}
}

func _BlockGetResponse_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*BlockGetResponse)
	// msg
	switch x := m.Msg.(type) {
	case *BlockGetResponse_Info_:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Info); err != nil {
			return err
		}
	case *BlockGetResponse_Data_:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Data); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("BlockGetResponse.Msg has unexpected type %T", x)
	}
	return nil
}

func _BlockGetResponse_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*BlockGetResponse)
	switch tag {
	case 1: // msg.info
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(BlockGetResponse_Info)
		err := b.DecodeMessage(msg)
		m.Msg = &BlockGetResponse_Info_{msg}
		return true, err
	case 2: // msg.data
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(BlockGetResponse_Data)
		err := b.DecodeMessage(msg)
		m.Msg = &BlockGetResponse_Data_{msg}
		return true, err
	default:
		return false, nil
	}
}

func _BlockGetResponse_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*BlockGetResponse)
	// msg
	switch x := m.Msg.(type) {
	case *BlockGetResponse_Info_:
		s := proto.Size(x.Info)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *BlockGetResponse_Data_:
		s := proto.Size(x.Data)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type BlockGetResponse_Info struct {
	Header               *BlockResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BlockGetResponse_Info) Reset()         { *m = BlockGetResponse_Info{} }
func (m *BlockGetResponse_Info) String() string { return proto.CompactTextString(m) }
func (*BlockGetResponse_Info) ProtoMessage()    {}
func (*BlockGetResponse_Info) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{1, 0}
}
func (m *BlockGetResponse_Info) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockGetResponse_Info.Unmarshal(m, b)
}
func (m *BlockGetResponse_Info) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockGetResponse_Info.Marshal(b, m, deterministic)
}
func (dst *BlockGetResponse_Info) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockGetResponse_Info.Merge(dst, src)
}
func (m *BlockGetResponse_Info) XXX_Size() int {
	return xxx_messageInfo_BlockGetResponse_Info.Size(m)
}
func (m *BlockGetResponse_Info) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockGetResponse_Info.DiscardUnknown(m)
}

var xxx_messageInfo_BlockGetResponse_Info proto.InternalMessageInfo

func (m *BlockGetResponse_Info) GetHeader() *BlockResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

type BlockGetResponse_Data struct {
	Blocks               [][]byte `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockGetResponse_Data) Reset()         { *m = BlockGetResponse_Data{} }
func (m *BlockGetResponse_Data) String() string { return proto.CompactTextString(m) }
func (*BlockGetResponse_Data) ProtoMessage()    {}
func (*BlockGetResponse_Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{1, 1}
}
func (m *BlockGetResponse_Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockGetResponse_Data.Unmarshal(m, b)
}
func (m *BlockGetResponse_Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockGetResponse_Data.Marshal(b, m, deterministic)
}
func (dst *BlockGetResponse_Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockGetResponse_Data.Merge(dst, src)
}
func (m *BlockGetResponse_Data) XXX_Size() int {
	return xxx_messageInfo_BlockGetResponse_Data.Size(m)
}
func (m *BlockGetResponse_Data) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockGetResponse_Data.DiscardUnknown(m)
}

var xxx_messageInfo_BlockGetResponse_Data proto.InternalMessageInfo

func (m *BlockGetResponse_Data) GetBlocks() [][]byte {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type BlockDeleteRequest struct {
	Hashes               [][]byte `protobuf:"bytes,1,rep,name=hashes,proto3" json:"hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockDeleteRequest) Reset()         { *m = BlockDeleteRequest{} }
func (m *BlockDeleteRequest) String() string { return proto.CompactTextString(m) }
func (*BlockDeleteRequest) ProtoMessage()    {}
func (*BlockDeleteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{2}
}
func (m *BlockDeleteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockDeleteRequest.Unmarshal(m, b)
}
func (m *BlockDeleteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockDeleteRequest.Marshal(b, m, deterministic)
}
func (dst *BlockDeleteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockDeleteRequest.Merge(dst, src)
}
func (m *BlockDeleteRequest) XXX_Size() int {
	return xxx_messageInfo_BlockDeleteRequest.Size(m)
}
func (m *BlockDeleteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockDeleteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockDeleteRequest proto.InternalMessageInfo

func (m *BlockDeleteRequest) GetHashes() [][]byte {
	if m != nil {
		return m.Hashes
	}
	return nil
}

type BlockDeleteResponse struct {
	Header               *BlockResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BlockDeleteResponse) Reset()         { *m = BlockDeleteResponse{} }
func (m *BlockDeleteResponse) String() string { return proto.CompactTextString(m) }
func (*BlockDeleteResponse) ProtoMessage()    {}
func (*BlockDeleteResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{3}
}
func (m *BlockDeleteResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockDeleteResponse.Unmarshal(m, b)
}
func (m *BlockDeleteResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockDeleteResponse.Marshal(b, m, deterministic)
}
func (dst *BlockDeleteResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockDeleteResponse.Merge(dst, src)
}
func (m *BlockDeleteResponse) XXX_Size() int {
	return xxx_messageInfo_BlockDeleteResponse.Size(m)
}
func (m *BlockDeleteResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockDeleteResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockDeleteResponse proto.InternalMessageInfo

func (m *BlockDeleteResponse) GetHeader() *BlockResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

type BlockReserveRequest struct {
	Hashes               [][]byte `protobuf:"bytes,1,rep,name=hashes,proto3" json:"hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockReserveRequest) Reset()         { *m = BlockReserveRequest{} }
func (m *BlockReserveRequest) String() string { return proto.CompactTextString(m) }
func (*BlockReserveRequest) ProtoMessage()    {}
func (*BlockReserveRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{4}
}
func (m *BlockReserveRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockReserveRequest.Unmarshal(m, b)
}
func (m *BlockReserveRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockReserveRequest.Marshal(b, m, deterministic)
}
func (dst *BlockReserveRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockReserveRequest.Merge(dst, src)
}
func (m *BlockReserveRequest) XXX_Size() int {
	return xxx_messageInfo_BlockReserveRequest.Size(m)
}
func (m *BlockReserveRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockReserveRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockReserveRequest proto.InternalMessageInfo

func (m *BlockReserveRequest) GetHashes() [][]byte {
	if m != nil {
		return m.Hashes
	}
	return nil
}

type BlockReserveResponse struct {
	Header               *BlockResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	ReservationId        string               `protobuf:"bytes,2,opt,name=reservation_id,json=reservationId,proto3" json:"reservation_id,omitempty"`
	MissingBlocks        []int32              `protobuf:"varint,3,rep,packed,name=missing_blocks,json=missingBlocks,proto3" json:"missing_blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BlockReserveResponse) Reset()         { *m = BlockReserveResponse{} }
func (m *BlockReserveResponse) String() string { return proto.CompactTextString(m) }
func (*BlockReserveResponse) ProtoMessage()    {}
func (*BlockReserveResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{5}
}
func (m *BlockReserveResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockReserveResponse.Unmarshal(m, b)
}
func (m *BlockReserveResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockReserveResponse.Marshal(b, m, deterministic)
}
func (dst *BlockReserveResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockReserveResponse.Merge(dst, src)
}
func (m *BlockReserveResponse) XXX_Size() int {
	return xxx_messageInfo_BlockReserveResponse.Size(m)
}
func (m *BlockReserveResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockReserveResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockReserveResponse proto.InternalMessageInfo

func (m *BlockReserveResponse) GetHeader() *BlockResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *BlockReserveResponse) GetReservationId() string {
	if m != nil {
		return m.ReservationId
	}
	return ""
}

func (m *BlockReserveResponse) GetMissingBlocks() []int32 {
	if m != nil {
		return m.MissingBlocks
	}
	return nil
}

type BlockPutRequest struct {
	// Types that are valid to be assigned to Msg:
	//	*BlockPutRequest_Info_
	//	*BlockPutRequest_Data_
	Msg                  isBlockPutRequest_Msg `protobuf_oneof:"msg"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *BlockPutRequest) Reset()         { *m = BlockPutRequest{} }
func (m *BlockPutRequest) String() string { return proto.CompactTextString(m) }
func (*BlockPutRequest) ProtoMessage()    {}
func (*BlockPutRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{6}
}
func (m *BlockPutRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockPutRequest.Unmarshal(m, b)
}
func (m *BlockPutRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockPutRequest.Marshal(b, m, deterministic)
}
func (dst *BlockPutRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockPutRequest.Merge(dst, src)
}
func (m *BlockPutRequest) XXX_Size() int {
	return xxx_messageInfo_BlockPutRequest.Size(m)
}
func (m *BlockPutRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockPutRequest.DiscardUnknown(m)
}

var xxx_messageInfo_BlockPutRequest proto.InternalMessageInfo

type isBlockPutRequest_Msg interface {
	isBlockPutRequest_Msg()
}

type BlockPutRequest_Info_ struct {
	Info *BlockPutRequest_Info `protobuf:"bytes,1,opt,name=info,proto3,oneof"`
}

type BlockPutRequest_Data_ struct {
	Data *BlockPutRequest_Data `protobuf:"bytes,2,opt,name=data,proto3,oneof"`
}

func (*BlockPutRequest_Info_) isBlockPutRequest_Msg() {}

func (*BlockPutRequest_Data_) isBlockPutRequest_Msg() {}

func (m *BlockPutRequest) GetMsg() isBlockPutRequest_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (m *BlockPutRequest) GetInfo() *BlockPutRequest_Info {
	if x, ok := m.GetMsg().(*BlockPutRequest_Info_); ok {
		return x.Info
	}
	return nil
}

func (m *BlockPutRequest) GetData() *BlockPutRequest_Data {
	if x, ok := m.GetMsg().(*BlockPutRequest_Data_); ok {
		return x.Data
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*BlockPutRequest) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _BlockPutRequest_OneofMarshaler, _BlockPutRequest_OneofUnmarshaler, _BlockPutRequest_OneofSizer, []interface{}{
		(*BlockPutRequest_Info_)(nil),
		(*BlockPutRequest_Data_)(nil),
	}
}

func _BlockPutRequest_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*BlockPutRequest)
	// msg
	switch x := m.Msg.(type) {
	case *BlockPutRequest_Info_:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Info); err != nil {
			return err
		}
	case *BlockPutRequest_Data_:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Data); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("BlockPutRequest.Msg has unexpected type %T", x)
	}
	return nil
}

func _BlockPutRequest_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*BlockPutRequest)
	switch tag {
	case 1: // msg.info
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(BlockPutRequest_Info)
		err := b.DecodeMessage(msg)
		m.Msg = &BlockPutRequest_Info_{msg}
		return true, err
	case 2: // msg.data
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(BlockPutRequest_Data)
		err := b.DecodeMessage(msg)
		m.Msg = &BlockPutRequest_Data_{msg}
		return true, err
	default:
		return false, nil
	}
}

func _BlockPutRequest_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*BlockPutRequest)
	// msg
	switch x := m.Msg.(type) {
	case *BlockPutRequest_Info_:
		s := proto.Size(x.Info)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case *BlockPutRequest_Data_:
		s := proto.Size(x.Data)
		n += 1 // tag and wire
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type BlockPutRequest_Info struct {
	ReservationId        string   `protobuf:"bytes,1,opt,name=reservation_id,json=reservationId,proto3" json:"reservation_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockPutRequest_Info) Reset()         { *m = BlockPutRequest_Info{} }
func (m *BlockPutRequest_Info) String() string { return proto.CompactTextString(m) }
func (*BlockPutRequest_Info) ProtoMessage()    {}
func (*BlockPutRequest_Info) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{6, 0}
}
func (m *BlockPutRequest_Info) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockPutRequest_Info.Unmarshal(m, b)
}
func (m *BlockPutRequest_Info) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockPutRequest_Info.Marshal(b, m, deterministic)
}
func (dst *BlockPutRequest_Info) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockPutRequest_Info.Merge(dst, src)
}
func (m *BlockPutRequest_Info) XXX_Size() int {
	return xxx_messageInfo_BlockPutRequest_Info.Size(m)
}
func (m *BlockPutRequest_Info) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockPutRequest_Info.DiscardUnknown(m)
}

var xxx_messageInfo_BlockPutRequest_Info proto.InternalMessageInfo

func (m *BlockPutRequest_Info) GetReservationId() string {
	if m != nil {
		return m.ReservationId
	}
	return ""
}

type BlockPutRequest_Data struct {
	Blocks               []*HashBlockPair `protobuf:"bytes,1,rep,name=blocks,proto3" json:"blocks,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *BlockPutRequest_Data) Reset()         { *m = BlockPutRequest_Data{} }
func (m *BlockPutRequest_Data) String() string { return proto.CompactTextString(m) }
func (*BlockPutRequest_Data) ProtoMessage()    {}
func (*BlockPutRequest_Data) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{6, 1}
}
func (m *BlockPutRequest_Data) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockPutRequest_Data.Unmarshal(m, b)
}
func (m *BlockPutRequest_Data) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockPutRequest_Data.Marshal(b, m, deterministic)
}
func (dst *BlockPutRequest_Data) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockPutRequest_Data.Merge(dst, src)
}
func (m *BlockPutRequest_Data) XXX_Size() int {
	return xxx_messageInfo_BlockPutRequest_Data.Size(m)
}
func (m *BlockPutRequest_Data) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockPutRequest_Data.DiscardUnknown(m)
}

var xxx_messageInfo_BlockPutRequest_Data proto.InternalMessageInfo

func (m *BlockPutRequest_Data) GetBlocks() []*HashBlockPair {
	if m != nil {
		return m.Blocks
	}
	return nil
}

type BlockPutResponse struct {
	Header               *BlockResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *BlockPutResponse) Reset()         { *m = BlockPutResponse{} }
func (m *BlockPutResponse) String() string { return proto.CompactTextString(m) }
func (*BlockPutResponse) ProtoMessage()    {}
func (*BlockPutResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{7}
}
func (m *BlockPutResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockPutResponse.Unmarshal(m, b)
}
func (m *BlockPutResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockPutResponse.Marshal(b, m, deterministic)
}
func (dst *BlockPutResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockPutResponse.Merge(dst, src)
}
func (m *BlockPutResponse) XXX_Size() int {
	return xxx_messageInfo_BlockPutResponse.Size(m)
}
func (m *BlockPutResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockPutResponse.DiscardUnknown(m)
}

var xxx_messageInfo_BlockPutResponse proto.InternalMessageInfo

func (m *BlockPutResponse) GetHeader() *BlockResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

type HashBlockPair struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Block                []byte   `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashBlockPair) Reset()         { *m = HashBlockPair{} }
func (m *HashBlockPair) String() string { return proto.CompactTextString(m) }
func (*HashBlockPair) ProtoMessage()    {}
func (*HashBlockPair) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{8}
}
func (m *HashBlockPair) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashBlockPair.Unmarshal(m, b)
}
func (m *HashBlockPair) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashBlockPair.Marshal(b, m, deterministic)
}
func (dst *HashBlockPair) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashBlockPair.Merge(dst, src)
}
func (m *HashBlockPair) XXX_Size() int {
	return xxx_messageInfo_HashBlockPair.Size(m)
}
func (m *HashBlockPair) XXX_DiscardUnknown() {
	xxx_messageInfo_HashBlockPair.DiscardUnknown(m)
}

var xxx_messageInfo_HashBlockPair proto.InternalMessageInfo

func (m *HashBlockPair) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *HashBlockPair) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

type BlockResponseHeader struct {
	Error                int32    `protobuf:"varint,1,opt,name=error,proto3" json:"error,omitempty"`
	ErrorMsg             string   `protobuf:"bytes,2,opt,name=error_msg,json=errorMsg,proto3" json:"error_msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *BlockResponseHeader) Reset()         { *m = BlockResponseHeader{} }
func (m *BlockResponseHeader) String() string { return proto.CompactTextString(m) }
func (*BlockResponseHeader) ProtoMessage()    {}
func (*BlockResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_block_573b194e6d346b5c, []int{9}
}
func (m *BlockResponseHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_BlockResponseHeader.Unmarshal(m, b)
}
func (m *BlockResponseHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_BlockResponseHeader.Marshal(b, m, deterministic)
}
func (dst *BlockResponseHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BlockResponseHeader.Merge(dst, src)
}
func (m *BlockResponseHeader) XXX_Size() int {
	return xxx_messageInfo_BlockResponseHeader.Size(m)
}
func (m *BlockResponseHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_BlockResponseHeader.DiscardUnknown(m)
}

var xxx_messageInfo_BlockResponseHeader proto.InternalMessageInfo

func (m *BlockResponseHeader) GetError() int32 {
	if m != nil {
		return m.Error
	}
	return 0
}

func (m *BlockResponseHeader) GetErrorMsg() string {
	if m != nil {
		return m.ErrorMsg
	}
	return ""
}

func init() {
	proto.RegisterType((*BlockGetRequest)(nil), "api.BlockGetRequest")
	proto.RegisterType((*BlockGetResponse)(nil), "api.BlockGetResponse")
	proto.RegisterType((*BlockGetResponse_Info)(nil), "api.BlockGetResponse.Info")
	proto.RegisterType((*BlockGetResponse_Data)(nil), "api.BlockGetResponse.Data")
	proto.RegisterType((*BlockDeleteRequest)(nil), "api.BlockDeleteRequest")
	proto.RegisterType((*BlockDeleteResponse)(nil), "api.BlockDeleteResponse")
	proto.RegisterType((*BlockReserveRequest)(nil), "api.BlockReserveRequest")
	proto.RegisterType((*BlockReserveResponse)(nil), "api.BlockReserveResponse")
	proto.RegisterType((*BlockPutRequest)(nil), "api.BlockPutRequest")
	proto.RegisterType((*BlockPutRequest_Info)(nil), "api.BlockPutRequest.Info")
	proto.RegisterType((*BlockPutRequest_Data)(nil), "api.BlockPutRequest.Data")
	proto.RegisterType((*BlockPutResponse)(nil), "api.BlockPutResponse")
	proto.RegisterType((*HashBlockPair)(nil), "api.HashBlockPair")
	proto.RegisterType((*BlockResponseHeader)(nil), "api.BlockResponseHeader")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// BlockStoreClient is the client API for BlockStore service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type BlockStoreClient interface {
	Get(ctx context.Context, in *BlockGetRequest, opts ...grpc.CallOption) (BlockStore_GetClient, error)
	Put(ctx context.Context, opts ...grpc.CallOption) (BlockStore_PutClient, error)
	Reserve(ctx context.Context, in *BlockReserveRequest, opts ...grpc.CallOption) (*BlockReserveResponse, error)
	Delete(ctx context.Context, in *BlockDeleteRequest, opts ...grpc.CallOption) (*BlockDeleteResponse, error)
}

type blockStoreClient struct {
	cc *grpc.ClientConn
}

func NewBlockStoreClient(cc *grpc.ClientConn) BlockStoreClient {
	return &blockStoreClient{cc}
}

func (c *blockStoreClient) Get(ctx context.Context, in *BlockGetRequest, opts ...grpc.CallOption) (BlockStore_GetClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockStore_serviceDesc.Streams[0], "/api.BlockStore/Get", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockStoreGetClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BlockStore_GetClient interface {
	Recv() (*BlockGetResponse, error)
	grpc.ClientStream
}

type blockStoreGetClient struct {
	grpc.ClientStream
}

func (x *blockStoreGetClient) Recv() (*BlockGetResponse, error) {
	m := new(BlockGetResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blockStoreClient) Put(ctx context.Context, opts ...grpc.CallOption) (BlockStore_PutClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BlockStore_serviceDesc.Streams[1], "/api.BlockStore/Put", opts...)
	if err != nil {
		return nil, err
	}
	x := &blockStorePutClient{stream}
	return x, nil
}

type BlockStore_PutClient interface {
	Send(*BlockPutRequest) error
	CloseAndRecv() (*BlockPutResponse, error)
	grpc.ClientStream
}

type blockStorePutClient struct {
	grpc.ClientStream
}

func (x *blockStorePutClient) Send(m *BlockPutRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *blockStorePutClient) CloseAndRecv() (*BlockPutResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(BlockPutResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *blockStoreClient) Reserve(ctx context.Context, in *BlockReserveRequest, opts ...grpc.CallOption) (*BlockReserveResponse, error) {
	out := new(BlockReserveResponse)
	err := c.cc.Invoke(ctx, "/api.BlockStore/Reserve", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *blockStoreClient) Delete(ctx context.Context, in *BlockDeleteRequest, opts ...grpc.CallOption) (*BlockDeleteResponse, error) {
	out := new(BlockDeleteResponse)
	err := c.cc.Invoke(ctx, "/api.BlockStore/Delete", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BlockStoreServer is the server API for BlockStore service.
type BlockStoreServer interface {
	Get(*BlockGetRequest, BlockStore_GetServer) error
	Put(BlockStore_PutServer) error
	Reserve(context.Context, *BlockReserveRequest) (*BlockReserveResponse, error)
	Delete(context.Context, *BlockDeleteRequest) (*BlockDeleteResponse, error)
}

func RegisterBlockStoreServer(s *grpc.Server, srv BlockStoreServer) {
	s.RegisterService(&_BlockStore_serviceDesc, srv)
}

func _BlockStore_Get_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(BlockGetRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BlockStoreServer).Get(m, &blockStoreGetServer{stream})
}

type BlockStore_GetServer interface {
	Send(*BlockGetResponse) error
	grpc.ServerStream
}

type blockStoreGetServer struct {
	grpc.ServerStream
}

func (x *blockStoreGetServer) Send(m *BlockGetResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _BlockStore_Put_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BlockStoreServer).Put(&blockStorePutServer{stream})
}

type BlockStore_PutServer interface {
	SendAndClose(*BlockPutResponse) error
	Recv() (*BlockPutRequest, error)
	grpc.ServerStream
}

type blockStorePutServer struct {
	grpc.ServerStream
}

func (x *blockStorePutServer) SendAndClose(m *BlockPutResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *blockStorePutServer) Recv() (*BlockPutRequest, error) {
	m := new(BlockPutRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _BlockStore_Reserve_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockReserveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockStoreServer).Reserve(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.BlockStore/Reserve",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockStoreServer).Reserve(ctx, req.(*BlockReserveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BlockStore_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BlockDeleteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlockStoreServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.BlockStore/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlockStoreServer).Delete(ctx, req.(*BlockDeleteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _BlockStore_serviceDesc = grpc.ServiceDesc{
	ServiceName: "api.BlockStore",
	HandlerType: (*BlockStoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Reserve",
			Handler:    _BlockStore_Reserve_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _BlockStore_Delete_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Get",
			Handler:       _BlockStore_Get_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Put",
			Handler:       _BlockStore_Put_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "block.proto",
}

func init() { proto.RegisterFile("block.proto", fileDescriptor_block_573b194e6d346b5c) }

var fileDescriptor_block_573b194e6d346b5c = []byte{
	// 492 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xdd, 0x8a, 0x13, 0x31,
	0x14, 0x6e, 0x9c, 0xb6, 0xba, 0xa7, 0xad, 0xca, 0xb1, 0xea, 0xec, 0x08, 0x52, 0x06, 0x84, 0x2a,
	0x6e, 0x77, 0xa9, 0x20, 0x7a, 0x21, 0xc8, 0x52, 0x68, 0x7b, 0x21, 0x2c, 0xf1, 0x01, 0x4a, 0xd6,
	0x66, 0xdb, 0xc1, 0xed, 0xa4, 0x4e, 0x52, 0x1f, 0xc5, 0x57, 0xf3, 0xca, 0x47, 0x11, 0x24, 0x67,
	0xd2, 0xce, 0xa4, 0x3b, 0xb8, 0x6c, 0xef, 0x26, 0xc9, 0xf7, 0x93, 0x73, 0xce, 0x97, 0x81, 0xd6,
	0xe5, 0xb5, 0xfa, 0xf6, 0x7d, 0xb0, 0xce, 0x94, 0x51, 0x18, 0x88, 0x75, 0x12, 0xbf, 0x86, 0x47,
	0xe7, 0x76, 0x6f, 0x2c, 0x0d, 0x97, 0x3f, 0x36, 0x52, 0x1b, 0x7c, 0x06, 0xcd, 0xa5, 0xd0, 0x4b,
	0xa9, 0x43, 0xd6, 0x0b, 0xfa, 0x6d, 0xee, 0x56, 0xf1, 0x6f, 0x06, 0x8f, 0x0b, 0xac, 0x5e, 0xab,
	0x54, 0x4b, 0x3c, 0x83, 0x7a, 0x92, 0x5e, 0xa9, 0x90, 0xf5, 0x58, 0xbf, 0x35, 0x8c, 0x06, 0x62,
	0x9d, 0x0c, 0xf6, 0x41, 0x83, 0x69, 0x7a, 0xa5, 0x26, 0x35, 0x4e, 0x48, 0xcb, 0x98, 0x0b, 0x23,
	0xc2, 0x7b, 0xff, 0x63, 0x8c, 0x84, 0x11, 0x96, 0x61, 0x91, 0xd1, 0x07, 0xa8, 0x4f, 0x73, 0x66,
	0x73, 0x29, 0xc5, 0x5c, 0x66, 0xce, 0x2d, 0x2c, 0xb8, 0x5b, 0xe2, 0x84, 0xce, 0xb9, 0xc3, 0x45,
	0x2f, 0xa1, 0x6e, 0x95, 0x6c, 0x49, 0x54, 0xf9, 0xae, 0xa4, 0x7c, 0x75, 0xde, 0x80, 0x60, 0xa5,
	0x17, 0xf1, 0x5b, 0x40, 0x52, 0x19, 0xc9, 0x6b, 0x69, 0xe4, 0x6d, 0x7d, 0x18, 0xc3, 0x13, 0x0f,
	0xbd, 0xeb, 0xc4, 0x1d, 0x6f, 0x17, 0x9f, 0x38, 0x21, 0x2e, 0xb5, 0xcc, 0x7e, 0xde, 0xea, 0xfb,
	0x8b, 0x41, 0xd7, 0xc7, 0x1f, 0xea, 0x8c, 0xaf, 0xe0, 0x61, 0x46, 0x22, 0xc2, 0x24, 0x2a, 0x9d,
	0x25, 0x73, 0x9a, 0xc6, 0x11, 0xef, 0x94, 0x76, 0xa7, 0x73, 0x0b, 0x5b, 0x25, 0x5a, 0x27, 0xe9,
	0x62, 0xe6, 0xda, 0x17, 0xf4, 0x82, 0x7e, 0x83, 0x77, 0xdc, 0x2e, 0x59, 0xe8, 0xf8, 0x0f, 0x73,
	0x21, 0xba, 0xd8, 0xec, 0x42, 0x74, 0xea, 0xe5, 0xe2, 0xb8, 0xb8, 0x51, 0x81, 0xf1, 0x63, 0x71,
	0xea, 0xc5, 0xa2, 0x9a, 0xe0, 0xa5, 0xe2, 0xc4, 0xa5, 0xe2, 0x66, 0x2d, 0xac, 0xa2, 0x96, 0x68,
	0xe8, 0xa2, 0xf0, 0xc6, 0x8b, 0x42, 0x6b, 0x88, 0xe4, 0x34, 0x11, 0x7a, 0x99, 0xbb, 0x89, 0x24,
	0xdb, 0x8f, 0xc7, 0xc8, 0xe5, 0x9e, 0x6e, 0x72, 0xf0, 0xb4, 0x3f, 0x42, 0xc7, 0x73, 0x41, 0x84,
	0xba, 0x9d, 0x2c, 0x09, 0xb4, 0x39, 0x7d, 0x63, 0x17, 0x1a, 0xe4, 0x4d, 0x6d, 0x68, 0xf3, 0x7c,
	0x11, 0x4f, 0x8a, 0xa0, 0x94, 0x94, 0x2d, 0x58, 0x66, 0x99, 0xca, 0xaf, 0xd0, 0xe0, 0xf9, 0x02,
	0x5f, 0xc0, 0x11, 0x7d, 0xcc, 0x56, 0x7a, 0xe1, 0xc6, 0xfa, 0x80, 0x36, 0xbe, 0xe8, 0xc5, 0xf0,
	0x2f, 0x03, 0x20, 0xa9, 0xaf, 0x46, 0x65, 0x12, 0xdf, 0x43, 0x30, 0x96, 0x06, 0xbb, 0x7b, 0x8f,
	0x90, 0xba, 0x1d, 0x3d, 0xad, 0x7c, 0x9a, 0x71, 0xed, 0x8c, 0x59, 0xde, 0xc5, 0xc6, 0xe3, 0x15,
	0x53, 0x2a, 0xf3, 0x4a, 0x1d, 0x8b, 0x6b, 0x7d, 0x86, 0x9f, 0xe1, 0xbe, 0x0b, 0x2f, 0xfa, 0x0d,
	0x2b, 0xe5, 0x3f, 0x3a, 0xae, 0x38, 0xd9, 0x6a, 0xe0, 0x27, 0x68, 0xe6, 0xef, 0x0e, 0x9f, 0x17,
	0x30, 0xef, 0xdd, 0x46, 0xe1, 0xcd, 0x83, 0x2d, 0xfd, 0xb2, 0x49, 0xbf, 0xbe, 0x77, 0xff, 0x02,
	0x00, 0x00, 0xff, 0xff, 0x2d, 0x27, 0x38, 0xf6, 0x09, 0x05, 0x00, 0x00,
}
