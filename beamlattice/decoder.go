package beamlattice

import (
	"encoding/xml"
	"strconv"

	"github.com/qmuntal/go3mf"
	specerr "github.com/qmuntal/go3mf/errors"
	"github.com/qmuntal/go3mf/spec/encoding"
)

func (e Spec) NewElementDecoder(ctx encoding.ElementDecoderContext) encoding.ElementDecoder {
	if ctx.Name.Local == attrBeamLattice {
		return &beamLatticeDecoder{mesh: ctx.ParentElement.(*go3mf.Mesh), ew: ctx.ErrorWrapper}
	}
	return nil
}

func (e Spec) DecodeAttribute(_ interface{}, _ encoding.XMLAttr) error { return nil }

type beamLatticeDecoder struct {
	baseDecoder
	mesh *go3mf.Mesh
	ew   encoding.ErrorWrapper
}

func (d *beamLatticeDecoder) Start(attrs []encoding.XMLAttr) error {
	var errs error
	beamLattice := new(BeamLattice)
	d.mesh.Any = append(d.mesh.Any, beamLattice)
	for _, a := range attrs {
		if a.Name.Space != "" {
			continue
		}
		switch a.Name.Local {
		case attrRadius:
			val, err := strconv.ParseFloat(string(a.Value), 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			beamLattice.Radius = float32(val)
		case attrMinLength, attrPrecision: // lib3mf legacy
			val, err := strconv.ParseFloat(string(a.Value), 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			beamLattice.MinLength = float32(val)
		case attrClippingMode, attrClipping: // lib3mf legacy
			var ok bool
			beamLattice.ClipMode, ok = newClipMode(string(a.Value))
			if !ok {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
		case attrClippingMesh:
			val, err := strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
			beamLattice.ClippingMeshID = uint32(val)
		case attrRepresentationMesh:
			val, err := strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
			beamLattice.RepresentationMeshID = uint32(val)
		case attrCap:
			var ok bool
			beamLattice.CapMode, ok = newCapMode(string(a.Value))
			if !ok {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
		}
	}
	if errs != nil {
		return specerr.Wrap(errs, GetBeamLattice(d.mesh))
	}
	return nil
}

func (d *beamLatticeDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.Wrap(err, GetBeamLattice(d.mesh)))
}

func (d *beamLatticeDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace {
		if name.Local == attrBeams {
			child = &beamsDecoder{mesh: d.mesh, ew: d}
		} else if name.Local == attrBeamSets {
			child = &beamSetsDecoder{mesh: d.mesh, ew: d}
		}
	}
	return
}

type beamsDecoder struct {
	baseDecoder
	mesh        *go3mf.Mesh
	beamDecoder beamDecoder
	ew          encoding.ErrorWrapper
}

func (d *beamsDecoder) Start(_ []encoding.XMLAttr) error {
	d.beamDecoder.mesh = d.mesh
	return nil
}

func (d *beamsDecoder) Wrap(err error) error {
	return d.ew.Wrap(err)
}

func (d *beamsDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrBeam {
		child = &d.beamDecoder
	}
	return
}

type beamDecoder struct {
	baseDecoder
	mesh *go3mf.Mesh
}

func (d *beamDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		beam             Beam
		hasCap1, hasCap2 bool
		errs             error
	)
	beamLattice := GetBeamLattice(d.mesh)
	for _, a := range attrs {
		if a.Name.Space != "" {
			continue
		}
		switch a.Name.Local {
		case attrV1:
			val, err := strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			beam.Indices[0] = uint32(val)
		case attrV2:
			val, err := strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			beam.Indices[1] = uint32(val)
		case attrR1:
			val, err := strconv.ParseFloat(string(a.Value), 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
			beam.Radius[0] = float32(val)
		case attrR2:
			val, err := strconv.ParseFloat(string(a.Value), 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, false))
			}
			beam.Radius[1] = float32(val)
		case attrCap1:
			var ok bool
			beam.CapMode[0], ok = newCapMode(string(a.Value))
			if ok {
				hasCap1 = true
			}
		case attrCap2:
			var ok bool
			beam.CapMode[1], ok = newCapMode(string(a.Value))
			if ok {
				hasCap2 = true
			}
		}
	}
	if beam.Radius[0] == 0 {
		beam.Radius[0] = beamLattice.Radius
	}
	if beam.Radius[1] == 0 {
		beam.Radius[1] = beam.Radius[0]
	}
	if !hasCap1 {
		beam.CapMode[0] = beamLattice.CapMode
	}
	if !hasCap2 {
		beam.CapMode[1] = beamLattice.CapMode
	}
	beamLattice.Beams = append(beamLattice.Beams, beam)
	if errs != nil {
		return specerr.WrapIndex(errs, beam, len(beamLattice.Beams)-1)
	}
	return nil
}

type beamSetsDecoder struct {
	baseDecoder
	mesh *go3mf.Mesh
	ew   encoding.ErrorWrapper
}

func (d *beamSetsDecoder) Wrap(err error) error {
	return d.ew.Wrap(err)
}

func (d *beamSetsDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrBeamSet {
		child = &beamSetDecoder{mesh: d.mesh, ew: d}
	}
	return
}

type beamSetDecoder struct {
	baseDecoder
	mesh           *go3mf.Mesh
	beamSet        BeamSet
	beamRefDecoder beamRefDecoder
	ew             encoding.ErrorWrapper
}

func (d *beamSetDecoder) End() {
	beamLattice := GetBeamLattice(d.mesh)
	beamLattice.BeamSets = append(beamLattice.BeamSets, d.beamSet)
}

func (d *beamSetDecoder) Start(attrs []encoding.XMLAttr) error {
	d.beamRefDecoder.beamSet = &d.beamSet
	for _, a := range attrs {
		if a.Name.Space != "" {
			continue
		}
		switch a.Name.Local {
		case attrName:
			d.beamSet.Name = string(a.Value)
		case attrIdentifier:
			d.beamSet.Identifier = string(a.Value)
		}
	}
	return nil
}

func (d *beamSetDecoder) Wrap(err error) error {
	return d.ew.Wrap(specerr.WrapIndex(err, &d.beamSet, len(GetBeamLattice(d.mesh).BeamSets)))
}

func (d *beamSetDecoder) Child(name xml.Name) (child encoding.ElementDecoder) {
	if name.Space == Namespace && name.Local == attrRef {
		child = &d.beamRefDecoder
	}
	return
}

type beamRefDecoder struct {
	baseDecoder
	beamSet *BeamSet
}

func (d *beamRefDecoder) Start(attrs []encoding.XMLAttr) error {
	var (
		val  uint64
		errs error
	)
	for _, a := range attrs {
		if a.Name.Space == "" && a.Name.Local == attrIndex {
			var err error
			val, err = strconv.ParseUint(string(a.Value), 10, 32)
			if err != nil {
				errs = specerr.Append(errs, specerr.NewParseAttrError(a.Name.Local, true))
			}
			break
		}
	}
	d.beamSet.Refs = append(d.beamSet.Refs, uint32(val))
	if errs != nil {
		return specerr.WrapIndex(errs, uint32(0), len(d.beamSet.Refs)-1)
	}
	return nil
}

type baseDecoder struct {
}

func (d *baseDecoder) Start([]encoding.XMLAttr) error { return nil }
func (d *baseDecoder) End()                        {}
