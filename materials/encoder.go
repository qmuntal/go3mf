package materials

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/qmuntal/go3mf/spec/encoding"
)

// Marshal3MF encodes the resource.
func (r *ColorGroup) Marshal3MF(x encoding.XMLEncoder) error {
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrColorGroup}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(r.ID), 10)},
	}}
	x.EncodeToken(xs)
	x.SetAutoClose(true)
	for _, c := range r.Colors {
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrColor}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrColor}, Value: encoding.FormatRGBA(c)},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xs.End())
	return nil
}

// Marshal3MF encodes the resource.
func (r *Texture2DGroup) Marshal3MF(x encoding.XMLEncoder) error {
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrTexture2DGroup}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(r.ID), 10)},
		{Name: xml.Name{Local: attrTexID}, Value: strconv.FormatUint(uint64(r.TextureID), 10)},
	}}
	x.EncodeToken(xs)
	x.SetAutoClose(true)
	prec := x.FloatPresicion()
	for _, c := range r.Coords {
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrTex2DCoord}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrU}, Value: strconv.FormatFloat(float64(c.U()), 'f', prec, 32)},
			{Name: xml.Name{Local: attrV}, Value: strconv.FormatFloat(float64(c.V()), 'f', prec, 32)},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xs.End())
	return nil
}

// Marshal3MF encodes the resource.
func (r *CompositeMaterials) Marshal3MF(x encoding.XMLEncoder) error {
	indices := make([]string, len(r.Indices))
	for i, idx := range r.Indices {
		indices[i] = strconv.FormatUint(uint64(idx), 10)
	}
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrCompositematerials}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(r.ID), 10)},
		{Name: xml.Name{Local: attrMatID}, Value: strconv.FormatUint(uint64(r.MaterialID), 10)},
		{Name: xml.Name{Local: attrMatIndices}, Value: strings.Join(indices, " ")},
	}}
	x.EncodeToken(xs)
	x.SetAutoClose(true)
	for _, c := range r.Composites {
		values := make([]string, len(c.Values))
		for i, v := range c.Values {
			values[i] = strconv.FormatFloat(float64(v), 'f', x.FloatPresicion(), 32)
		}
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrComposite}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrValues}, Value: strings.Join(values, " ")},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xs.End())
	return nil
}

// Marshal3MF encodes the resource.
func (r *MultiProperties) Marshal3MF(x encoding.XMLEncoder) error {
	pids := make([]string, len(r.PIDs))
	for i, idx := range r.PIDs {
		pids[i] = strconv.FormatUint(uint64(idx), 10)
	}
	methods := make([]string, len(r.BlendMethods))
	for i, method := range r.BlendMethods {
		methods[i] = method.String()
	}
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrMultiProps}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(r.ID), 10)},
		{Name: xml.Name{Local: attrPIDs}, Value: strings.Join(pids, " ")},
		{Name: xml.Name{Local: attrBlendMethods}, Value: strings.Join(methods, " ")},
	}}
	x.EncodeToken(xs)
	x.SetAutoClose(true)
	for _, mu := range r.Multis {
		indices := make([]string, len(mu.PIndices))
		for i, v := range mu.PIndices {
			indices[i] = strconv.FormatUint(uint64(v), 10)
		}
		x.EncodeToken(xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrMulti}, Attr: []xml.Attr{
			{Name: xml.Name{Local: attrPIndices}, Value: strings.Join(indices, " ")},
		}})
	}
	x.SetAutoClose(false)
	x.EncodeToken(xs.End())
	return nil
}

// Marshal3MF encodes the resource.
func (r *Texture2D) Marshal3MF(x encoding.XMLEncoder) error {
	x.AddRelationship(encoding.Relationship{Path: r.Path, Type: RelTypeTexture3D})
	xs := xml.StartElement{Name: xml.Name{Space: Namespace, Local: attrTexture2D}, Attr: []xml.Attr{
		{Name: xml.Name{Local: attrID}, Value: strconv.FormatUint(uint64(r.ID), 10)},
		{Name: xml.Name{Local: attrPath}, Value: r.Path},
		{Name: xml.Name{Local: attrContentType}, Value: r.ContentType.String()},
	}}
	if r.TileStyleU != TileWrap {
		xs.Attr = append(xs.Attr, xml.Attr{Name: xml.Name{Local: attrTileStyleU}, Value: r.TileStyleU.String()})
	}
	if r.TileStyleV != TileWrap {
		xs.Attr = append(xs.Attr, xml.Attr{Name: xml.Name{Local: attrTileStyleV}, Value: r.TileStyleV.String()})
	}
	if r.Filter != TextureFilterAuto {
		xs.Attr = append(xs.Attr, xml.Attr{Name: xml.Name{Local: attrFilter}, Value: r.Filter.String()})
	}
	x.SetAutoClose(true)
	x.EncodeToken(xs)
	x.SetAutoClose(false)
	return nil
}
