package packets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	SeekRequestID uint8 = iota
	SeekResponseID
	ReadRequestID
	ReadResponseID
)

type Packet interface{}

type SeekRequest struct {
	Offset int64
	Whence int64
}

type SeekResponse struct {
	N  int64
	Ok bool
}

type ReadRequest struct {
	Size int64
}

type ReadResponse struct {
	Data []byte
}

func writeFull(w io.Writer, data []byte) error {
	n, err := w.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("did not write full data")
	}
	return nil
}

func aggregateErrors(errs ...error) error {
	for _, e := range errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func WritePacket(w io.Writer, p Packet) error {
	var err error
	switch p := p.(type) {
	case *SeekRequest:
		err = aggregateErrors(
			writeFull(w, []byte{SeekRequestID}),
			binary.Write(w, binary.LittleEndian, p.Offset),
			binary.Write(w, binary.LittleEndian, p.Whence),
		)
	case *SeekResponse:
		err = aggregateErrors(
			writeFull(w, []byte{SeekResponseID}),
			binary.Write(w, binary.LittleEndian, p.N),
			binary.Write(w, binary.LittleEndian, p.Ok),
		)
	case *ReadRequest:
		err = aggregateErrors(
			writeFull(w, []byte{ReadRequestID}),
			binary.Write(w, binary.LittleEndian, p.Size),
		)
	case *ReadResponse:
		err = aggregateErrors(
			writeFull(w, []byte{ReadResponseID}),
			binary.Write(w, binary.LittleEndian, int64(len(p.Data))),
			writeFull(w, p.Data),
		)
	default:
		err = fmt.Errorf("undefined packet %v", p)
	}
	return err
}

func ReadPacket(r io.Reader) (Packet, error) {
	buf := make([]byte, 1)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	var rv Packet
	switch buf[0] {
	case SeekRequestID:
		p := &SeekRequest{}
		err = aggregateErrors(
			binary.Read(r, binary.LittleEndian, &p.Offset),
			binary.Read(r, binary.LittleEndian, &p.Whence),
		)
		rv = p
	case SeekResponseID:
		p := &SeekResponse{}
		err = aggregateErrors(
			binary.Read(r, binary.LittleEndian, &p.N),
			binary.Read(r, binary.LittleEndian, &p.Ok),
		)
		rv = p
	case ReadRequestID:
		p := &ReadRequest{}
		err = aggregateErrors(
			binary.Read(r, binary.LittleEndian, &p.Size),
		)
		rv = p
	case ReadResponseID:
		p := &ReadResponse{}
		var sz int64
		err = binary.Read(r, binary.LittleEndian, &sz)
		if err != nil {
			return nil, err
		}
		d := make([]byte, sz)
		_, err = io.ReadFull(r, d)
		p.Data = d
		rv = p
	default:
		err = fmt.Errorf("undefined packet id: %v", buf[0])
	}
	return rv, err
}
