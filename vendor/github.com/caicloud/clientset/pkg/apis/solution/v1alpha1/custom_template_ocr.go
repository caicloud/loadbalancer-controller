package v1alpha1

// Identification specific tag's name and local
type Identification struct {
	// ID is Identification's unique identification
	ID string `json:"id"`
	// Name is target name
	Name string `json:"name"`
	// Coordinates is graph local
	Coordinates []Coordinate `json:"coordinates"`
}

type CustomTemplateOCR struct {
	// References means reference var list
	References []Identification `json:"references"`
	// Targets means target var list
	Targets []Identification `json:"targets"`
}
