package cep

import "github.com/mitchellh/copystructure"

type Emitter interface {
	OnEvent(Event) (Flow, error)
}

type StatementEmitter struct {
	Statement Statement
}

func (e *StatementEmitter) OnEvent(event Event) (Flow, error) {
	if !e.Statement.Expression.Matches(event) {
		return nil, nil
	}

	dup, err := copystructure.Copy(&e.Statement)
	if err != nil {
		return nil, err
	}

	return &StatementFlow{
		Statement: dup.(*Statement),
	}, nil
}
