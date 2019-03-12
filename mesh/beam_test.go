package mesh

import (
	"reflect"
	"testing"
)

func Test_newbeamLattice(t *testing.T) {
	tests := []struct {
		name string
		want *beamLattice
	}{
		{"new", &beamLattice{CapMode: CapModeSphere, DefaultRadius: 1.0, MinLength: 0.0001}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newbeamLattice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newbeamLattice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beamLattice_clearBeamLattice(t *testing.T) {
	b := new(beamLattice)
	b.Beams = append(b.Beams, Beam{})
	b.BeamSets = append(b.BeamSets, BeamSet{})
	tests := []struct {
		name string
		b    *beamLattice
	}{
		{"base", b},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.clearBeamLattice()
			if len(tt.b.Beams) != 0 || len(tt.b.BeamSets) != 0 {
				t.Error("beamLattice.clearBeamLattice() have not cleared all the arrays")
			}
		})
	}
}

func Test_beamLattice_checkSanity(t *testing.T) {
	type args struct {
		nodeCount uint32
	}
	tests := []struct {
		name string
		b    *beamLattice
		args args
		want bool
	}{
		{"max", &beamLattice{maxBeamCount: 1, Beams: []Beam{{}, {}}}, args{0}, false},
		{"eq", &beamLattice{Beams: []Beam{{NodeIndices: [2]uint32{1, 1}}}}, args{0}, false},
		{"high1", &beamLattice{Beams: []Beam{{NodeIndices: [2]uint32{2, 1}}}}, args{2}, false},
		{"high2", &beamLattice{Beams: []Beam{{NodeIndices: [2]uint32{1, 2}}}}, args{2}, false},
		{"good", &beamLattice{Beams: []Beam{{NodeIndices: [2]uint32{1, 2}}}}, args{3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.checkSanity(tt.args.nodeCount); got != tt.want {
				t.Errorf("beamLattice.checkSanity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_beamLattice_merge(t *testing.T) {
	type args struct {
		newNodes []uint32
	}
	tests := []struct {
		name  string
		b     *beamLattice
		args  args
		times int
	}{
		{"err", &beamLattice{Beams: []Beam{{}}}, args{[]uint32{0, 0}}, 1},
		{"zero", new(beamLattice), args{make([]uint32, 0)}, 0},
		{"merged", new(beamLattice), args{[]uint32{0, 1}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beam := Beam{NodeIndices: [2]uint32{0, 1}, Radius: [2]float64{1.0, 2.0}, CapMode: [2]BeamCapMode{CapModeButt, CapModeHemisphere}}
			mockMesh := NewMesh()
			for i := 0; i < tt.times; i++ {
				mockMesh.Beams = append(mockMesh.Beams, beam)
			}
			tt.b.merge(&mockMesh.beamLattice, tt.args.newNodes)
			emptyBeam := Beam{}
			if len(tt.b.Beams) > 0 && tt.b.Beams[0] != emptyBeam {
				for i := 0; i < len(tt.b.Beams); i++ {
					want := beam
					want.Index = uint32(i)
					if got := tt.b.Beams[i]; got != want {
						t.Errorf("beamLattice.merge() = %v, want %v", got, want)
						return
					}
				}
			}
		})
	}
}