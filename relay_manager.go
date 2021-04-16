package main

import "sync"

type relayAccessor interface {
	getRelay(id string) *streamRelay
	deleteRelay(id string)
	createRelay(id string) *streamRelay
}

type relayManager struct {
	store sync.Map
}

func newRelayManager() *relayManager {
	return &relayManager{}
}

func (m *relayManager) getRelay(id string) *streamRelay {
	if relay, ok := m.store.Load(id); ok {
		return relay.(*streamRelay)
	}
	return nil
}

func (m *relayManager) createRelay(id string) *streamRelay {
	relay, err := newStreamRelay(id)
	if err != nil {
		logger.Errorf("failed to create relay, %v\n", err)
		return nil
	}
	m.store.Store(id, relay)
	return relay
}

func (m *relayManager) deleteRelay(id string) {
	m.store.Delete(id)
}
