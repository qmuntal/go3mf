package go3mf

import (
	"bytes"
	"encoding/xml"
	"image/color"
	"strconv"
	"strings"
	"unsafe"

	specerr "github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
)

type modelDecoder struct {
	baseDecoder
	ctx    *decoderContext
	model  *Model
	isRoot bool
}

func (d *modelDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace {
		switch name.Local {
		case attrResources:
			resources, _ := d.model.FindResources(d.ctx.modelPath)
			child = &resourceDecoder{resources: resources, ctx: d.ctx}
		case attrBuild:
			if d.isRoot {
				child = &buildDecoder{build: &d.model.Build, ctx: d.ctx}
			}
		case attrMetadata:
			if d.isRoot {
				child = &metadataDecoder{metadatas: &d.model.Metadata, ctx: d.ctx}
			}
		}
	} else if ext, ok := d.ctx.extensionDecoder[name.Space]; ok {
		child = ext.NewElementDecoder(encoding.ElementDecoderContext{
			ParentElement: d.model,
			Name:          name,
		})
	}
	return
}

func (d *modelDecoder) Start(attrs []encoding.XMLAttr) (err error) {
	if !d.isRoot {
		return
	}
	var requiredExts []string
	for _, a := range attrs {
		if a.Name.Space == "" {
			switch a.Name.Local {
			case attrUnit:
				var ok bool
				if d.model.Units, ok = newUnits(string(a.Value)); !ok {
					err = specerr.Append(err, specerr.NewParseAttrError(a.Name.Local, false))
				}
			case attrThumbnail:
				d.model.Thumbnail = string(a.Value)
			case attrReqExt:
				requiredExts = strings.Fields(string(a.Value))
			}
		} else {
			err = specerr.Append(err, d.noCoreAttribute(a))
		}
	}

	for _, ext := range requiredExts {
		for _, x := range d.model.Specs {
			if x.Local() == ext {
				x.SetRequired(true)
				break
			}
		}
	}
	return
}

func (d *modelDecoder) noCoreAttribute(a encoding.XMLAttr) (err error) {
	switch a.Name.Space {
	case nsXML:
		if a.Name.Local == attrLang {
			d.model.Language = string(a.Value)
		}
	case attrXmlns:
		if ext, ok := d.model.Specs[string(a.Value)]; ok {
			ext.SetLocal(a.Name.Local)
		} else {
			d.model.WithSpec(&spec.UnknownSpec{SpaceName: string(a.Value), LocalName: a.Name.Local})
		}
	default:
		if ext, ok := d.ctx.extensionDecoder[a.Name.Space]; ok {
			err = specerr.Append(err, ext.DecodeAttribute(d.model, a))
		}
	}
	return
}

type metadataGroupDecoder struct {
	baseDecoder
	ctx       *decoderContext
	metadatas *[]Metadata
}

func (d *metadataGroupDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrMetadata {
		child = &metadataDecoder{metadatas: d.metadatas, ctx: d.ctx}
	}
	return
}

type metadataDecoder struct {
	baseDecoder
	ctx       *decoderContext
	metadatas *[]Metadata
	metadata  Metadata
}

func (d *metadataDecoder) Start(attrs []encoding.XMLAttr) error {
	for _, a := range attrs {
		if a.Name.Space != "" {
			continue
		}
		switch a.Name.Local {
		case attrName:
			d.metadata.Name = a.Name
			i := bytes.IndexByte(a.Value, ':')
			if i < 0 {
				d.metadata.Name.Local = string(a.Value)
			} else if _, ok := d.ctx.namespace(string(a.Value[0:i])); ok {
				d.metadata.Name.Space = string(a.Value[0:i])
				d.metadata.Name.Local = string(a.Value[i+1:])
			} else {
				d.metadata.Name.Local = string(a.Value)
			}
		case attrType:
			d.metadata.Type = string(a.Value)
		case attrPreserve:
			d.metadata.Preserve, _ = strconv.ParseBool(string(a.Value))
		}
	}
	return nil
}

func (d *metadataDecoder) CharData(txt []byte) {
	d.metadata.Value = string(txt)
}

func (d *metadataDecoder) End() {
	*d.metadatas = append(*d.metadatas, d.metadata)
}

type buildDecoder struct {
	baseDecoder
	ctx   *decoderContext
	build *Build
}

func (d *buildDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrItem {
		child = &buildItemDecoder{build: d.build, ctx: d.ctx}
	}
	return
}

func (d *buildDecoder) Wrap(err error) error {
	return specerr.Wrap(err, d.build)
}

