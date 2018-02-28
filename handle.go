package skelington

import (
	"path/filepath"
	"sort"
)

type Handle interface {
	Sequencer
	Tagger
	Pather
	Caller
	Ignorer
}

type Sequencer interface {
	Sequence() *Sequence
	SetSequence(*Sequence)
}

type Tagger interface {
	Root() *Tag
	Family() []*Tag
	Unit() *Tag
	Tagged(bool) TagSort
}

type HandleFunc func(Handle) error

type Caller interface {
	Call() error
	SetCall(...HandleFunc)
}

type Ignorer interface {
	Ignore()
	Ignored() bool
}

type handle struct {
	id       string
	sequence *Sequence
	root     *Tag
	family   []*Tag
	unit     *Tag
	calls    []HandleFunc
	ignore   bool
	callOnce bool
}

func newHandle(s *Sequence, root *Tag, family []*Tag, unit *Tag) *handle {
	h := &handle{
		V4Quick(),
		s,
		root,
		family,
		unit,
		make([]HandleFunc, 0),
		false,
		true,
	}
	return h
}

func (h *handle) Sequence() *Sequence {
	return h.sequence
}

func (h *handle) SetSequence(s *Sequence) {
	h.sequence = s
}

func (h *handle) Root() *Tag {
	return h.root
}

func (h *handle) Family() []*Tag {
	return h.family
}

func (h *handle) Unit() *Tag {
	return h.unit
}

type TagSort []*Tag

func (s TagSort) Len() int           { return len(s) }
func (s TagSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TagSort) Less(i, j int) bool { return s[i].Order < s[j].Order }
func (s TagSort) List() []string {
	var ret []string
	for _, v := range s {
		ret = append(ret, v.Value)
	}
	return ret
}

func (h *handle) Tagged(seq bool) TagSort {
	var ret TagSort
	ret = append(ret, h.Root())
	ret = append(ret, h.Family()...)
	u := h.Unit()
	ret = append(ret, u)
	if seq {
		s := h.Sequence()
		ret = append(ret, &Tag{u.Order + 1, s.String()})
	}
	sort.Sort(ret)
	return ret
}

// Key() for Pather interface
func (h *handle) Key() string {
	return h.id
}

// Path() for Pather interface
func (h *handle) Path() string {
	s := h.Tagged(true)
	return filepath.Join(s.List()...)
}

// SetPath() for pather interface, but do not set a path for this case
func (h *handle) SetPath(path string) {}

// Tag() for Pather interface
func (h *handle) Tag() *Tag {
	return &Tag{-1, h.Key()}
}

func (h *handle) Ignore() {
	h.ignore = true
}

func (h *handle) Ignored() bool {
	return h.ignore
}

func (h *handle) Call() error {
	var err error
	for _, fn := range h.calls {
		if !h.Ignored() {
			err = fn(h)
			if err != nil {
				return err
			}
		}
	}
	if h.callOnce {
		h.deleteCall()
	}
	return err
}

func (h *handle) SetCall(c ...HandleFunc) {
	h.calls = append(h.calls, c...)
}

func (h *handle) deleteCall() {
	h.calls = make([]HandleFunc, 0)
}
