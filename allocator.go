package skelington

type Allocator interface {
	New() Allocator
	Open(string) error
	Allocate(*Tag, string) (*Handles, error)
}

// Reallocating Shrinking Pie
type rsp struct {
	l *Level
}

func (r *rsp) New() Allocator {
	na := *r
	return &na
}

func (r *rsp) Open(path string) error {
	lv, err := Read(path)
	if err != nil {
		return err
	}
	r.l = lv
	return nil
}

func enumerate(lv *Level, from int) error {
	var numRelative int

	for _, level := range lv.Levels {
		if level.Relative {
			level.Percent = (float64(level.Number) / 100)
			numRelative++
		} else {
			level.Percent = float64(level.Number) / float64(from)
		}
	}

	if numRelative > 0 {
		var actualP float64
		for _, level := range lv.Levels {
			actualP = actualP + level.Percent
		}
		if actualP != 1 {
			distribute := (1 - actualP) / float64(numRelative)
			for _, level := range lv.Levels {
				if level.Relative {
					level.Percent = level.Percent + distribute
				}
			}
		}
	}

	for _, level := range lv.Levels {
		level.Actual = int(float64(from) * level.Percent)

		err := enumerate(
			level,
			level.Actual,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *rsp) Allocate(root *Tag, offset string) (*Handles, error) {
	var z *Level = r.l

	if offset != "" {
		if o := z.Offset(offset); o != nil {
			z = o
		}
	}

	err := enumerate(z, z.Number)
	if err != nil {
		return nil, err
	}

	h := NewHandles()
	add := make([]Handle, 0)

	for _, lv := range flatten(z) {
		for i := 1; i <= lv.Actual; i++ {
			nh := newHandle(&Sequence{i, lv.Actual}, root, lv.Family(), lv.Unit())
			add = append(add, nh)
		}
	}

	h.Add(add...)

	return h, nil
}

// Branching Expansion
type bge struct {
	l *Level
}

func (b *bge) New() Allocator {
	na := *b
	return &na
}

func (b *bge) Open(path string) error {
	lv, err := Read(path)
	if err != nil {
		return err
	}
	b.l = lv
	return nil
}

func (b *bge) Allocate(root *Tag, offset string) (*Handles, error) {
	z := b.l

	if offset != "" {
		if o := z.Offset(offset); o != nil {
			z = o
		}
	}

	z.Iter(branch)

	h := NewHandles()
	add := make([]Handle, 0)

	fn := func(lv *Level) {
		if isLeaf(lv) {
			nh := newHandle(nil, root, lv.Family(), lv.Unit())
			add = append(add, nh)
		}
	}

	z.Iter(fn)

	h.Add(add...)

	return h, nil
}

type allocators struct {
	has map[string]Allocator
}

func (a *allocators) Get(k string) Allocator {
	if g, ok := a.has[k]; ok {
		return g.New()
	}
	return nil
}

func (a *allocators) Set(k string, fn Allocator) {
	a.has[k] = fn
}

var Allocators *allocators

func init() {
	Allocators = &allocators{make(map[string]Allocator, 0)}
	Allocators.Set("rsp", &rsp{})
	Allocators.Set("bge", &bge{})
}