func (d *buildDecoder) Start(attrs []encoding.XMLAttr) error {
	var errs error
	for _, a := range attrs {
		if ext, ok := d.ctx.extensionDecoder[a.Name.Space]; ok {
			errs = specerr.Append(errs, ext.DecodeAttribute(d.build, a))
		}
	}
	if errs != nil {
		return d.Wrap(errs)
	}
	return errs
}

type buildItemDecoder struct {
	baseDecoder
	ctx   *decoderContext
	build *Build
	item  Item
	ew    encoding.ErrorWrapper
}

func (d *buildItemDecoder) End() {
	d.build.Items = append(d.build.Items, &d.item)
}

func (d *buildItemDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.WrapIndex(err, &d.item, len(d.build.Items)))
}

func (d *buildItemDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrMetadataGroup {
		child = &metadataGroupDecoder{metadatas: &d.item.Metadata, ctx: d.ctx}
	}
	return
}

func (d *buildItemDecoder) Start(attrs []encoding.XMLAttr) error {
	var errs error
	for _, a := range attrs {
		if a.Name.Space == "" {
			errs = specerr.Append(errs, d.parseCoreAttr(a))
		} else if ext, ok := d.ctx.extensionDecoder[a.Name.Space]; ok {
			errs = specerr.Append(errs, ext.DecodeAttribute(&d.item, a))
		}
	}
	if errs != nil {
		return specerr.WrapIndex(errs, &d.item, len(d.build.Items))
	}
	return errs
}

func (d *buildItemDecoder) parseCoreAttr(a encoding.XMLAttr) (errs error) {
	switch a.Name.Local {
	case attrObjectID:
		val, err := strconv.ParseUint(string(a.Value), 10, 32)
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
		}
		d.item.ObjectID = uint32(val)
	case attrPartNumber:
		d.item.PartNumber = string(a.Value)
	case attrTransform:
		var ok bool
		d.item.Transform, ok = encoding.ParseMatrix(string(a.Value))
		if !ok {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
		}
	}
	return
}

type resourceDecoder struct {
	baseDecoder
	ctx       *decoderContext
	resources *Resources
}

func (d *resourceDecoder) Wrap(err error) error {
	return specerr.Wrap(err, d.resources)
}

func (d *resourceDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace {
		switch name.Local {
		case attrObject:
			child = &objectDecoder{resources: d.resources, ctx: d.ctx, ew: d}
		case attrBaseMaterials:
			child = &baseMaterialsDecoder{resources: d.resources, ew: d}
		}
	} else if ext, ok := d.ctx.extensionDecoder[name.Space]; ok {
		child = ext.NewElementDecoder(encoding.ElementDecoderContext{
			ParentElement: d.resources,
			Name:          name,
			ErrorWrapper:  d,
		})
	}
	return
}

type baseMaterialsDecoder struct {
	baseDecoder
	resources           *Resources
	resource            BaseMaterials
	baseMaterialDecoder baseMaterialDecoder
	ew                  encoding.ErrorWrapper
}

func (d *baseMaterialsDecoder) End() {
	d.resources.Assets = append(d.resources.Assets, &d.resource)
}

func (d *baseMaterialsDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.WrapIndex(err, &d.resource, len(d.resources.Assets)))
}

func (d *baseMaterialsDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrBase {
		child = &d.baseMaterialDecoder
	}
	return
}

func (d *baseMaterialsDecoder) Start(attrs []encoding.XMLAttr) error {
	var errs error
	d.baseMaterialDecoder.resource = &d.resource
	for _, a := range attrs {
		if a.Name.Space == "" && a.Name.Local == attrID {
			id, err := strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			d.resource.ID = uint32(id)
			break
		}
	}
	if errs != nil {
		return specerr.WrapIndex(errs, &d.resource, len(d.resources.Assets))
	}
	return nil
}

type baseMaterialDecoder struct {
	baseDecoder
	resource *BaseMaterials
}

func (d *baseMaterialDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		name      string
		baseColor color.RGBA
		errs      error
	)
	for _, a := range attrs {
		switch a.Name.Local {
		case attrName:
			name = string(a.Value)
		case attrDisplayColor:
			var err error
			baseColor, err = encoding.ParseRGBA(string(a.Value))
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
		}
	}
	d.resource.Materials = append(d.resource.Materials, Base{Name: name, Color: baseColor})
	if errs != nil {
		return specerr.WrapIndex(errs, baseColor, len(d.resource.Materials)-1)
	}
	return nil
}

type meshDecoder struct {
	baseDecoder
	ctx      *decoderContext
	resource *Object
	ew       encoding.ErrorWrapper
}

