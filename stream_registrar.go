package main

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

type streamRegistrar interface {
	registerStream(name string) string
	unregisterStream(id string)
	checkStream(id string) (string, bool)
	listStreamIds() []string
}

type streamRegistrarImpl struct {
	store sync.Map
}

func newStreamRegistrar() *streamRegistrarImpl {
	s := &streamRegistrarImpl{}
	s.store.Store("test", "")
	return s
}

func (r *streamRegistrarImpl) registerStream(name string) string {
	id := uuid.New()
	r.store.LoadOrStore(id.String(), name)
	return id.String()
}

func (r *streamRegistrarImpl) unregisterStream(id string) {
	r.store.Delete(id)
}

func (r *streamRegistrarImpl) checkStream(id string) (string, bool) {
	name, ok := r.store.Load(id)
	if ok {
		return name.(string), ok
	} else {
		fmt.Println("not found " + id)
		return "", false
	}
}

func (r *streamRegistrarImpl) listStreamIds() []string {
	ids := make([]string, 0)
	r.store.Range(func(id interface{}, r interface{}) bool {
		ids = append(ids, id.(string))
		return true
	})
	return ids
}
