//go:generate mockgen -destination types_mock_test.go -package meshinfo -self_package github.com/qmuntal/go3mf/internal/meshinfo github.com/qmuntal/go3mf/internal/meshinfo Invalidator,Container,MeshInfo

package meshinfo

// Color represents a RGB color.
type Color = uint32

// FaceData defines a generic information of a face. Implementations could by NodeColor or TextureCoords.
type FaceData = interface{}

// InformationType is an enumerator that identifies different information types.
type InformationType int

const (
	// InfoAbstract defines abstract information.
	InfoAbstract InformationType = iota
	// InfoBaseMaterials defines base materials information.
	InfoBaseMaterials
	// InfoNodeColors defines node colors information.
	InfoNodeColors
	// InfoTextureCoords defines texture coordinates information.
	InfoTextureCoords
	infoLastType
)

// Invalidator is used to invalidate a specific data.
type Invalidator interface {
	// Invalidate sets the data with its default values.
	Invalidate(data FaceData)
}

// Repository defines an interface for interacting with a mesh information repository.
type Repository interface {
	// AddFaceData adds data to the last added face and returns the pointer to the data of the added face.
	// The parameter newFaceCount should indicate the faces information stored in the container, including the new one.
	// If the count is not equal to the one returned by GetCurrentFaceCount an error will be returned.
	AddFaceData(newFaceCount uint32) (val FaceData, err error)
	// GetFaceData returns the data of the face with the target index.
	GetFaceData(faceIndex uint32) (val FaceData, err error)
	// GetCurrentFaceCount returns the number of faces information stored in the container.
	GetCurrentFaceCount() uint32
	// Clear removes all the information stored in the container.
	Clear()
}

// Container provides a repository for holding information state of a complete mesh structure.
// It is intended to be used as a low level repository of mesh information,
// such a thin wrapper around an in memory map or a disk serializer.
type Container interface {
	Repository
	// Clone creates a copy of the container with all the faces invalidated.
	Clone(currentFaceCount uint32) Container
}

// MeshInfo defines the Mesh Information Class.
// This is a base struct for handling all the mesh-related linear information (like face colors, textures, etc...).
type MeshInfo interface {
	Invalidator
	Repository
	// GetType returns the type of information stored in this instance.
	GetType() InformationType
	// FaceHasData checks if the specific face has any associated data.
	FaceHasData(faceIndex uint32) bool
	// Clone creates a deep copy of this instance.
	Clone(currentFaceCount uint32) MeshInfo
	// resetFaceInformation clears the data of an specific face.
	resetFaceInformation(faceIndex uint32)
	// cloneFaceInfosFrom clones the data from another face.
	cloneFaceInfosFrom(faceIndex uint32, otherInfo MeshInfo, otherFaceIndex uint32)
	//permuteNodeInformation swap the data of the target mesh.
	permuteNodeInformation(faceIndex, nodeIndex1, nodeIndex2, nodeIndex3 uint32)
	// mergeInformationFrom merges the information of the input mesh with the current information.
	mergeInformationFrom(info MeshInfo)
	// setInternalID sets an ID for the whole mesh information.
	setInternalID(internalID uint64)
	// getInternalId gets the internal ID of the mesh information.
	getInternalID() uint64
}

// MeshInfoFactory allows a dynamic creation of different Structs of Information.
type MeshInfoFactory interface {
	// Create creates a new MeshInfo of the desired type.
	Create(infoType InformationType, currentFaceCount uint32) (MeshInfo, error)
}

// Handler allows to include different kinds of information in one mesh (like Textures AND colors)
type Handler interface {
	// AddInformation adds a new type of information to the handler.
	AddInformation(info MeshInfo) error
	// AddFace adds a new face to the handler.
	AddFace(newFaceCount uint32) error
	// GetInformationByType retrieves the information of the desried type.
	GetInformationByType(infoType InformationType) MeshInfo
	// GetInformationCount returns the number of informations added to the handler.
	GetInformationCount() uint32
	// AddInfoFromTable adds the information of the target handler.
	AddInfoFromTable(otherHandler Handler, currentFaceCount uint32) error
	// CloneFaceInfosFrom clones the data from another face.
	CloneFaceInfosFrom(faceIndex uint32, otherHandler Handler, otherFaceIndex uint32)
	// ResetFaceInformation clears the data of an specific face.
	ResetFaceInformation(faceIndex uint32)
	// RemoveInformation removes the information of the target type.
	RemoveInformation(infoType InformationType)
	// PermuteNodeInformation swap the data of the target mesh.
	PermuteNodeInformation(faceIndex, nodeIndex1, nodeIndex2, nodeIndex3 uint32)
}