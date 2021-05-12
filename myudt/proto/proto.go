package proto

import (
	"bytes"
	"encoding/gob"
	"github.com/google/uuid"
)

type Proto struct {
	Version string
	Id string
	Seq int
	Length int
	Data []byte
	Seek int64
}

type ACK struct {
	Id string
}

func newProtoId() string {
	return uuid.NewString()
}

func NewProto(data []byte, seq int, seek int64) Proto {
	defaultVersion := "1.0"
	length := len(data)
	id := newProtoId()
	newData := make([]byte, len(data))
	copy(newData, data)
	return Proto{
		Version: defaultVersion,
		Id: id,
		Seq: seq,
		Length: length,
		Data: newData,
		Seek: seek,
	}
}

func NewACK(id string) ACK {
	return ACK{id}
}

func Serialize(obj Proto) ([]byte, error) {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func Deserialize(data []byte) (Proto, error) {
	obj := Proto {}

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&obj)
	if err != nil {
		return Proto{}, err
	}

	return obj, nil

}

// Validate TODO add checksum
func Validate(packet Proto) bool {
	if len(packet.Data) == packet.Length &&
		packet.Version == "1.0" {
		return true
	} else {
		return false
	}
}

func ACKSer(obj ACK) ([]byte, error) {
	data := bytes.Buffer{}
	enc := gob.NewEncoder(&data)
	err := enc.Encode(obj)
	if err != nil {
		return nil, err
	}

	return data.Bytes(), nil
}

func ACKDes(data []byte) (ACK, error) {
	obj := ACK {}

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&obj)
	if err != nil {
		return ACK{}, err
	}

	return obj, nil
}

