package usecase

import (
	"context"

	"immortal-architecture-cqrs/backend/internal/domain/note"
	"immortal-architecture-cqrs/backend/internal/port"
)

// NoteQueryInteractor handles note query (read) use cases.
type NoteQueryInteractor struct {
	queryRepo port.NoteQueryRepository
	output    port.NoteQueryOutputPort
}

var _ port.NoteQueryInputPort = (*NoteQueryInteractor)(nil)

// NewNoteQueryInteractor creates NoteQueryInteractor.
func NewNoteQueryInteractor(queryRepo port.NoteQueryRepository, output port.NoteQueryOutputPort) *NoteQueryInteractor {
	return &NoteQueryInteractor{
		queryRepo: queryRepo,
		output:    output,
	}
}

// List returns notes from the read model.
func (u *NoteQueryInteractor) List(ctx context.Context, filters note.Filters) error {
	notes, err := u.queryRepo.List(ctx, filters)
	if err != nil {
		return err
	}
	return u.output.PresentNoteList(ctx, notes)
}

// Get returns a single note from the read model.
func (u *NoteQueryInteractor) Get(ctx context.Context, id string) error {
	n, err := u.queryRepo.Get(ctx, id)
	if err != nil {
		return err
	}
	return u.output.PresentNote(ctx, n)
}
