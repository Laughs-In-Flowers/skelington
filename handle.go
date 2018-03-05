package skelington

import (
	"path/filepath"
	"sort"
)

// An interface encapsulating one particular abstract item handled by a Skelington.
type Handle interface {
	Sequencer
	Tagger
	Pather
	Caller
}

// A sequencing interface.
type Sequencer interface {
	Sequence() *Sequence
	SetSequence(*Sequence)
}

// A tagging interface.
type Tagger interface {
	Root() *Tag
	Family() []*Tag
	Unit() *Tag
	Tagged(bool) TagSort
}

// A function taking a Handle and returning an error.
type HandleCall func(Handle) error

// An interface for managing HandleFunc.
type Caller interface {
	Call() error
	SetCall(...HandleCall)
	Clear()
}

type handle struct {
	id       string
	sequence *Sequence
	root     *Tag
	family   []*Tag
	unit     *Tag
	calls    []HandleCall
}

func newHandle(s *Sequence, root *Tag, family []*Tag, unit *Tag) *handle {
	h := &handle{
		v4Quick(),
		s,
		root,
		family,
		unit,
		make([]HandleCall, 0),
	}
	return h
}

//
func (h *handle) Sequence() *Sequence {
	return h.sequence
}

//
func (h *handle) SetSequence(s *Sequence) {
	h.sequence = s
}

//
func (h *handle) Root() *Tag {
	return h.root
}

//
func (h *handle) Family() []*Tag {
	return h.family
}

//
func (h *handle) Unit() *Tag {
	return h.unit
}

// An array of Tag instances for sorting.
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

//
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

// Key() for Pather interface.
func (h *handle) Key() string {
	return h.id
}

// Path() for Pather interface.
func (h *handle) Path() string {
	s := h.Tagged(true)
	return filepath.Join(s.List()...)
}

// SetPath() for pather interface, but do not set a path for this case.
func (h *handle) SetPath(path string) {}

// Tag() for Pather interface.
func (h *handle) Tag() *Tag {
	return &Tag{-1, h.Key()}
}

// Runs through every set HandleFunc for this handle, returning any error immediately.
func (h *handle) Call() error {
	var err error
	for _, fn := range h.calls {
		err = fn(h)
		if err != nil {
			return err
		}
	}
	return err
}

// Sets any number of HandleFunc to be called on this handle.
func (h *handle) SetCall(c ...HandleCall) {
	h.calls = append(h.calls, c...)
}

func (h *handle) Clear() {
	h.calls = make([]HandleCall, 0)
}
