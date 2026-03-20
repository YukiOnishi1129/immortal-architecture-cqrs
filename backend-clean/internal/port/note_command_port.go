package port

import (
	"context"

	"immortal-architecture-cqrs/backend/internal/domain/note"
)

// NoteCommandInputPort defines command (write) use case inputs for notes.
type NoteCommandInputPort interface {
	Create(ctx context.Context, input NoteCreateInput) error
	Update(ctx context.Context, input NoteUpdateInput) error
	ChangeStatus(ctx context.Context, input NoteStatusChangeInput) error
	Delete(ctx context.Context, id, ownerID string) error
}

// NoteCommandOutputPort defines command (write) presenters for notes.
type NoteCommandOutputPort interface {
	PresentNote(ctx context.Context, note *note.WithMeta) error
	PresentNoteDeleted(ctx context.Context) error
}

// NoteCommandRepository abstracts command (write) persistence for notes.
type NoteCommandRepository interface {
	Get(ctx context.Context, id string) (*note.WithMeta, error)
	Create(ctx context.Context, n note.Note) (*note.Note, error)
	Update(ctx context.Context, n note.Note) (*note.Note, error)
	UpdateStatus(ctx context.Context, id string, status note.NoteStatus) (*note.Note, error)
	Delete(ctx context.Context, id string) error
	ReplaceSections(ctx context.Context, noteID string, sections []note.Section) error
}
