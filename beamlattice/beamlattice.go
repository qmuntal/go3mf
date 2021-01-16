package beamlattice

import (
	"errors"

	"github.com/qmuntal/go3mf"
)

// Namespace is the canonical name of this extension.
const Namespace = "http://schemas.microsoft.com/3dmanufacturing/beamlattice/2017/02"

var (
	ErrLatticeObjType       = errors.New("MUST only be added to a mesh object of type model or solidsupport")
	ErrLatticeClippedNoMesh = errors.New("if clipping mode is not equal to none, a clippingmesh resource MUST be specified")
	ErrLatticeInvalidMesh   = errors.New("the clippingmesh and representationmesh MUST be a mesh object of type model and MUST NOT contain a beamlattice")
	ErrLatticeSameVertex    = errors.New("a beam MUST consist of two distinct vertex indices")
	ErrLatticeBeamR2        = errors.New("r2 MUST not be defined, if r1 is not defined")
)

type Spec struct {
	LocalName string
	m         *go3mf.Model
}

func (e *Spec) SetModel(m *go3mf.Model) { e.m = m }
func (e Spec) Namespace() string        { return Namespace }
func (e Spec) Required() bool           { return true }
func (e *Spec) SetRequired(r bool)      {}
func (e *Spec) SetLocal(l string)       { e.LocalName = l }

func (e Spec) Local() string {
	if e.LocalName != "" {
		return e.LocalName
	}
	return "b"
}

// ClipMode defines the clipping modes for the beam lattices.
type ClipMode uint8

// Supported clip modes.
const (
	ClipNone ClipMode = iota
	ClipInside
	ClipOutside
)

func newClipMode(s string) (c ClipMode, ok bool) {
	c, ok = map[string]ClipMode{
		"none":    ClipNone,
		"inside":  ClipInside,
		"outside": ClipOutside,
	}[s]
	return
}

func (c ClipMode) String() string {
	return map[ClipMode]string{
		ClipNone:    "none",
		ClipInside:  "inside",
		ClipOutside: "outside",
	}[c]
}

// A CapMode is an enumerable for the different capping modes.
type CapMode uint8

// Supported cap modes.
const (
	CapModeSphere CapMode = iota
	CapModeHemisphere
	CapModeButt
)

func newCapMode(s string) (t CapMode, ok bool) {
	t, ok = map[string]CapMode{
		"sphere":     CapModeSphere,
		"hemisphere": CapModeHemisphere,
		"butt":       CapModeButt,
	}[s]
	return
}

func (b CapMode) String() string {
	return map[CapMode]string{
		CapModeSphere:     "sphere",
		CapModeHemisphere: "hemisphere",
		CapModeButt:       "butt",
	}[b]
}

// BeamLattice defines the Model Mesh BeamLattice Attributes class and is part of the BeamLattice extension to 3MF.
type BeamLattice struct {
	ClipMode             ClipMode
	ClippingMeshID       uint32
	RepresentationMeshID uint32
	Beams                []Beam
	BeamSets             []BeamSet
	MinLength, Radius    float32
	CapMode              CapMode
}

func GetBeamLattice(mesh *go3mf.Mesh) *BeamLattice {
	for _, a := range mesh.Any {
		if a, ok := a.(*BeamLattice); ok {
			return a
		}
	}
	return nil
}

// BeamSet defines a set of beams.
type BeamSet struct {
	Refs       []uint32
	Name       string
	Identifier string
}

// Beam defines a single beam.
type Beam struct {
	Indices [2]uint32  // Indices of the two nodes that defines the beam.
	Radius  [2]float32 // Radius of both ends of the beam.
	CapMode [2]CapMode // Capping mode.
}

const (
	attrBeamLattice        = "beamlattice"
	attrRadius             = "radius"
	attrMinLength          = "minlength"
	attrPrecision          = "precision"
	attrClippingMode       = "clippingmode"
	attrClipping           = "clipping"
	attrClippingMesh       = "clippingmesh"
	attrRepresentationMesh = "representationmesh"
	attrCap                = "cap"
	attrBeams              = "beams"
	attrBeam               = "beam"
	attrBeamSets           = "beamsets"
	attrBeamSet            = "beamset"
	attrR1                 = "r1"
	attrR2                 = "r2"
	attrCap1               = "cap1"
	attrCap2               = "cap2"
	attrV1                 = "v1"
	attrV2                 = "v2"
	attrV3                 = "v3"
	attrName               = "name"
	attrIdentifier         = "identifier"
	attrRef                = "ref"
	attrIndex              = "index"
)
