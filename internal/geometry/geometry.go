package geometry

import (
	"github.com/ungerik/go3d/vec3"
)

// Vec3I represents a 3D vector typed as int32
type Vec3I struct {
	X int32 // X coordinate
	Y int32 // Y coordinate
	Z int32 // Z coordinate
}

// PairMatch defines an interface which is able to identify duplicate pairs of numbers in a given data set.
type PairMatch interface {
	// AddMatch adds a match to the set.
	// If the match exists it is overriden.
	AddMatch(data1, data2, param int32)
	// CheckMatch check if a match is in the set.
	CheckMatch(data1, data2 int32) (val int32, ok bool)
	// DeleteMatch deletes a match from the set.
	// If match doesn't exist it bevavhe as a no-op
	DeleteMatch(data1, data2 int32)
}

// VectorDic defines an interface which is able to identify vectors by their position
type VectorDic interface {
	Units() float32
	SetUnits(units float32)
	AddVector(vec vec3.T, value uint32)
	FindVector(vec vec3.T) uint32
	RemoveVector(vec vec3.T)
}