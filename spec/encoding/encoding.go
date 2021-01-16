package encoding

import (
	"encoding/xml"

	"github.com/qmuntal/go3mf/spec"
)

type XMLAttr struct {
	Name  xml.Name
	Value []byte
}

type Relationship struct {
	Path string
	Type string
	ID   string
}

// Marshaler is the interface implemented by objects
// that can marshal themselves into valid XML elements.
type Marshaler interface {
	Marshal3MF(XMLEncoder) error
}

// MarshalerAttr is the interface implemented by objects that can marshal
// themselves into valid XML attributes.
type MarshalerAttr interface {
	Marshal3MFAttr(XMLEncoder) ([]xml.Attr, error)
}

type ElementDecoderContext struct {
	ParentElement interface{}
	Name          xml.Name
	ErrorWrapper  ErrorWrapper
}

// DecoderSpec must be implemented by specs that want to support
// direct decoding from xml.
type DecoderSpec interface {
	spec.Spec
	DecodeAttribute(parent interface{}, attr XMLAttr) error
	NewElementDecoder(ElementDecoderContext) ElementDecoder
}

type ErrorWrapper interface {
	Wrap(error) error
}

// ElementDecoder defines the minimum contract to decode a 3MF node.
type ElementDecoder interface {
	Start([]XMLAttr) error
	End()
}

// ChildElementDecoder must be implemented by element decoders
// that need decoding nested elements.
type ChildElementDecoder interface {
	Child(xml.Name) ElementDecoder
}

// CharDataElementDecoder must be implemented by element decoders
// that need to decode raw text.
type CharDataElementDecoder interface {
	CharData([]byte)
}

// XMLEncoder provides de necessary methods to encode specs.
// It should not be implemented by spec authors but
// will be provided be go3mf itself.
type XMLEncoder interface {
	AddRelationship(r Relationship)
	FloatPresicion() int
	EncodeToken(t xml.Token)
	Flush() error
	SetAutoClose(autoClose bool)
}
