package beamlattice

import (
	"reflect"
	"testing"

	"github.com/qmuntal/go3mf/spec"
	"github.com/qmuntal/go3mf/spec/encoding"
	"github.com/qmuntal/go3mf/spec/validation"
)

var _ encoding.DecoderSpec = new(Spec)
var _ spec.Spec = new(Spec)
var _ validation.ValidatorSpec = new(Spec)
var _ encoding.Marshaler = new(BeamLattice)

func TestCapMode_String(t *testing.T) {
	tests := []struct {
		name string
		b    CapMode
	}{
		{"sphere", CapModeSphere},
		{"hemisphere", CapModeHemisphere},
		{"butt", CapModeButt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.String(); got != tt.name {
				t.Errorf("CapMode.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func Test_newCapMode(t *testing.T) {
	tests := []struct {
		name   string
		wantT  CapMode
		wantOk bool
	}{
		{"sphere", CapModeSphere, true},
		{"hemisphere", CapModeHemisphere, true},
		{"butt", CapModeButt, true},
		{"empty", CapModeSphere, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotT, gotOk := newCapMode(tt.name)
			if !reflect.DeepEqual(gotT, tt.wantT) {
				t.Errorf("newCapMode() gotT = %v, want %v", gotT, tt.wantT)
			}
			if gotOk != tt.wantOk {
				t.Errorf("newCapMode() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func Test_newClipMode(t *testing.T) {
	tests := []struct {
		name   string
		wantC  ClipMode
		wantOk bool
	}{
		{"none", ClipNone, true},
		{"inside", ClipInside, true},
		{"outside", ClipOutside, true},
		{"empty", ClipNone, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotC, gotOk := newClipMode(tt.name)
			if !reflect.DeepEqual(gotC, tt.wantC) {
				t.Errorf("newClipMode() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotOk != tt.wantOk {
				t.Errorf("newClipMode() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestClipMode_String(t *testing.T) {
	tests := []struct {
		name string
		c    ClipMode
	}{
		{"none", ClipNone},
		{"inside", ClipInside},
		{"outside", ClipOutside},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.name {
				t.Errorf("ClipMode.String() = %v, want %v", got, tt.name)
			}
		})
	}
}
