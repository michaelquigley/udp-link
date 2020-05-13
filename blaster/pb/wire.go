package pb

import (
	"bytes"
	"encoding/binary"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func NewSync(sequence int32) *WireMessage {
	return &WireMessage{
		Sequence: sequence,
		Type:     MessageType_SYNC,
	}
}

func NewHello(sequence int32, session string) *WireMessage {
	return &WireMessage{
		Sequence: sequence,
		Type:     MessageType_HELLO,
		HelloPayload: &WireMessage_HelloPayload{
			Session: session,
		},
	}
}

func NewOk(sequence int32) *WireMessage {
	return &WireMessage{
		Sequence: sequence,
		Type:     MessageType_OK,
	}
}

func ToData(wireMessage *WireMessage) ([]byte, error) {
	pbData, err := proto.Marshal(wireMessage)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	data := make([]byte, len(pbData)+4)
	buffer := bytes.NewBuffer(data)
	if err := binary.Write(buffer, binary.LittleEndian, len(pbData)); err != nil {
		return nil, errors.Wrap(err, "write length")
	}
	n, err := buffer.Write(pbData)
	if err != nil {
		return nil, errors.Wrap(err, "write pb data")
	}
	if n != len(pbData) {
		return nil, errors.New("short pb data")
	}

	return data, nil
}
