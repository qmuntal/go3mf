package go3mf

import (
	"encoding/xml"
	"image/color"
	"io"
	"sort"
	"sync"

	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
)

const (
	// Namespace is the canonical name of this extension.
	Namespace = "http://schemas.microsoft.com/3dmanufacturing/core/2015/02"

	// RelType3DModel is the canonical 3D model relationship type.
	RelType3DModel = "http://schemas.microsoft.com/3dmanufacturing/2013/01/3dmodel"
	// RelTypeThumbnail is the canonical thumbnail relationship type.
	RelTypeThumbnail = "http://schemas.openxmlformats.org/package/2006/relationships/metadata/thumbnail"
	// RelTypePrintTicket is the canonical print ticket relationship type.
	RelTypePrintTicket = "http://schemas.microsoft.com/3dmanufacturing/2013/01/printticket"
	// RelTypeMustPreserve is the canonical must preserve relationship type.
	RelTypeMustPreserve = "http://schemas.openxmlformats.org/package/2006/relationships/mustpreserve"

	// DefaultModelPath is the recommended root model part name.
	DefaultModelPath = "/3D/3dmodel.model"
	// DefaultPrintTicketName is the recommended print ticket part name.
	DefaultPrintTicketName = "/3D/Metadata/Model_PT.xml"
	// Default3DTexturesDir is the recommended directory for 3D textures.
	Default3DTexturesDir = "/3D/Textures/"
	// Default3DOtherDir is the recommended directory for non-standard parts.
	Default3DOtherDir = "/3D/Other/"
	// DefaultMetadataDir is the recommended directory for standard metadata.
	DefaultMetadataDir = "/Metadata/"

	// ContentType3DModel is the 3D model content type.
	ContentType3DModel = "application/vnd.ms-package.3dmanufacturing-3dmodel+xml"
	// ContentTypePrintTicket is the print ticket content type.
	ContentTypePrintTicket = "application/vnd.ms-printing.printticket+xml"
)

// Units define the allowed model units.
type Units uint8

// Supported units.
const (
	UnitMillimeter Units = iota
	UnitMicrometer
	UnitCentimeter
	UnitInch
	UnitFoot
	UnitMeter
)

func (u Units) String() string {
	return map[Units]string{
		UnitMillimeter: "millimeter",
		UnitMicrometer: "micron",
		UnitCentimeter: "centimeter",
		UnitInch:       "inch",
		UnitFoot:       "foot",
		UnitMeter:      "meter",
	}[u]
}

// ObjectType defines the allowed object types.
type ObjectType int8

// Supported object types.
const (
	ObjectTypeModel ObjectType = iota
	ObjectTypeOther
	ObjectTypeSupport
	ObjectTypeSolidSupport
	ObjectTypeSurface
)

func (o ObjectType) String() string {
	return map[ObjectType]string{
		ObjectTypeModel:        "model",
		ObjectTypeOther:        "other",
		ObjectTypeSupport:      "support",
		ObjectTypeSolidSupport: "solidsupport",
		ObjectTypeSurface:      "surface",
	}[o]
}

// Asset defines build resource.
type Asset interface {
	Identify() uint32
}

// Metadata item is an in memory representation of the 3MF metadata,
// and can be attached to any 3MF model node.
type Metadata struct {
	Name     xml.Name
	Value    string
	Type     string
	Preserve bool
}

// Attachment defines the Model Attachment.
type Attachment struct {
	Stream      io.Reader
	Path        string
	ContentType string
}

// Relationship defines a dependency between
// the owner of the relationsip and the attachment
// referenced by path. ID is optional, if not set a random
// value will be used when encoding.
type Relationship struct {
	Path string
	Type string
	ID   string
}

// Build contains one or more items to manufacture as part of processing the job.
type Build struct {
	Items   []*Item
	AnyAttr AnyAttr
}

// The Resources element acts as the root element of a library of constituent
// pieces of the overall 3D object definition.
type Resources struct {
	Assets  []Asset
	Objects []*Object
	AnyAttr AnyAttr
}

// UnusedID returns the lowest unused ID.
func (rs *Resources) UnusedID() uint32 {
	if len(rs.Assets) == 0 && len(rs.Objects) == 0 {
		return 1
	}
	ids := make([]int, len(rs.Assets)+len(rs.Objects)+1)
	ids[0] = 0
	for i, r := range rs.Assets {
		id := r.Identify()
		ids[i+1] = int(id)
	}
	for i, o := range rs.Objects {
		ids[len(rs.Assets)+i+1] = int(o.ID)
	}
	sort.Ints(ids)
	lowest := 0
	for i, id := range ids {
		if id != i {
			lowest = i
		}
	}
	if lowest == 0 {
		lowest = ids[len(ids)-1] + 1
	}
	return uint32(lowest)
}

