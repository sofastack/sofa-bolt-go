package bolt

import (
	"fmt"
	"io"

	"github.com/sofastack/sofa-bolt-go/sofabolt"
)

var (
	// compiler hints
	_ sofabolt.ClientConnProtocolDecoder = (*BOLTProtocolDecoder)(nil)
	_ sofabolt.ClientConnProtocolEncoder = (*BOLTProtocolEncoder)(nil)
)

type BOLTProtocolEncoder struct {
}

func (be *BOLTProtocolEncoder) Encode(eo *sofabolt.ClientConnProtocolEncoderOption,
	dst []byte, cmd interface{}) ([]byte, error) {
	var err error
	switch x := cmd.(type) {
	case *sofabolt.Request:
		dst, err = x.Write(&sofabolt.WriteOption{}, dst)
		if err != nil {
			return nil, err
		}
	case *sofabolt.Response:
		dst, err = x.Write(&sofabolt.WriteOption{}, dst)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown command type: %+v", x)
	}
	return dst, err
}

type BOLTProtocolDecoder struct {
}

func (bd *BOLTProtocolDecoder) Decode(do *sofabolt.ClientConnProtocolDecoderOption, r io.Reader) (interface{}, error) {
	var cmd sofabolt.Command
	_, err := cmd.Read(&sofabolt.ReadOption{}, r)
	return &cmd, err
}
