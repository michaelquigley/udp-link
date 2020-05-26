// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fb

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type WireMessage struct {
	_tab flatbuffers.Table
}

func GetRootAsWireMessage(buf []byte, offset flatbuffers.UOffsetT) *WireMessage {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &WireMessage{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *WireMessage) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *WireMessage) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *WireMessage) Sequence() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WireMessage) MutateSequence(n int32) bool {
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *WireMessage) Type() Type {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return Type(rcv._tab.GetInt8(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *WireMessage) MutateType(n Type) bool {
	return rcv._tab.MutateInt8Slot(6, int8(n))
}

func (rcv *WireMessage) Ack() int32 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *WireMessage) MutateAck(n int32) bool {
	return rcv._tab.MutateInt32Slot(8, n)
}

func WireMessageStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func WireMessageAddSequence(builder *flatbuffers.Builder, sequence int32) {
	builder.PrependInt32Slot(0, sequence, 0)
}
func WireMessageAddType(builder *flatbuffers.Builder, type_ Type) {
	builder.PrependInt8Slot(1, int8(type_), 0)
}
func WireMessageAddAck(builder *flatbuffers.Builder, ack int32) {
	builder.PrependInt32Slot(2, ack, 0)
}
func WireMessageEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
