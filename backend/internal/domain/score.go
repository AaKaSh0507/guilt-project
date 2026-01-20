package domain

type CreateScoreDTO struct {
	SessionID string      `validate:"required,uuid4"`
	Score     int32       `validate:"gte=0,lte=100"`
	Meta      interface{} `validate:"-"`
}

type GetScoreDTO struct {
	SessionID string `validate:"required,uuid4"`
}
