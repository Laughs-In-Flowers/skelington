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
	Handles
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

// An interface for attaching any arbitrary item to the handle as an interface{}.
type Handles interface {
	Item() interface{}
	SetItem(interface{})
}

type handle struct {
	id       string
	sequence *Sequence
	root     *Tag
	family   []*Tag
	unit     *Tag
	calls    []HandleCall
	item     interface{}
}

func newHandle(s *Sequence, root *Tag, family []*Tag, unit *Tag) *handle {
	h := &handle{
		uuidString(),
		s,
		root,
		family,
		unit,
		make([]HandleCall, 0),
		nil,
	}
	return h
}

// Returns the handle Sequence.
func (h *handle) Sequence() *Sequence {
	return h.sequence
}

// Sets the provided Sequence to the handle.
func (h *handle) SetSequence(s *Sequence) {
	h.sequence = s
}

// Returns the handle's root Tag.
func (h *handle) Root() *Tag {
	return h.root
}

// Returns an array of Tag corresponding to the handles family.
func (h *handle) Family() []*Tag {
	return h.family
}

// Returns the handle's unit Tag.
func (h *handle) Unit() *Tag {
	return h.unit
}

// An array of Tag instances for sorting.
type TagSort []*Tag

func (s TagSort) Len() int           { return len(s) }
func (s TagSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s TagSort) Less(i, j int) bool { return s[i].Order < s[j].Order }
func (s TagSort) Sort()              { sort.Sort(s) }
func (s TagSort) List() []string {
	var ret []string
	for _, v := range s {
		ret = append(ret, v.Value)
	}
	return ret
}

// Returns a sorted array of Tag for this handle. If provided parameter is true,
// a separate tag is added corresponding to the handle's Sequence.
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
	ret.Sort()
	return ret
}

// Returns a string key for Pather interface.
func (h *handle) Key() string {
	return h.id
}

// Returns a string path for Pather interface.
func (h *handle) Path() string {
	s := h.Tagged(true)
	return filepath.Join(s.List()...)
}

// SetPath() for pather interface.  Does not set a path for this package
// specific type(allocation and tagging manages this).
func (h *handle) SetPath(path string) {
	// not implemented
}

// Returns a Tag, for Pather interface.
func (h *handle) GetTag() *Tag {
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

//
func (h *handle) Clear() {
	h.calls = make([]HandleCall, 0)
}

//
func (h *handle) Item() interface{} {
	return h.item
}

//
func (h *handle) SetItem(i interface{}) {
	h.item = i
}
