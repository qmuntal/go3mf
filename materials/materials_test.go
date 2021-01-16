package materials

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
)

var _ go3mf.Asset = new(Texture2D)
var _ go3mf.Asset = new(Texture2DGroup)
var _ go3mf.Asset = new(CompositeMaterials)
var _ go3mf.Asset = new(MultiProperties)
var _ go3mf.Asset = new(ColorGroup)
var _ encoding.Decoder = new(Spec)
var _ encoding.Marshaler = new(Texture2D)
var _ encoding.Marshaler = new(Texture2DGroup)
var _ encoding.Marshaler = new(CompositeMaterials)
var _ encoding.Marshaler = new(ColorGroup)
var _ encoding.Marshaler = new(MultiProperties)
var _ spec.Spec = new(Spec)
var _ spec.Validator = new(Spec)
var _ spec.PropertyGroup = new(ColorGroup)
var _ spec.PropertyGroup = new(Texture2DGroup)
var _ spec.PropertyGroup = new(CompositeMaterials)
var _ spec.PropertyGroup = new(MultiProperties)

func TestTexture2D_Identify(t *testing.T) {
	tests := []struct {
		name string
		t    *Texture2D
		want uint32
	}{
		{"base", &Texture2D{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.Identify()
			if got != tt.want {
				t.Errorf("Texture2D.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextureCoord_U(t *testing.T) {
	tests := []struct {
		name string
		t    TextureCoord
		want float32
	}{
		{"base", TextureCoord{1, 2}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.U(); got != tt.want {
				t.Errorf("TextureCoord.U() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextureCoord_V(t *testing.T) {
	tests := []struct {
		name string
		t    TextureCoord
		want float32
	}{
		{"base", TextureCoord{1, 2}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.V(); got != tt.want {
				t.Errorf("TextureCoord.V() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTexture2DGroup_Identify(t *testing.T) {
	tests := []struct {
		name string
		t    *Texture2DGroup
		want uint32
	}{
		{"base", &Texture2DGroup{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.t.Identify()
			if got != tt.want {
				t.Errorf("Texture2DGroup.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColorGroup_Identify(t *testing.T) {
	tests := []struct {
		name string
		c    *ColorGroup
		want uint32
	}{
		{"base", &ColorGroup{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Identify()
			if got != tt.want {
				t.Errorf("ColorGroup.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositeMaterials_Identify(t *testing.T) {
	tests := []struct {
		name string
		c    *CompositeMaterials
		want uint32
	}{
		{"base", &CompositeMaterials{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Identify()
			if got != tt.want {
				t.Errorf("CompositeMaterials.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTexture2DType_String(t *testing.T) {
	tests := []struct {
		name string
		t    Texture2DType
	}{
		{"image/png", TextureTypePNG},
		{"image/jpeg", TextureTypeJPEG},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.name {
				t.Errorf("Texture2DType.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestBlendMethod_String(t *testing.T) {
	tests := []struct {
		name string
		b    BlendMethod
	}{
		{"mix", BlendMix},
		{"multiply", BlendMultiply},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.name {
				t.Errorf("BlendMethod.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestTileStyle_String(t *testing.T) {
	tests := []struct {
		name string
		t    TileStyle
	}{
		{"wrap", TileWrap},
		{"mirror", TileMirror},
		{"clamp", TileClamp},
		{"none", TileNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.name {
				t.Errorf("TileStyle.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestTextureFilter_String(t *testing.T) {
	tests := []struct {
		name string
		t    TextureFilter
	}{
		{"auto", TextureFilterAuto},
		{"linear", TextureFilterLinear},
		{"nearest", TextureFilterNearest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.name {
				t.Errorf("TextureFilter.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func TestMultiProperties_Identify(t *testing.T) {
	tests := []struct {
		name string
		c    *MultiProperties
		want uint32
	}{
		{"base", &MultiProperties{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Identify()
			if got != tt.want {
				t.Errorf("MultiProperties.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newBlendMethod(t *testing.T) {
	tests := []struct {
		name   string
		wantB  BlendMethod
		wantOk bool
	}{
		{"mix", BlendMix, true},
		{"multiply", BlendMultiply, true},
		{"empty", BlendMix, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotB, gotOk := newBlendMethod(tt.name)
			if !reflect.DeepEqual(gotB, tt.wantB) {
				t.Errorf("newBlendMethod() gotB = %v, want %v", gotB, tt.wantB)
			}
			if gotOk != tt.wantOk {
				t.Errorf("newBlendMethod() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_newTextureFilter(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		want   TextureFilter
		wantOk bool
	}{
		{"auto", TextureFilterAuto, true},
		{"linear", TextureFilterLinear, true},
		{"nearest", TextureFilterNearest, true},
		{"empty", TextureFilterAuto, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := newTextureFilter(tt.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTextureFilter() got = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("newTextureFilter() got1 = %v, want %v", got1, tt.want)
			}
		})
	}
}

func Test_newTileStyle(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name  string
		want  TileStyle
		want1 bool
	}{
		{"wrap", TileWrap, true},
		{"mirror", TileMirror, true},
		{"clamp", TileClamp, true},
		{"none", TileNone, true},
		{"empty", TileWrap, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := newTileStyle(tt.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTileStyle() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newTileStyle() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_newTexture2DType(t *testing.T) {
	tests := []struct {
		name  string
		want  Texture2DType
		want1 bool
	}{
		{"image/png", TextureTypePNG, true},
		{"image/jpeg", TextureTypeJPEG, true},
		{"", Texture2DType(0), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := newTexture2DType(tt.name)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTexture2DType() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newTexture2DType() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestColorGroup_Len(t *testing.T) {
	tests := []struct {
		name string
		r    *ColorGroup
		want int
	}{
		{"empty", new(ColorGroup), 0},
		{"base", &ColorGroup{Colors: make([]color.RGBA, 3)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("ColorGroup.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositeMaterials_Len(t *testing.T) {
	tests := []struct {
		name string
		r    *CompositeMaterials
		want int
	}{
		{"empty", new(CompositeMaterials), 0},
		{"base", &CompositeMaterials{Composites: make([]Composite, 3)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("CompositeMaterials.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiProperties_Len(t *testing.T) {
	tests := []struct {
		name string
		r    *MultiProperties
		want int
	}{
		{"empty", new(MultiProperties), 0},
		{"base", &MultiProperties{Multis: make([]Multi, 3)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("MultiProperties.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTexture2DGroup_Len(t *testing.T) {
	tests := []struct {
		name string
		r    *Texture2DGroup
		want int
	}{
		{"empty", new(Texture2DGroup), 0},
		{"base", &Texture2DGroup{Coords: make([]TextureCoord, 3)}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Len(); got != tt.want {
				t.Errorf("Texture2DGroup.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}
