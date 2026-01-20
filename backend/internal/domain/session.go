package domain

type CreateSessionDTO struct {
	UserID string `validate:"required,uuid4"`
	Notes  *string
}

type ListSessionsDTO struct {
	UserID string `validate:"required,uuid4"`
	Limit  int32  `validate:"gte=0"`
	Offset int32  `validate:"gte=0"`
}

type EndSessionDTO struct {
	SessionID string `validate:"required,uuid4"`
}
