package lsp

import (
	"container/list"
	"context"
	"sync"
	"time"

	"go.lsp.dev/protocol"
	"go.lsp.dev/uri"
)

type _DiagnosticBatch struct {
	generation uint64
	client     protocol.Client
	params     *protocol.PublishDiagnosticsParams
	removed    bool
}

// The publisher owns transport I/O. There is at most one queued batch per URI;
// unopened documents that disappear before publication need no clearing batch.
// Memory is bounded by current documents plus documents still visible to the
// client, independently of the number of edits received while a send is blocked.
type _DiagnosticPublisher struct {
	mu               sync.Mutex
	generation       uint64
	pending          map[uri.URI]*list.Element
	order            list.List
	visible          map[uri.URI]bool
	active           uri.URI
	running, stopped bool
	ctx              context.Context
	cancel           context.CancelFunc
}

func newDiagnosticPublisher() *_DiagnosticPublisher {
	ctx, cancel := context.WithCancel(context.Background())
	return new(_DiagnosticPublisher{pending: map[uri.URI]*list.Element{}, visible: map[uri.URI]bool{}, ctx: ctx, cancel: cancel})
}

// advance discards queued obsolete batches and returns the URIs whose current
// syntax must be republished. An in-flight send precedes every newer batch.
func (p *_DiagnosticPublisher) advance(generation uint64) []uri.URI {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.generation = generation
	uris := make([]uri.URI, 0, len(p.pending))
	for documentURI := range p.pending {
		uris = append(uris, documentURI)
	}
	p.pending = map[uri.URI]*list.Element{}
	p.order.Init()
	return uris
}

func (p *_DiagnosticPublisher) enqueue(batch _DiagnosticBatch) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped || batch.generation != p.generation {
		return
	}
	documentURI := batch.params.URI
	if batch.removed && !p.visible[documentURI] && p.active != documentURI {
		if previous := p.pending[documentURI]; previous != nil {
			p.order.Remove(previous)
			delete(p.pending, documentURI)
		}
		return
	}
	if previous := p.pending[documentURI]; previous != nil {
		previous.Value = batch
	} else {
		p.pending[documentURI] = p.order.PushBack(batch)
	}
	if !p.running {
		p.running = true
		go p.run()
	}
}

func (p *_DiagnosticPublisher) run() {
	for {
		p.mu.Lock()
		first := p.order.Front()
		if p.stopped || first == nil {
			p.running = false
			p.mu.Unlock()
			return
		}
		batch := first.Value.(_DiagnosticBatch)
		p.order.Remove(first)
		delete(p.pending, batch.params.URI)
		if batch.generation != p.generation {
			p.mu.Unlock()
			continue
		}
		p.active = batch.params.URI
		p.mu.Unlock()

		ctx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
		err := batch.client.PublishDiagnostics(ctx, batch.params)
		cancel()

		p.mu.Lock()
		p.active = ""
		if err == nil {
			if batch.removed {
				delete(p.visible, batch.params.URI)
			} else {
				p.visible[batch.params.URI] = true
			}
		}
		p.mu.Unlock()
	}
}

func (p *_DiagnosticPublisher) stop() {
	p.mu.Lock()
	p.stopped = true
	p.pending = map[uri.URI]*list.Element{}
	p.order.Init()
	p.mu.Unlock()
	p.cancel()
}