// FindObject returns the resource with the target ID.
func (rs *Resources) FindObject(id uint32) (*Object, bool) {
	for _, value := range rs.Objects {
		if value.ID == id {
			return value, true
		}
	}
	return nil, false
}

// FindAsset returns the resource with the target ID.
func (rs *Resources) FindAsset(id uint32) (Asset, bool) {
	for _, value := range rs.Assets {
		if rID := value.Identify(); rID == id {
			return value, true
		}
	}
	return nil, false
}

// ChildModel repreents de content of a non-root model file.
//
// It is not supported by the core spec but a common concept
// for multiple official specs.
// The relationships are usually managed by the extensions themself,
// but they are usefull to reference custom attachments.
type ChildModel struct {
	Resources     Resources
	Relationships []Relationship
	Any           Any
}

// A Model is an in memory representation of the 3MF file.
//
// If path is empty, the default path '/3D/3dmodel.model' will be used.
// The relationships are usually managed by the extensions themself,
// but they are usefull to reference custom attachments.
// Childs keys cannot be an empty string.
// RootRelationships are the OPC root relationships.
type Model struct {
	Path              string
	Language          string
	Units             Units
	Thumbnail         string
	Resources         Resources
	Build             Build
	Attachments       []Attachment
	Specs             map[string]spec.Spec // space -> spec
	Metadata          []Metadata
	Childs            map[string]*ChildModel // path -> child
	RootRelationships []Relationship
	Relationships     []Relationship
	Any               Any
	AnyAttr           AnyAttr
}

// WithSpec adds a new extension
func (m *Model) WithSpec(extension spec.Spec) {
	if m.Specs == nil {
		m.Specs = make(map[string]spec.Spec)
	}
	m.Specs[extension.Namespace()] = extension
}

// PathOrDefault returns Path if not empty, else DefaultModelPath.
func (m *Model) PathOrDefault() string {
	if m.Path == "" {
		return DefaultModelPath
	}
	return m.Path
}

