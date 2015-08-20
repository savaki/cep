package cep

import "time"

type EventType int64

type UnixTime int64

const (
	Undefined EventType = iota
	Foo
	Bar
)

type Event struct {
	Time UnixTime
	Type EventType
}

// ----------------------------------------------------------------

type Statement struct {
	Expression Expression
	FollowedBy *FollowedBy
	Triggers   EventType

	// may multiple instances of this statement be in effect at one time
	AllowMultiple bool
}

type FollowedBy struct {
	Statement *Statement
	Within    time.Duration
}

// ----------------------------------------------------------------

type Context interface {
	Unregister(id string) bool
}

type Flow interface {
	Id() string
	OnEvent(Context, Event) []Event
}

type StatementFlow struct {
	id        string
	expiresAt UnixTime
	Statement *Statement
}

func (s *StatementFlow) Id() string {
	return s.id
}

func (s *StatementFlow) OnEvent(ctx Context, event Event) []Event {
	if s.expiresAt != 0 && s.expiresAt < event.Time {
		ctx.Unregister(s.id)
		return nil
	}

	if !s.Statement.Expression.Matches(event) {
		return nil
	}

	if followedBy := s.Statement.FollowedBy; followedBy != nil {
		s.Statement = followedBy.Statement
		s.expiresAt = event.Time + UnixTime(followedBy.Within.Seconds())
		return nil
	}

	ctx.Unregister(s.id)
	return []Event{
		{
			Type: s.Statement.Triggers,
		},
	}
}
