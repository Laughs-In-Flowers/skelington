package skelington

type Handle interface {
	Sequence() *Sequence
	SetSequence(*Sequence)
	Root() *Tag
	Family() []*Tag
	Unit() *Tag
}

type handle struct {
	sequence *Sequence
	root     *Tag
	family   []*Tag
	unit     *Tag
	assets   []string
}

func newHandle(s *Sequence, root *Tag, family []*Tag, unit *Tag) *handle {
	return &handle{
		s, root, family, unit, make([]string, 0),
	}
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