func (d *meshDecoder) Start(_ []encoding.XMLAttr) error {
	d.resource.Mesh = new(Mesh)
	return nil
}

func (d *meshDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.Wrap(err, d.resource.Mesh))
}

func (d *meshDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace {
		if name.Local == attrVertices {
			child = &verticesDecoder{mesh: d.resource.Mesh, ew: d}
		} else if name.Local == attrTriangles {
			child = &trianglesDecoder{resource: d.resource, ew: d}
		}
	} else if ext, ok := d.ctx.extensionDecoder[name.Space]; ok {
		child = ext.NewElementDecoder(encoding.ElementDecoderContext{
			ParentElement: d.resource.Mesh,
			Name:          name,
			ErrorWrapper:  d,
		})
	}
	return
}

type verticesDecoder struct {
	baseDecoder
	mesh          *Mesh
	vertexDecoder vertexDecoder
	ew            encoding.ErrorWrapper
}

func (d *verticesDecoder) Start(_ []encoding.XMLAttr) error {
	d.vertexDecoder.mesh = d.mesh
	return nil
}

func (d *verticesDecoder) Wrap(err error) error {
	return d.ew.Wrap(err)
}

func (d *verticesDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrVertex {
		child = &d.vertexDecoder
	}
	return
}

type vertexDecoder struct {
	baseDecoder
	mesh *Mesh
}

func (d *vertexDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		x, y, z float32
		errs    error
	)
	for _, a := range attrs {
		val, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&a.Value)), 32)
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
		}
		switch a.Name.Local {
		case attrX:
			x = float32(val)
		case attrY:
			y = float32(val)
		case attrZ:
			z = float32(val)
		}
	}
	d.mesh.Vertices = append(d.mesh.Vertices, Point3D{x, y, z})
	if errs != nil {
		return specerr.WrapIndex(errs, Point3D{x, y, z}, len(d.mesh.Vertices)-1)
	}
	return nil
}

type trianglesDecoder struct {
	baseDecoder
	resource        *Object
	triangleDecoder triangleDecoder
	ew              encoding.ErrorWrapper
}

func (d *trianglesDecoder) Wrap(err error) error {
	return d.ew.Wrap(err)
}

func (d *trianglesDecoder) Start(_ []encoding.XMLAttr) error {
	d.triangleDecoder.mesh = d.resource.Mesh
	d.triangleDecoder.defaultPropertyID = d.resource.PID
	d.triangleDecoder.defaultPropertyIndex = d.resource.PIndex

	if len(d.resource.Mesh.Triangles) == 0 && len(d.resource.Mesh.Vertices) > 0 {
		d.resource.Mesh.Triangles = make([]Triangle, 0, len(d.resource.Mesh.Vertices)*2)
	}
	return nil
}

func (d *trianglesDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrTriangle {
		child = &d.triangleDecoder
	}
	return
}

type triangleDecoder struct {
	baseDecoder
	mesh                                    *Mesh
	defaultPropertyIndex, defaultPropertyID uint32
}

func (d *triangleDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		t                           Triangle
		pid, p1, p2, p3             uint32
		hasPID, hasP1, hasP2, hasP3 bool
		errs                        error
	)

	for _, a := range attrs {
		required := true
		val, err := strconv.ParseUint(string(a.Value), 10, 24)
		switch a.Name.Local {
		case attrV1:
			t[0] = ToUint24(uint32(val))
		case attrV2:
			t[1] = ToUint24(uint32(val))
		case attrV3:
			t[2] = ToUint24(uint32(val))
		case attrPID:
			pid = uint32(val)
			hasPID = true
			required = false
		case attrP1:
			p1 = uint32(val)
			hasP1 = true
			required = false
		case attrP2:
			p2 = uint32(val)
			hasP2 = true
			required = false
		case attrP3:
			p3 = uint32(val)
			hasP3 = true
			required = false
		}
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, required))
		}
	}

	p1 = applyDefault(p1, d.defaultPropertyIndex, hasP1)
	p2 = applyDefault(p2, p1, hasP2)
	p3 = applyDefault(p3, p1, hasP3)
	pid = applyDefault(pid, d.defaultPropertyID, hasPID)
	t.SetPID(pid)
	t.SetPIndices(p1, p2, p3)
	d.mesh.Triangles = append(d.mesh.Triangles, t)
	if errs != nil {
		return specerr.WrapIndex(errs, t, len(d.mesh.Triangles)-1)
	}
	return nil
}

