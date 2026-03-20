// Package factory provides constructors for driver-level wiring.
package factory

import (
	"immortal-architecture-cqrs/backend/internal/port"
	"immortal-architecture-cqrs/backend/internal/usecase"
)

// NewAccountInputFactory returns a factory for AccountInteractor.
func NewAccountInputFactory() func(repo port.AccountRepository, output port.AccountOutputPort) port.AccountInputPort {
	return func(repo port.AccountRepository, output port.AccountOutputPort) port.AccountInputPort {
		return usecase.NewAccountInteractor(repo, output)
	}
}

// NewTemplateInputFactory returns a factory for TemplateInteractor.
func NewTemplateInputFactory() func(repo port.TemplateRepository, tx port.TxManager, output port.TemplateOutputPort) port.TemplateInputPort {
	return func(repo port.TemplateRepository, tx port.TxManager, output port.TemplateOutputPort) port.TemplateInputPort {
		return usecase.NewTemplateInteractor(repo, tx, output)
	}
}

// NewNoteInputFactory returns a factory for NoteInteractor.
func NewNoteInputFactory() func(noteRepo port.NoteRepository, tplRepo port.TemplateRepository, tx port.TxManager, output port.NoteOutputPort) port.NoteInputPort {
	return func(noteRepo port.NoteRepository, tplRepo port.TemplateRepository, tx port.TxManager, output port.NoteOutputPort) port.NoteInputPort {
		return usecase.NewNoteInteractor(noteRepo, tplRepo, tx, output)
	}
}

// NewNoteCommandInputFactory returns a factory for NoteCommandInteractor.
func NewNoteCommandInputFactory() func(commandRepo port.NoteCommandRepository, queryRepo port.NoteQueryRepository, tplRepo port.TemplateRepository, tx port.TxManager, output port.NoteCommandOutputPort) port.NoteCommandInputPort {
	return func(commandRepo port.NoteCommandRepository, queryRepo port.NoteQueryRepository, tplRepo port.TemplateRepository, tx port.TxManager, output port.NoteCommandOutputPort) port.NoteCommandInputPort {
		return usecase.NewNoteCommandInteractor(commandRepo, queryRepo, tplRepo, tx, output)
	}
}

// NewNoteQueryInputFactory returns a factory for NoteQueryInteractor.
func NewNoteQueryInputFactory() func(queryRepo port.NoteQueryRepository, output port.NoteQueryOutputPort) port.NoteQueryInputPort {
	return func(queryRepo port.NoteQueryRepository, output port.NoteQueryOutputPort) port.NoteQueryInputPort {
		return usecase.NewNoteQueryInteractor(queryRepo, output)
	}
}
