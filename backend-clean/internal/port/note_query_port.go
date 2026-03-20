package port

import (
	"context"

	"immortal-architecture-cqrs/backend/internal/domain/note"
)

// NoteQueryInputPort defines query (read) use case inputs for notes.
type NoteQueryInputPort interface {
	List(ctx context.Context, filters note.Filters) error
	Get(ctx context.Context, id string) error
}

// NoteQueryOutputPort defines query (read) presenters for notes.
type NoteQueryOutputPort interface {
	PresentNoteList(ctx context.Context, notes []note.NoteReadModel) error
	PresentNote(ctx context.Context, note *note.NoteReadModel) error
}

// NoteQueryRepository abstracts query (read) persistence for notes.
// It reads from the denormalized read model table and also handles
// synchronization (Upsert/Delete) called from the command side.
type NoteQueryRepository interface {
	List(ctx context.Context, filters note.Filters) ([]note.NoteReadModel, error)
	Get(ctx context.Context, id string) (*note.NoteReadModel, error)
	Upsert(ctx context.Context, model note.NoteReadModel) error
	Delete(ctx context.Context, id string) error
}
