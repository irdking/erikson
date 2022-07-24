package erikson

import "sync"

type View interface {
	View() []Node
}

type Group interface {
	View
	Get(id string) Node
	Add(n Node)
	Remove(n Node)
}

type SourcedView struct {
	ScrapedSource
	state []Node
	lock  sync.RWMutex
}

func NewSourcedView(s ScrapedSource) View {

	f := ScrapeAsync(s)
	init := <-f

	var l sync.RWMutex
	sv := SourcedView{s, init, l}
	go sv.collect()
	return sv

}

// test if more optimistic locking is better
func (sv SourcedView) View() []Node {

	defer sv.lock.RUnlock()
	sv.lock.RLock()
	out := make([]Node, len(sv.state))

	copy(sv.state, out)

	return out

}

// collects updates from source channel and applies them as state
func (sv SourcedView) collect() {

	for x := range sv.C() {
		sv.lock.Lock()
		sv.state = x
		sv.lock.Unlock()
	}
}
