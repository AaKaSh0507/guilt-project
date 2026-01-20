package domain

type UpsertPreferencesDTO struct {
	UserID        string      `validate:"required,uuid4"`
	Theme         *string     `validate:"omitempty"`
	Notifications bool        `validate:"-"`
	Metadata      interface{} `validate:"-"`
}

type GetPreferencesDTO struct {
	UserID string `validate:"required,uuid4"`
}
