package bar

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// BetterAnyResolver is expected to be used in jsonpb.Marshaler
// WARNING: DON'T SUPPORT Unmarshaler
// although we do have a way, but it's considered too hacky
// if there is actual need for unmarshaler, we can use the hacky way
var BetterAnyResolver *ar

type ar struct{}

func (*ar) Resolve(typeUrl string) (proto.Message, error) {
	mname := typeUrl
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}
	mt := proto.MessageType(mname)
	if mt == nil {
		return &valMsg{
			url: typeUrl,
		}, nil
	}
	return reflect.New(mt.Elem()).Interface().(proto.Message), nil
}

type valMsg struct {
	url string
	V   []byte
}

func (*valMsg) ProtoMessage() {}

func (m *valMsg) Reset() { *m = valMsg{} }
func (m *valMsg) String() string {
	return fmt.Sprintf("%x", m.V) // not compatible w/ pb oct, never expect to be called
}
func (m *valMsg) Unmarshal(b []byte) error {
	m.V = append([]byte(nil), b...)
	return nil
}

// if we define XXX_WellKnownType, then we re-use jsonpb internal logic so don't need
// to implement our own MarshalJSONPB, but this is considered too hacky
// func (*valMsg) XXX_WellKnownType() string { return "BytesValue" }
// NOTE if we want to properly support UnmarshalJSONPB, best way is still use XXX_WellKnownType
// due to (u *Unmarshaler) unmarshalValue has `delete(jsonFields, "@type")` if not wkt
func (m *valMsg) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	// json.Marshal does base64-encoding and converts result to bytestring.
	v, err := json.Marshal(m.V)
	if err != nil {
		return nil, err
	}
	// v contains quotes "", no need to add quotes in format string.
	ret := fmt.Sprintf(`{"@type":"%s","value":%s}`, m.url, string(v))
	return []byte(ret), nil
}