// BoundingBox returns the bounding box of the model.
func (m *Model) BoundingBox() Box {
	if len(m.Build.Items) == 0 {
		return Box{}
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	box := newLimitBox()
	wg.Add(len(m.Build.Items))
	for i := range m.Build.Items {
		go func(i int) {
			defer wg.Done()
			item := m.Build.Items[i]
			if o, ok := m.FindObject(item.ObjectPath(), item.ObjectID); ok {
				ibox := o.boundingBox(m, "")
				if ibox != emptyBox {
					mu.Lock()
					box = box.extend(item.Transform.MulBox(ibox))
					mu.Unlock()
				}
			}
		}(i)
	}
	wg.Wait()
	return box
}

// FindResources returns the resource associated with path.
func (m *Model) FindResources(path string) (*Resources, bool) {
	if path == "" || path == m.Path || (m.Path == "" && path == DefaultModelPath) {
		return &m.Resources, true
	}
	if child, ok := m.Childs[path]; ok {
		return &child.Resources, true
	}
	return nil, false
}

// FindAsset returns the resource with the target path and ID.
func (m *Model) FindAsset(path string, id uint32) (Asset, bool) {
	if rs, ok := m.FindResources(path); ok {
		return rs.FindAsset(id)
	}
	return nil, false
}

// FindObject returns the object with the target path and ID.
func (m *Model) FindObject(path string, id uint32) (*Object, bool) {
	if rs, ok := m.FindResources(path); ok {
		return rs.FindObject(id)
	}
	return nil, false
}

// WalkAssets walks the assets of the root and child models, calling fn for asset and stopping
// if fn returns an error.
//
// The child models are first walked in lexical order and then the root model is walked.
// The root model path is always empty, regardless of the defined model path.
func (m *Model) WalkAssets(fn func(string, Asset) error) error {
	sortedChilds := m.sortedChilds()
	for _, path := range sortedChilds {
		c := m.Childs[path]
		for _, r := range c.Resources.Assets {
			if err := fn(path, r); err != nil {
				return err
			}
		}
	}
	for _, r := range m.Resources.Assets {
		if err := fn("", r); err != nil {
			return err
		}
	}
	return nil
}

// WalkObjects walks the objects of the root and child models, calling fn for object and stopping
// if fn returns an error.
//
// The child models are first walked in lexical order and then the root model is walked.
// The root model path is always empty, regardless of the defined model path.
func (m *Model) WalkObjects(fn func(string, *Object) error) error {
	sortedChilds := m.sortedChilds()
	for _, path := range sortedChilds {
		c := m.Childs[path]
		for _, r := range c.Resources.Objects {
			if err := fn(path, r); err != nil {
				return err
			}
		}
	}
	for _, r := range m.Resources.Objects {
		if err := fn("", r); err != nil {
			return err
		}
	}
	return nil
}

// Base defines the Model Base Material Resource.
// A model material resource is an in memory representation of the 3MF
// material resource object.
type Base struct {
	Name  string
	Color color.RGBA
}

// BaseMaterials defines a slice of Base.
type BaseMaterials struct {
	ID        uint32
	Materials []Base
}

// Len returns the materials count.
func (r *BaseMaterials) Len() int {
	return len(r.Materials)
}

// Identify returns the unique ID of the resource.
func (r *BaseMaterials) Identify() uint32 {
	return r.ID
}

// A Item is an in memory representation of the 3MF build item.
type Item struct {
	ObjectID   uint32
	Transform  Matrix
	PartNumber string
	Metadata   []Metadata
	AnyAttr    AnyAttr
}

// ObjectPath search an extension attribute with an ObjectPath
// function that return a non empty path.
// Else returns an empty path.
func (b *Item) ObjectPath() string {
	for _, att := range b.AnyAttr {
		if ext, ok := att.(objectPather); ok {
			path := ext.ObjectPath()
			if path != "" {
				return path
			}
		}
	}
	return ""
}

// HasTransform returns true if the transform is different than the identity.
func (b *Item) HasTransform() bool {
	return b.Transform != Matrix{} && b.Transform != Identity()
}

// An Object is an in memory representation of the 3MF model object.
type Object struct {
	ID         uint32
	Name       string
	PartNumber string
	Thumbnail  string
	PID        uint32
	PIndex     uint32
	Type       ObjectType
	Metadata   []Metadata
	Mesh       *Mesh
	Components []*Component
	AnyAttr    AnyAttr
}

func (o *Object) boundingBox(m *Model, path string) Box {
	if o.Mesh != nil {
		return o.Mesh.BoundingBox()
	}
	if len(o.Components) == 0 {
		return Box{}
	}
	box := newLimitBox()
	for _, c := range o.Components {
		if obj, ok := m.FindObject(c.ObjectPath(path), c.ObjectID); ok {
			cbox := obj.boundingBox(m, path)
			if cbox != emptyBox {
				box = box.extend(c.Transform.MulBox(cbox))
			}
		}
	}
	return box
}

// A Component is an in memory representation of the 3MF component.
type Component struct {
	ObjectID  uint32
	Transform Matrix
	AnyAttr   AnyAttr
}

// ObjectPath search an extension attribute with an ObjectPath
// function that return a non empty path.
// Else returns the default path.
func (c *Component) ObjectPath(defaultPath string) string {
	for _, att := range c.AnyAttr {
		if ext, ok := att.(objectPather); ok {
			path := ext.ObjectPath()
			if path != "" {
				return path
			}
		}
	}
	return defaultPath
}

// HasTransform returns true if the transform is different than the identity.
func (c *Component) HasTransform() bool {
	return c.Transform != Matrix{} && c.Transform != Identity()
}

// Triangle defines a triangle of a mesh.
//
// The 7 elements are: v1,v2,v3,pid,p1,p2,p3.
type Triangle [7]Uint24

func NewTriangle(v1, v2, v3 uint32) (t Triangle) {
	t.SetIndices(v1, v2, v3)
	return
}

func NewTrianglePID(v1, v2, v3, pid, p1, p2, p3 uint32) Triangle {
	t := NewTriangle(v1, v2, v3)
	t.SetPID(pid)
	t.SetPIndices(p1, p2, p3)
	return t
}

func (t *Triangle) SetIndices(v1, v2, v3 uint32) {
	t[0] = ToUint24(v1)
	t[1] = ToUint24(v2)
	t[2] = ToUint24(v3)
}

func (t *Triangle) SetPID(pid uint32) {
	t[3] = ToUint24(pid)
}

func (t *Triangle) SetPIndices(p1, p2, p3 uint32) {
	t[4] = ToUint24(p1)
	t[5] = ToUint24(p2)
	t[6] = ToUint24(p3)
}

func (t Triangle) Indices() (uint32, uint32, uint32) {
	return t[0].ToUint32(), t[1].ToUint32(), t[2].ToUint32()
}

func (t Triangle) PID() uint32 {
	return t[3].ToUint32()
}

func (t Triangle) PIndices() (uint32, uint32, uint32) {
	return t[4].ToUint32(), t[5].ToUint32(), t[6].ToUint32()
}

// A Mesh is an in memory representation of the 3MF mesh object.
// Each node and face have an ID, which allows to identify them. Each face have an
// orientation (i.e. the face can look up or look down) and have three nodes.
// The orientation is defined by the order of its nodes.
type Mesh struct {
	Vertices  []Point3D
	Triangles []Triangle
	AnyAttr   AnyAttr
	Any       Any
}

// BoundingBox returns the bounding box of the mesh.
func (m *Mesh) BoundingBox() Box {
	if len(m.Vertices) == 0 {
		return Box{}
	}
	box := newLimitBox()
	for _, v := range m.Vertices {
		box = box.extendPoint(v)
	}
	return box
}

// MeshBuilder is a helper that creates mesh following a configurable criteria.
// It must be instantiated using NewMeshBuilder.
type MeshBuilder struct {
	// True to automatically check if a node with the same coordinates already exists in the mesh
	// when calling AddVertex. If it exists, the return value will be the existing node and no node will be added.
	// Using this option produces an speed penalty.
	CalculateConnectivity bool
	// Do not modify the pointer to Mesh once the build process has started.
	Mesh       *Mesh
	vectorTree vectorTree
}

// NewMeshBuilder returns a new MeshBuilder.
func NewMeshBuilder(m *Mesh) *MeshBuilder {
	return &MeshBuilder{
		Mesh:                  m,
		CalculateConnectivity: true,
		vectorTree:            vectorTree{},
	}
}

// AddVertex adds a node the the mesh at the target position.
func (mb *MeshBuilder) AddVertex(node Point3D) uint32 {
	if mb.CalculateConnectivity {
		if index, ok := mb.vectorTree.FindVector(node); ok {
			return index
		}
	}
	mb.Mesh.Vertices = append(mb.Mesh.Vertices, node)
	index := uint32(len(mb.Mesh.Vertices)) - 1
	if mb.CalculateConnectivity {
		mb.vectorTree.AddVector(node, index)
	}
	return index
}

func newObjectType(s string) (o ObjectType, ok bool) {
	o, ok = map[string]ObjectType{
		"model":        ObjectTypeModel,
		"other":        ObjectTypeOther,
		"support":      ObjectTypeSupport,
		"solidsupport": ObjectTypeSolidSupport,
		"surface":      ObjectTypeSurface,
	}[s]
	return
}

func newUnits(s string) (u Units, ok bool) {
	u, ok = map[string]Units{
		"millimeter": UnitMillimeter,
		"micron":     UnitMicrometer,
		"centimeter": UnitCentimeter,
		"inch":       UnitInch,
		"foot":       UnitFoot,
		"meter":      UnitMeter,
	}[s]
	return
}

// AnyAttr is an extension point containing <anyAttribute> information.
// The key should be the extension namespace.
type AnyAttr []encoding.MarshalerAttr

// Any is an extension point containing <any> information.
// The key should be the extension namespace.
type Any []encoding.Marshaler

type objectPather interface {
	ObjectPath() string
}

const (
	nsXML   = "http://www.w3.org/XML/1998/namespace"
	nsXMLNs = "http://www.w3.org/2000/xmlns/"
)

const (
	attrXML           = "xml"
	attrXmlns         = "xmlns"
	attrID            = "id"
	attrName          = "name"
	attrObjectID      = "objectid"
	attrTransform     = "transform"
	attrUnit          = "unit"
	attrReqExt        = "requiredextensions"
	attrLang          = "lang"
	attrResources     = "resources"
	attrBuild         = "build"
	attrObject        = "object"
	attrBaseMaterials = "basematerials"
	attrBase          = "base"
	attrDisplayColor  = "displaycolor"
	attrPartNumber    = "partnumber"
	attrItem          = "item"
	attrModel         = "model"
	attrVertices      = "vertices"
	attrVertex        = "vertex"
	attrX             = "x"
	attrY             = "y"
	attrZ             = "z"
	attrV1            = "v1"
	attrV2            = "v2"
	attrV3            = "v3"
	attrType          = "type"
	attrThumbnail     = "thumbnail"
	attrPID           = "pid"
	attrPIndex        = "pindex"
	attrMesh          = "mesh"
	attrComponents    = "components"
	attrComponent     = "component"
	attrTriangles     = "triangles"
	attrTriangle      = "triangle"
	attrP1            = "p1"
	attrP2            = "p2"
	attrP3            = "p3"
	attrPreserve      = "preserve"
	attrMetadata      = "metadata"
	attrMetadataGroup = "metadatagroup"
	attrPath          = "path"
)
