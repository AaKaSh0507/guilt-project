package domain

type CreateEntryDTO struct {
	SessionID string `validate:"required,uuid4"`
	Text      string `validate:"required"`
	Level     int32  `validate:"gte=0,lte=10"`
}

type ListEntriesDTO struct {
	SessionID string `validate:"required,uuid4"`
}
