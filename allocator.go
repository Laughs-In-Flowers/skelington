package skelington

// An interface that handles Skelington allocation by specific strategy.
type Allocator interface {
	Tag() string
	New() Allocator
	Allocate(*Skelington, Pather, Pather, string, ErrorHandler) *Skelington
}

type innerOpenFn func(Pather, Pather, string) (*Level, error)

func openNone(Pather, Pather, string) (*Level, error) {
	return nil, nil
}

func openFile(file Pather, root Pather, offset string) (*Level, error) {
	path := file.Path()
	return ReadFromFile(path)
}

func openDir(file Pather, root Pather, offset string) (*Level, error) {
	path := root.Path()
	return ReadFromDirectory(path, offset)
}

type innerAllocatorFn func(*Skelington, *Level, *Tag, string, ErrorHandler) *Skelington

type allocator struct {
	tag string
	ofn innerOpenFn
	afn innerAllocatorFn
	l   *Level
}

func newAllocator(tag string, ofn innerOpenFn, afn innerAllocatorFn) Allocator {
	return &allocator{tag, ofn, afn, nil}
}

// A tag for this allocator.
func (a *allocator) Tag() string {
	return a.tag
}

// Provides a new instance of the allocator for use.
func (a *allocator) New() Allocator {
	na := *a
	return &na
}

// The primary allocation function of the allocator. Provided two pathers, an offset string
// and an Errorhandler function, allocates and returns a new Skelington instance.
func (a *allocator) Allocate(s *Skelington, p Pather, r Pather, offset string, eh ErrorHandler) *Skelington {
	lv, err := a.ofn(p, r, offset)
	if err != nil {
		eh(err)
		return nil
	}
	a.l = lv
	root := r.GetTag()
	return a.afn(s, a.l, root, offset, eh)
}

func isOffset(o string, z *Level) *Level {
	if o != "" {
		if nz := offset(z, o); nz != nil {
			return nz
		}
	}
	return z
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

// An empty allocation, i.e. returns a skeleton with nothing.
func empAllocate(s *Skelington, z *Level, root *Tag, offset string, eh ErrorHandler) *Skelington {
	return s
}

//wsiwyg allocator, no calc, no branch, nothing just turn file to handles

// A continually reallocating shrinking proportion allocation. Attempts to
// allocate handles by proportion of handles remaining to allocate.
func rspAllocate(s *Skelington, z *Level, root *Tag, offset string, eh ErrorHandler) *Skelington {
	z = isOffset(offset, z)

	err := enumerate(z, z.Number)
	if err != nil {
		eh(err)
		return nil
	}

	s.AddHook(HPost, SkelingtonSequence)
	s.RunHook(HBefore)
	add := make([]Handle, 0)
	for _, lv := range flatten(z) {
		for i := 1; i <= lv.Actual; i++ {
			nh := newHandle(lv.sequence, root, lv.Family(), lv.Unit())
			add = append(add, nh)
		}
	}
	s.Add(add...)
	s.RunHook(HAfter)
	return s
}

// A branching expansion allocation. Branches expand from a root to create handles
// as directed and necessary.
func bgeAllocate(s *Skelington, z *Level, root *Tag, offset string, eh ErrorHandler) *Skelington {
	z = isOffset(offset, z)

	z.Iter(branch)

	s.AddHook(HPost, SkelingtonSequence)
	s.RunHook(HBefore)
	add := make([]Handle, 0)
	fn := func(lv *Level) {
		if isLeaf(lv) {
			nh := newHandle(lv.sequence, root, lv.Family(), lv.Unit())
			add = append(add, nh)
		}
	}
	z.Iter(fn)
	s.Add(add...)
	s.RunHook(HAfter)

	return s
}

// An allocation derived an existing directory of files.
func edfAllocate(s *Skelington, z *Level, root *Tag, offset string, eh ErrorHandler) *Skelington {
	toAdd := make(map[*Level]int)
	add := make([]Handle, 0)
	s.AddHook(HPost, SkelingtonSequence)
	s.RunHook(HBefore)
	z.Iter(func(iv *Level) {
		if iv.Number > 0 {
			toAdd[iv] = iv.Number - 1
		}
	})
	for k, v := range toAdd {
		for i := 0; i <= v; i = i + 1 {
			nh := newHandle(k.sequence, root, k.Family(), k.Unit())
			add = append(add, nh)
		}
	}
	s.Add(add...)
	s.RunHook(HAfter)
	return s
}

type allocators struct {
	has map[string]Allocator
}

// Provided a string key, attempts to return a new Allocator of that key.
func (a *allocators) Get(k string) Allocator {
	if g, ok := a.has[k]; ok {
		return g.New()
	}
	return nil
}

// Sets an allocator instance for future use.
func (a *allocators) Set(c Allocator) {
	a.has[c.Tag()] = c
}

// A struct maintaining available allocators.
// Package defaults provide the following allocators:
// emp - only provides an empty Skelington instance for further use.
// rsp - reallocating shrinking proportion
// bge - branching expansion
// edf - existing directory of files
var Allocators *allocators

func init() {
	Allocators = &allocators{make(map[string]Allocator, 0)}
	Allocators.Set(newAllocator("emp", openNone, empAllocate))
	Allocators.Set(newAllocator("rsp", openFile, rspAllocate))
	Allocators.Set(newAllocator("bge", openFile, bgeAllocate))
	Allocators.Set(newAllocator("edf", openDir, edfAllocate))
}
