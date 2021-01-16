package materials

import (
	"errors"
	"image/color"
)

const (
	// Namespace is the canonical name of this extension.
	Namespace = "http://schemas.microsoft.com/3dmanufacturing/material/2015/02"
	// RelTypeTexture3D is the canonical 3D texture relationship type.
	RelTypeTexture3D = "http://schemas.microsoft.com/3dmanufacturing/2013/01/3dtexture"
)

var (
	ErrMultiBlend         = errors.New("there MUST NOT be more blendmethods than layers – 1")
	ErrMaterialMulti      = errors.New("a material, if included, MUST be positioned as the first layer")
	ErrMultiRefMulti      = errors.New("the pids list MUST NOT contain any references to a multiproperties")
	ErrMultiColors        = errors.New("the pids list MUST NOT contain more than one reference to a colorgroup")
	ErrTextureReference   = errors.New("MUST reference to a texture resource")
	ErrCompositeBase      = errors.New("MUST reference to a basematerials group")
	ErrMissingTexturePart = errors.New("texture part MUST be added as an attachment")
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
	return "m"
}

// Texture2DType defines the allowed texture 2D types.
type Texture2DType uint8

// Supported texture types.
const (
	TextureTypePNG Texture2DType = iota + 1
	TextureTypeJPEG
)

func (t Texture2DType) String() string {
	return map[Texture2DType]string{
		TextureTypePNG:  "image/png",
		TextureTypeJPEG: "image/jpeg",
	}[t]
}

// TileStyle defines the allowed tile styles.
type TileStyle uint8

// Supported tile style.
const (
	TileWrap TileStyle = iota
	TileMirror
	TileClamp
	TileNone
)

func (t TileStyle) String() string {
	return map[TileStyle]string{
		TileWrap:   "wrap",
		TileMirror: "mirror",
		TileClamp:  "clamp",
		TileNone:   "none",
	}[t]
}

// TextureFilter defines the allowed texture filters.
type TextureFilter uint8

// Supported texture filters.
const (
	TextureFilterAuto TextureFilter = iota
	TextureFilterLinear
	TextureFilterNearest
)

func (t TextureFilter) String() string {
	return map[TextureFilter]string{
		TextureFilterAuto:    "auto",
		TextureFilterLinear:  "linear",
		TextureFilterNearest: "nearest",
	}[t]
}

// BlendMethod defines the equation to use when blending a layer with the previous layer.
type BlendMethod uint8

// Supported blend methods.
const (
	BlendMix BlendMethod = iota
	BlendMultiply
)

func (b BlendMethod) String() string {
	return map[BlendMethod]string{
		BlendMix:      "mix",
		BlendMultiply: "multiply",
	}[b]
}

// Texture2D defines the Model Texture 2D.
type Texture2D struct {
	ID          uint32
	Path        string
	ContentType Texture2DType
	TileStyleU  TileStyle
	TileStyleV  TileStyle
	Filter      TextureFilter
}

// Identify returns the unique ID of the resource.
func (t *Texture2D) Identify() uint32 {
	return t.ID
}

// TextureCoord map a vertex of a triangle to a position in image space (U, V coordinates)
type TextureCoord [2]float32

// U returns the first coordinate.
func (t TextureCoord) U() float32 {
	return t[0]
}

// V returns the second coordinate.
func (t TextureCoord) V() float32 {
	return t[1]
}

// Texture2DGroup acts as a container for texture coordinate properties.
type Texture2DGroup struct {
	ID        uint32
	TextureID uint32
	Coords    []TextureCoord
}

// Len returns the materials count.
func (r *Texture2DGroup) Len() int {
	return len(r.Coords)
}

// Identify returns the unique ID of the resource.
func (r *Texture2DGroup) Identify() uint32 {
	return r.ID
}

// ColorGroup acts as a container for color properties.
type ColorGroup struct {
	ID     uint32
	Colors []color.RGBA
}

// Len returns the materials count.
func (r *ColorGroup) Len() int {
	return len(r.Colors)
}

// Identify returns the unique ID of the resource.
func (c *ColorGroup) Identify() uint32 {
	return c.ID
}

// A Composite specifies the proportion of the overall mixture for each material.
type Composite struct {
	Values []float32
}

// CompositeMaterials defines materials derived by mixing 2 or more base materials in defined ratios.
type CompositeMaterials struct {
	ID         uint32
	MaterialID uint32
	Indices    []uint32
	Composites []Composite
}

// Len returns the materials count.
func (r *CompositeMaterials) Len() int {
	return len(r.Composites)
}

// Identify returns the unique ID of the resource.
func (c *CompositeMaterials) Identify() uint32 {
	return c.ID
}

// The Multi element combines the constituent materials and properties.
type Multi struct {
	PIndices []uint32
}

// A MultiProperties element acts as a container for Multi
// elements which are indexable groups of property indices.
type MultiProperties struct {
	ID           uint32
	PIDs         []uint32
	BlendMethods []BlendMethod
	Multis       []Multi
}

// Len returns the materials count.
func (r *MultiProperties) Len() int {
	return len(r.Multis)
}

// Identify returns the unique ID of the resource.
func (c *MultiProperties) Identify() uint32 {
	return c.ID
}

func newTexture2DType(s string) (t Texture2DType, ok bool) {
	t, ok = map[string]Texture2DType{
		"image/png":  TextureTypePNG,
		"image/jpeg": TextureTypeJPEG,
	}[s]
	return
}

func newTextureFilter(s string) (t TextureFilter, ok bool) {
	t, ok = map[string]TextureFilter{
		"auto":    TextureFilterAuto,
		"linear":  TextureFilterLinear,
		"nearest": TextureFilterNearest,
	}[s]
	return
}

func newTileStyle(s string) (t TileStyle, ok bool) {
	t, ok = map[string]TileStyle{
		"wrap":   TileWrap,
		"mirror": TileMirror,
		"clamp":  TileClamp,
		"none":   TileNone,
	}[s]
	return
}

func newBlendMethod(s string) (b BlendMethod, ok bool) {
	b, ok = map[string]BlendMethod{
		"mix":      BlendMix,
		"multiply": BlendMultiply,
	}[s]
	return
}

const (
	attrPath               = "path"
	attrID                 = "id"
	attrColorGroup         = "colorgroup"
	attrColor              = "color"
	attrTexture2DGroup     = "texture2dgroup"
	attrTex2DCoord         = "tex2coord"
	attrTexID              = "texid"
	attrU                  = "u"
	attrV                  = "v"
	attrContentType        = "contenttype"
	attrTileStyleU         = "tilestyleu"
	attrTileStyleV         = "tilestylev"
	attrFilter             = "filter"
	attrTexture2D          = "texture2d"
	attrComposite          = "composite"
	attrCompositematerials = "compositematerials"
	attrValues             = "values"
	attrMatID              = "matid"
	attrMatIndices         = "matindices"
	attrMultiProps         = "multiproperties"
	attrMulti              = "multi"
	attrPIndices           = "pindices"
	attrPIDs               = "pids"
	attrBlendMethods       = "blendmethods"
)
