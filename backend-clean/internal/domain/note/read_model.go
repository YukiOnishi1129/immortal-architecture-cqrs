package note

import "time"

// NoteReadModel is the denormalized read model for notes.
// It contains all data needed to display a note without JOINs.
type NoteReadModel struct {
	ID             string
	Title          string
	Status         NoteStatus
	TemplateID     string
	TemplateName   string
	OwnerID        string
	OwnerFirstName string
	OwnerLastName  string
	OwnerThumbnail *string
	Sections       []SectionReadModel
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SectionReadModel is the denormalized read model for sections.
type SectionReadModel struct {
	ID         string
	FieldID    string
	FieldLabel string
	FieldOrder int
	IsRequired bool
	Content    string
}
