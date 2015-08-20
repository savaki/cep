package cep

type Expression interface {
	Matches(Event) bool
}

type Equals struct {
	Type EventType
}

func (e *Equals) Matches(event Event) bool {
	return event.Type == e.Type
}

type NotEquals struct {
	Type EventType
}

func (n *NotEquals) Matches(event Event) bool {
	return event.Type != n.Type
}

type Or struct {
	Expressions []Expression
}

func (o *Or) Matches(event Event) bool {
	for _, expr := range o.Expressions {
		if expr.Matches(event) {
			return true
		}
	}

	return false
}

type And struct {
	Expressions []Expression
}

func (a *And) Matches(event Event) bool {
	for _, expr := range a.Expressions {
		if !expr.Matches(event) {
			return false
		}
	}

	return true
}
