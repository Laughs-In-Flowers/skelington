package skelington

import "strings"

type Handles struct {
	has []Handle
}

func NewHandles() *Handles {
	return &Handles{
		make([]Handle, 0),
	}
}

type HandlesFunc func(Handle)

func (h *Handles) Iter(fn HandlesFunc) {
	for _, v := range h.has {
		fn(v)
	}
}

func (h *Handles) Add(nhs ...Handle) {
	h.has = append(h.has, nhs...)
	h.post()
}

func (h *Handles) post() {
	c := categorize(h)
	sortBy(h, c)
}

func categorize(h *Handles) []string {
	c := make(map[string]*Tag)
	for _, v := range h.has {
		t := v.Unit()
		c[t.Value] = t
	}
	var ret []string
	for k, _ := range c {
		ret = append(ret, k)
	}
	return ret
}

func sortBy(h *Handles, categories []string) {
	sort := make(map[string][]Handle)
	for _, k := range categories {
		sort[k] = make([]Handle, 0)
	}
	add := func(hn Handle) {
		u := hn.Unit()
		sort[u.Value] = append(sort[u.Value], hn)
	}
	for _, hn := range h.has {
		add(hn)
	}
	sequence(sort)
	var nh []Handle
	for _, v := range sort {
		nh = append(nh, v...)
	}
	h.has = nh
}

func sequence(m map[string][]Handle) {
	for _, v := range m {
		for i, vv := range v {
			vv.SetSequence(&Sequence{i + 1, len(v)})
		}
	}
}

type stats struct {
	d     map[string]int
	funcs []statFunc
}

func (s *stats) run() {
	for _, fn := range s.funcs {
		fn(s.d)
	}
}

func statistics() *stats {
	return &stats{
		d:     make(map[string]int),
		funcs: defaultStatFuncs,
	}
}

type statFunc func(map[string]int)

var defaultStatFuncs = []statFunc{
	func(m map[string]int) { m["TOTAL"] = m["TOTAL"] + 1 },
}

func (h *Handles) Statistics() map[string]int {
	var stats = statistics()
	for _, v := range h.has {
		t := v.Unit()
		tag := strings.ToUpper(t.Value)
		stats.d[tag] = stats.d[tag] + 1
		stats.run()
	}
	return stats.d
}
