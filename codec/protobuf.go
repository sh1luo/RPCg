package codec

import (
	"encoding/json"
	"fmt"
	"github.com/gogo/protobuf/proto"
	pb "github.com/golang/protobuf/proto"
	"io"
	reply "rpcg/codec/pb"
)

type PBCodec struct {
	conn io.ReadWriteCloser
	buf  *pb.Buffer
}

var _ Codec = (*PBCodec)(nil)

func NewPBCodec(conn io.ReadWriteCloser) Codec {
	return &PBCodec{
		conn: conn,
		buf:  &pb.Buffer{},
	}
}

func (p *PBCodec) Close() error {
	return p.conn.Close()
}

func (p *PBCodec) ReadHeader(header *Header) (err error) {
	r := &reply.ProtoReply{}
	err = p.buf.DecodeMessage(r)

	header.ServiceMethod = r.ServiceMethod
	header.Seq = r.Seq
	header.Error = r.Error
	return
}

func (p *PBCodec) ReadBody(i interface{}) error {
	panic("implement me")
}

func (p *PBCodec) Write(header *Header, i interface{}) (err error) {
	defer func() {
		p.buf.Reset()
		if err != nil {
			_ = p.conn.Close()
		}
	}()

	hb, err := json.Marshal(header)
	bb, err := json.Marshal(i)
	if err != nil {
		_ = fmt.Errorf("protobuf write err:%s", err)
		return err
	}
	hb = append(hb, bb...)
	return p.buf.EncodeRawBytes(hb)
}

func Encode(data interface{}) ([]byte, error) {
	if m, ok := data.(proto.Marshaler); ok {
		return m.Marshal()
	}

	if m, ok := data.(pb.Message); ok {
		return pb.Marshal(m)
	}

	return nil, fmt.Errorf("%T is not a proto.Message type", data)
}

func Decode(data []byte, i interface{}) error {
	if m, ok := i.(proto.Unmarshaler); ok {
		return m.Unmarshal(data)
	}

	if m, ok := i.(pb.Message); ok {
		return pb.Unmarshal(data, m)
	}

	return fmt.Errorf("%T is not a proto.Unmarshaler", i)
}
