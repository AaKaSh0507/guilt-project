package ml

type Persona int

const (
	PersonaNeutral Persona = iota
	PersonaRoast
	PersonaCoach
	PersonaChill
)

type HybridInput struct {
	Text      string
	UserID    string
	Intensity int
	Persona   Persona
	History   []string
}

type HybridOutput struct {
	GuiltScore  float64
	RoastText   string
	Tags        []string
	SafetyFlags []string
}
