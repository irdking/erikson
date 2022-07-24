package erikson

import (
	"errors"
	"time"
)

type Labels map[string]string
type Metrics map[string]float64

type Node interface {
	Address() string
	Labels() Labels
	Metrics() Metrics
	WithLabels(l Labels)
	WithMetrics(m Metrics)
}

type Peer struct {
	address string
	metrics map[string]float64
	labels  map[string]string
}

func (p *Peer) Labels() map[string]string {
	return p.labels
}

func (p *Peer) Address() string { return p.address }
func (p *Peer) Metrics() map[string]float64 {
	return p.metrics
}

func NewPeer(addr string) Peer {
	metrics := map[string]float64{}
	labels := map[string]string{}
	return Peer{addr, metrics, labels}
}

type Source interface {
	C() chan []Node
}

type Scraper interface {
	Scrape() ([]Node, error)
}

func ScrapeAsync(s Scraper) chan []Node {

	out := make(chan []Node)

	f := func() {

		p, e := s.Scrape()

		if e != nil {
			out <- []Node{}
			return
		}

		out <- p
	}

	go f()

	return out
}

type joinScraper struct {
	a, b Scraper
}

// fetches all scrapers within asynchronously and returns output
func (s *joinScraper) Scrape() ([]Node, error) {
	l := ScrapeAsync(s.a)
	r := ScrapeAsync(s.b)

	a := <-l
	b := <-r

	out := append(a, b...)

	if len(out) == 0 {
		return out, errors.New("Empty State")
	}
	return out, nil
}

//the behaviour of the outputted scraper
// is to run both scrapers aynchronously await both and join the result
// remember you can also join the outputted scraper to eachother which scales very nicely
// so you can run 15 instances and only be bounded by the slowest call
// it ignores failures because there are multiple
func JoinScrapers(a Scraper, b Scraper) Scraper {
	return &joinScraper{a, b}
}

type ScrapedSource struct {
	c chan []Node
	t *time.Ticker
	Scraper
}

func (s *ScrapedSource) Push() {
	p, e := s.Scrape()

	if e != nil {
		return
	}

	s.c <- p

}

func (s *ScrapedSource) loop() {

	for _ = range s.t.C {
		s.Push()
	}
}

func (s *ScrapedSource) Stop() {
	s.t.Stop()
	close(s.c)
}

func (s *ScrapedSource) C() chan []Node {
	return s.C()
}

func NewScrapedSource(i time.Duration, s Scraper) ScrapedSource {

	c := make(chan []Node)
	t := time.NewTicker(i)

	return ScrapedSource{c, t, s}
}
