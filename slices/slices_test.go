package slices

import (
	"reflect"
	"testing"

	"github.com/qmuntal/go3mf"
	"github.com/qmuntal/go3mf/spec/encoding"
	"github.com/qmuntal/go3mf/spec/validation"
)

var _ go3mf.Asset = new(SliceStack)
var _ validation.ValidatorSpec = new(Spec)
var _ encoding.DecoderSpec = new(Spec)
var _ encoding.Marshaler = new(SliceStack)
var _ encoding.MarshalerAttr = new(ObjectAttr)

func TestSliceStack_Identify(t *testing.T) {
	tests := []struct {
		name string
		s    *SliceStack
		want uint32
	}{
		{"base", &SliceStack{ID: 1}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.s.Identify()
			if got != tt.want {
				t.Errorf("SliceStack.Identify() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeshResolution_String(t *testing.T) {
	tests := []struct {
		name string
		c    MeshResolution
	}{
		{"fullres", ResolutionFull},
		{"lowres", ResolutionLow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.name {
				t.Errorf("MeshResolution.String() = %v, want %v", got, tt.name)
			}
		})
	}
}

func Test_newMeshResolution(t *testing.T) {
	tests := []struct {
		name   string
		wantR  MeshResolution
		wantOk bool
	}{
		{"fullres", ResolutionFull, true},
		{"lowres", ResolutionLow, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR, gotOk := newMeshResolution(tt.name)
			if !reflect.DeepEqual(gotR, tt.wantR) {
				t.Errorf("newMeshResolution() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotOk != tt.wantOk {
				t.Errorf("newMeshResolution() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
