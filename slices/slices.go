package slices

import (
	"errors"

	"github.com/qmuntal/go3mf"
)

// Namespace is the canonical name of this extension.
const Namespace = "http://schemas.microsoft.com/3dmanufacturing/slice/2015/07"

var (
	ErrSliceExtRequired          = errors.New("a 3MF package which uses low resolution objects MUST enlist the slice extension as required")
	ErrNonSliceStack             = errors.New("slicestackid MUST reference a slice stack resource")
	ErrSlicesAndRefs             = errors.New("may either contain slices or refs, but they MUST NOT contain both element types")
	ErrSliceRefSamePart          = errors.New("the path of the referenced slice stack MUST be different than the path of the original slice stack")
	ErrSliceRefRef               = errors.New("a referenced slice stack MUST NOT contain any further sliceref elements")
	ErrSliceSmallTopZ            = errors.New("slice ztop is smaller than stack zbottom")
	ErrSliceNoMonotonic          = errors.New("the first ztop in the next slicestack MUST be greater than the last ztop in the previous slicestack")
	ErrSliceInsufficientVertices = errors.New("slice MUST contain at least 2 vertices")
	ErrSliceInsufficientPolygons = errors.New("slice MUST contain at least 1 polygon")
	ErrSliceInsufficientSegments = errors.New("slice polygon MUST contain at least 1 segment")
	ErrSlicePolygonNotClosed     = errors.New("objects with type 'model' and 'solidsupport' MUST not reference slices with open polygons")
	ErrSliceInvalidTranform      = errors.New("any transform applied to an object that references a slice stack MUST be planar")
)

type Spec struct {
	LocalName  string
	IsRequired bool
}

func (e Spec) Namespace() string   { return Namespace }
func (e Spec) Required() bool      { return e.IsRequired }
func (e *Spec) SetRequired(r bool) { e.IsRequired = r }
func (e *Spec) SetLocal(l string)  { e.LocalName = l }

func (e Spec) Local() string {
	if e.LocalName != "" {
		return e.LocalName
	}
	return "s"
}

// A Segment element represents a single line segment (or edge) of a polygon.
// It runs from the vertex specified by the previous segment
// (or the startv Polygon attribute for the first segment) to the specified vertex, v2.
type Segment struct {
	V2  uint32
	PID uint32
	P1  uint32
	P2  uint32
}

// The Polygon element contains a set of 1 or more Segment elements
// to describe a 2D contour. If a Slice contains content,
// there MUST be at least one Polygon to describe it.
type Polygon struct {
	StartV   uint32
	Segments []Segment
}

// Slice defines the resource object for slices.
type Slice struct {
	TopZ     float32
	Vertices []go3mf.Point2D
	Polygons []Polygon
}

// MeshResolution defines the resolutions for a slice.
type MeshResolution uint8

// Supported slice resolution.
const (
	ResolutionFull MeshResolution = iota
	ResolutionLow
)

func newMeshResolution(s string) (r MeshResolution, ok bool) {
	r, ok = map[string]MeshResolution{
		"fullres": ResolutionFull,
		"lowres":  ResolutionLow,
	}[s]
	return
}

func (c MeshResolution) String() string {
	return map[MeshResolution]string{
		ResolutionFull: "fullres",
		ResolutionLow:  "lowres",
	}[c]
}

// SliceRef reference to a slice stack.
type SliceRef struct {
	SliceStackID uint32
	Path         string
}

// SliceStack defines a slice stack resource.
// It can either contain a SliceStack or a Refs slice.
type SliceStack struct {
	ID      uint32
	BottomZ float32
	Slices  []*Slice
	Refs    []SliceRef
}

// Identify returns the unique ID of the resource.
func (s *SliceStack) Identify() uint32 {
	return s.ID
}

func GetObjectAttr(obj *go3mf.Object) *ObjectAttr {
	for _, a := range obj.AnyAttr {
		if a, ok := a.(*ObjectAttr); ok {
			return a
		}
	}
	return nil
}

// ObjectAttr defines the attributes added to Object.
type ObjectAttr struct {
	SliceStackID   uint32
	MeshResolution MeshResolution
}

const (
	attrSliceStack = "slicestack"
	attrID         = "id"
	attrZBottom    = "zbottom"
	attrSlice      = "slice"
	attrSliceRef   = "sliceref"
	attrZTop       = "ztop"
	attrVertices   = "vertices"
	attrVertex     = "vertex"
	attrPolygon    = "polygon"
	attrX          = "x"
	attrY          = "y"
	attrZ          = "z"
	attrSegment    = "segment"
	attrV1         = "v1"
	attrV2         = "v2"
	attrV3         = "v3"
	attrStartV     = "startv"
	attrSliceRefID = "slicestackid"
	attrSlicePath  = "slicepath"
	attrMeshRes    = "meshresolution"
	attrPID        = "pid"
	attrP1         = "p1"
	attrP2         = "p2"
)
