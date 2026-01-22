package queue

type EntryMLJob struct {
	EntryID   string
	UserID    string
	Text      string
	Persona   string
	Intensity int
	History   []string
}