func applyDefault(val, defVal uint32, noDef bool) uint32 {
	if noDef {
		return val
	}
	return defVal
}

type objectDecoder struct {
	baseDecoder
	ctx       *decoderContext
	resources *Resources
	resource  Object
	ew        encoding.ErrorWrapper
}

func (d *objectDecoder) End() {
	d.resources.Objects = append(d.resources.Objects, &d.resource)
}

func (d *objectDecoder) Start(attrs []encoding.XMLAttr) error {
	var errs error
	for _, a := range attrs {
		if a.Name.Space == "" {
			errs = specerr.Append(errs, d.parseCoreAttr(a))
		} else if ext, ok := d.ctx.extensionDecoder[a.Name.Space]; ok {
			errs = specerr.Append(errs, ext.DecodeAttribute(&d.resource, a))
		}
	}
	if errs != nil {
		return specerr.WrapIndex(errs, &d.resource, len(d.resources.Objects))
	}
	return errs
}

func (d *objectDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.WrapIndex(err, &d.resource, len(d.resources.Objects)))
}

func (d *objectDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace {
		if name.Local == attrMesh {
			child = &meshDecoder{resource: &d.resource, ctx: d.ctx, ew: d}
		} else if name.Local == attrComponents {
			child = &componentsDecoder{resource: &d.resource, ctx: d.ctx, ew: d}
		} else if name.Local == attrMetadataGroup {
			child = &metadataGroupDecoder{metadatas: &d.resource.Metadata, ctx: d.ctx}
		}
	}
	return
}

func (d *objectDecoder) parseCoreAttr(a encoding.XMLAttr) (errs error) {
	switch a.Name.Local {
	case attrID:
		id, err := strconv.ParseUint(string(a.Value), 10, 32)
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
		}
		d.resource.ID = uint32(id)
	case attrType:
		var ok bool
		d.resource.Type, ok = newObjectType(string(a.Value))
		if !ok {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
		}
	case attrThumbnail:
		d.resource.Thumbnail = string(a.Value)
	case attrName:
		d.resource.Name = string(a.Value)
	case attrPartNumber:
		d.resource.PartNumber = string(a.Value)
	case attrPID:
		val, err := strconv.ParseUint(string(a.Value), 10, 32)
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
		}
		d.resource.PID = uint32(val)
	case attrPIndex:
		val, err := strconv.ParseUint(string(a.Value), 10, 32)
		if err != nil {
			errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
		}
		d.resource.PIndex = uint32(val)
	}
	return
}

type componentsDecoder struct {
	baseDecoder
	ctx              *decoderContext
	resource         *Object
	componentDecoder componentDecoder
	ew               encoding.ErrorWrapper
}

func (d *componentsDecoder) Start(_ []encoding.XMLAttr) error {
	d.resource.Components = make([]*Component, 0)
	d.componentDecoder.resource = d.resource
	d.componentDecoder.ctx = d.ctx
	return nil
}

func (d *componentsDecoder) Wrap(err error) error {
	return d.ew.Wrap(err)
}

func (d *componentsDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrComponent {
		child = &d.componentDecoder
	}
	return
}

type componentDecoder struct {
	baseDecoder
	ctx      *decoderContext
	resource *Object
}

func (d *componentDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		component Component
		errs      error
	)
	for _, a := range attrs {
		if a.Name.Space == "" {
			if a.Name.Local == attrObjectID {
				val, err := strconv.ParseUint(string(a.Value), 10, 32)
				if err != nil {
					errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
				}
				component.ObjectID = uint32(val)
			} else if a.Name.Local == attrTransform {
				var ok bool
				component.Transform, ok = encoding.ParseMatrix(string(a.Value))
				if !ok {
					errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
				}
			}
		} else if ext, ok := d.ctx.extensionDecoder[a.Name.Space]; ok {
			errs = specerr.Append(errs, ext.DecodeAttribute(&component, a))
		}
	}
	d.resource.Components = append(d.resource.Components, &component)
	if errs != nil {
		return specerr.WrapIndex(errs, &component, len(d.resource.Components)-1)
	}
	return nil
}

type baseDecoder struct {
}

func (d *baseDecoder) Start([]encoding.XMLAttr) error { return nil }
func (d *baseDecoder) End()                        {}

type topLevelDecoder struct {
	baseDecoder
	ctx    *decoderContext
	model  *Model
	isRoot bool
}

func (d *topLevelDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	modelName := xml.Name{Space: Namespace, Local: attrModel}
	if name == modelName {
		child = &modelDecoder{model: d.model, isRoot: d.isRoot, ctx: d.ctx}
	}
	return
}
