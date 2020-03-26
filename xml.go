package go3mf

import (
	"bufio"
	"encoding/xml"
	"io"
)

// XMLEncoder is based on the encoding/xml.Encoder implementation.
// It is modified to allow custom local namespaces and selfclosing nodes.
type XMLEncoder struct {
	floatPresicion int
	relationships  []Relationship
	p              printer
}

// newXMLEncoder returns a new encoder that writes to w.
func newXMLEncoder(w io.Writer, floatPresicion int) *XMLEncoder {
	return &XMLEncoder{
		floatPresicion: floatPresicion,
		p:              printer{Writer: bufio.NewWriter(w)},
	}
}

// AddRelationship adds a relationship to the encoded model.
// Duplicated relationships will be removed before encoding.
func (enc *XMLEncoder) AddRelationship(r Relationship) {
	enc.relationships = append(enc.relationships, r)
}

// FloatPresicion returns the float presicion to use
// when encoding floats.
func (enc *XMLEncoder) FloatPresicion() int {
	return enc.floatPresicion
}

// EncodeToken writes the given XML token to the stream.
func (enc *XMLEncoder) EncodeToken(t xml.Token) {
	p := &enc.p
	switch t := t.(type) {
	case xml.StartElement:
		p.writeStart(&t)
	case xml.EndElement:
		p.writeEnd(t.Name)
	case xml.CharData:
		xml.EscapeText(p, t)
	}
}

// Flush flushes any buffered XML to the underlying writer.
func (enc *XMLEncoder) Flush() error {
	return enc.p.Flush()
}

// SetAutoClose define if a start token will be self closed.
// Callers should not end the start token if the encode is in
// auto close mode.
func (enc *XMLEncoder) SetAutoClose(autoClose bool) {
	enc.p.autoClose = autoClose
}

type printer struct {
	*bufio.Writer
	attrPrefix map[string]string // map name space -> prefix
	autoClose  bool
}

// createAttrPrefix finds the name space prefix attribute to use for the given name space,
// defining a new prefix if necessary. It returns the prefix.
func (p *printer) createAttrPrefix(attr *xml.Attr) string {
	if prefix := p.attrPrefix[attr.Name.Space]; prefix != "" {
		return prefix
	}
	if attr.Name.Space == nsXML {
		return attrXML
	}

	// Need to define a new name space.
	if p.attrPrefix == nil {
		p.attrPrefix = make(map[string]string)
	}

	ns, prefix := attr.Name.Space, attr.Name.Local
	if attr.Name.Space == attrXmlns {
		ns, prefix = attr.Value, attr.Name.Local
	}
	p.attrPrefix[ns] = prefix

	return attr.Name.Space
}

// EscapeString writes to p the properly escaped XML equivalent
// of the plain text data s.
func (p *printer) EscapeString(s string) {
	xml.EscapeText(p, []byte(s))
}

// writeStart writes the given start element.
func (p *printer) writeStart(start *xml.StartElement) {
	p.WriteByte('<')
	if start.Name.Space != "" {
		if prefix := p.attrPrefix[start.Name.Space]; prefix != "" {
			p.WriteString(prefix)
			p.WriteByte(':')
		}
	}
	p.WriteString(start.Name.Local)

	// Attributes
	for _, attr := range start.Attr {
		name := attr.Name
		if name.Local == "" {
			continue
		}
		p.WriteByte(' ')
		if name.Space != "" {
			p.WriteString(p.createAttrPrefix(&attr))
			p.WriteByte(':')
		}
		p.WriteString(name.Local)
		p.WriteString(`="`)
		p.EscapeString(attr.Value)
		p.WriteByte('"')
	}
	if p.autoClose {
		p.WriteByte('/')
	}
	p.WriteByte('>')
}

func (p *printer) writeEnd(name xml.Name) {
	p.WriteByte('<')
	p.WriteByte('/')
	if name.Space != "" {
		if prefix := p.attrPrefix[name.Space]; prefix != "" {
			p.WriteString(prefix)
			p.WriteByte(':')
		}
	}
	p.WriteString(name.Local)
	p.WriteByte('>')
}