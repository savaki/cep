package cep

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type MockContext struct {
	Ids []string
}

func (m *MockContext) Unregister(id string) bool {
	if m.Ids == nil {
		m.Ids = []string{}
	}
	m.Ids = append(m.Ids, id)
	return true
}

var (
	mockContext      = &MockContext{}
	FarIntoTheFuture = UnixTime(time.Now().Second()) + UnixTime((100 * 365 * 24 * time.Hour).Seconds())
)

func BenchmarkStatement(b *testing.B) {
	stmt := Statement{
		Expression: &Equals{Type: Foo},
		Triggers:   Bar,
	}
	flow := StatementFlow{
		Statement: &stmt,
	}
	event := Event{
		Type: Foo,
	}

	for n := 0; n < b.N; n++ {
		flow.OnEvent(mockContext, event)
	}
}

func TestExpression(t *testing.T) {
	Convey("Given an Expression", t, func() {
		expr := &Equals{Type: Foo}

		Convey("I expect matching things to return true", func() {
			So(expr.Matches(Event{Type: Foo}), ShouldBeTrue)
		})

		Convey("And non-matches to return false", func() {
			So(expr.Matches(Event{Type: Bar}), ShouldBeFalse)
		})
	})
}

func TestStatementFlow(t *testing.T) {
	Convey("Given a Flow", t, func() {
		flow := StatementFlow{
			Statement: &Statement{
				Expression: &Equals{Type: Foo},
				Triggers:   Bar,
			},
		}

		Convey("When an event matches", func() {
			events := flow.OnEvent(mockContext, Event{Type: Foo})

			Convey("Then I expect the triggered event to be sent", func() {
				So(events, ShouldResemble, []Event{Event{Type: flow.Statement.Triggers}})
			})
		})

		Convey("When the event doesn't match", func() {
			events := flow.OnEvent(mockContext, Event{Type: Bar})

			Convey("Then I expect no events to be sent", func() {
				So(events, ShouldBeNil)
			})
		})
	})

	Convey("Given a Flow with a FollowedBy clause", t, func() {
		eventToTrigger := Bar
		flow := StatementFlow{
			Statement: &Statement{
				Expression: &Equals{
					Type: Foo,
				},
				FollowedBy: &FollowedBy{
					Statement: &Statement{
						Expression: &Equals{Type: Bar},
						Triggers:   eventToTrigger,
					},
					Within: 3 * time.Minute,
				},
			},
		}
		ctx := &MockContext{}

		Convey("When the correct sequence of events are invoked", func() {
			events1 := flow.OnEvent(ctx, Event{Type: Foo})
			So(len(ctx.Ids), ShouldEqual, 0) // flow should not yet be deregistered

			events2 := flow.OnEvent(ctx, Event{Type: Bar})

			Convey("Then I expect events to be emitted", func() {
				So(events1, ShouldBeNil)
				So(events2, ShouldResemble, []Event{Event{Type: eventToTrigger}})
			})

			Convey("And I expect the flow to be unregistered", func() {
				So(len(ctx.Ids), ShouldEqual, 1)
			})
		})

		Convey("When the correct sequence of events are invoked with some junk in the middle", func() {
			events1 := flow.OnEvent(ctx, Event{Type: Foo})
			events2 := flow.OnEvent(ctx, Event{Type: Foo})
			events3 := flow.OnEvent(ctx, Event{Type: Foo})
			events4 := flow.OnEvent(ctx, Event{Type: Foo})
			So(len(ctx.Ids), ShouldEqual, 0) // flow should not yet be deregistered

			events5 := flow.OnEvent(ctx, Event{Type: Bar})

			Convey("Then I expect events to be emitted", func() {
				So(events1, ShouldBeNil)
				So(events2, ShouldBeNil)
				So(events3, ShouldBeNil)
				So(events4, ShouldBeNil)
				So(events5, ShouldResemble, []Event{Event{Type: eventToTrigger}})
			})

			Convey("And I expect the flow to be unregistered", func() {
				So(len(ctx.Ids), ShouldEqual, 1)
			})
		})

		Convey("When an incorrect sequence of events are invoked", func() {
			events1 := flow.OnEvent(ctx, Event{Type: Foo})
			events2 := flow.OnEvent(ctx, Event{Type: Foo})

			Convey("Then I expect no events to be emitted", func() {
				So(events1, ShouldBeNil)
				So(events2, ShouldBeNil)
			})

			Convey("And I expect the flow to still be around", func() {
				So(len(ctx.Ids), ShouldBeZeroValue)
			})
		})

		Convey("When the time expires on a sequence", func() {
			events1 := flow.OnEvent(ctx, Event{Type: Foo})
			events2 := flow.OnEvent(ctx, Event{Time: FarIntoTheFuture, Type: Foo})

			Convey("Then I expect no events to be emitted", func() {
				So(events1, ShouldBeNil)
				So(events2, ShouldBeNil)
			})

			Convey("And I expect the flow to be unregistered", func() {
				So(len(ctx.Ids), ShouldEqual, 1)
			})
		})
	})
}
