package cep

import (
	"fmt"
	"sync"
)

var (
	ErrFlowAlreadyExists = fmt.Errorf("flow already exists")
)

type Flows struct {
	entries map[string]Flow
	mux     sync.Mutex
}

func (f Flows) Register(flow Flow) error {
	f.mux.Lock()
	defer f.mux.Unlock()

	id := flow.Id()
	if _, found := f.entries[id]; found {
		return ErrFlowAlreadyExists
	}

	f.entries[id] = flow
	return nil
}

func (f Flows) Unregister(id string) bool {
	if _, found := f.entries[id]; !found {
		return false
	}

	delete(f.entries, id)
	return true
}

// ----------------------------------------------------------------

type Graph struct {
	Emitters []Emitter
	Flows    Flows
}

func (g *Graph) OnEvent(event Event) error {
	for _, emitter := range g.Emitters {
		flow, err := emitter.OnEvent(event)
		if err != nil {
			return err
		}

		if flow != nil {
			g.Flows.Register(flow)
		}
	}

	for _, flow := range g.Flows.entries {
		events := flow.OnEvent(g.Flows, event)
		if events != nil {
			for _, event := range events {
				g.OnEvent(event)
			}
		}
	}

	return nil
}

type FlowUnregisterFunc func(id string) bool
