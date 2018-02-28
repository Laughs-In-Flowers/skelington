package skelington

//
type Allocator interface {
	Tag() string
	New() Allocator
	Allocate(Pather, Pather, string, ErrorHandler) *Skeleton
}

//
type ReadFrom int

const (
	RFNone ReadFrom = iota
	RFConfFile
	RFDirectory
)

//
type InnerAllocatorFn func(*Level, *Tag, string, ErrorHandler) *Skeleton

type allocator struct {
	tag string
	rf  ReadFrom
	fn  InnerAllocatorFn
	l   *Level
}

func newAllocator(tag string, rf ReadFrom, fn InnerAllocatorFn) Allocator {
	return &allocator{tag, rf, fn, nil}
}

//
func (a *allocator) Tag() string {
	return a.tag
}

//
func (a *allocator) New() Allocator {
	na := *a
	return &na
}

//
func (a *allocator) Allocate(p Pather, r Pather, offset string, eh ErrorHandler) *Skeleton {
	err := open(a, p, r)
	if err != nil {
		eh(err)
		return nil
	}
	root := r.Tag()
	return a.fn(a.l, root, offset, eh)
}

func open(a *allocator, p, r Pather) error {
	var lv *Level
	var err error
	switch a.rf {
	case RFConfFile:
		path := p.Path()
		lv, err = ReadFromFile(path)
	case RFDirectory:
		root := r.Path()
		lv, err = ReadFromDirectory(root)
	}
	if err != nil {
		return err
	}
	a.l = lv
	return nil
}

func isOffset(offset string, z *Level) *Level {
	if offset != "" {
		if nz := z.Offset(offset); nz != nil {
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
func EMPAllocate(z *Level, root *Tag, offset string, eh ErrorHandler) *Skeleton {
	return newSkeleton()
}

// A continually reallocating shrinking proportion allocation. Given a number,
// will attempt to allocate handles by proportion of handles remaining to allocate.
func RSPAllocate(z *Level, root *Tag, offset string, eh ErrorHandler) *Skeleton {
	z = isOffset(offset, z)

	err := enumerate(z, z.Number)
	if err != nil {
		eh(err)
		return nil
	}

	s := newSkeleton()
	s.AddHook(HPost, SkeletonSequence)
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

// A branching expansion allocation. From the root will branch and create handles
// as directed and necessary.
func BGEAllocate(z *Level, root *Tag, offset string, eh ErrorHandler) *Skeleton {
	z = isOffset(offset, z)

	z.Iter(branch)

	s := newSkeleton()
	s.AddHook(HPost, SkeletonSequence)
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

// An allocation derived existing directory of files.
func EDFAllocate(z *Level, root *Tag, offset string, eh ErrorHandler) *Skeleton {
	s := newSkeleton()
	//add hooks
	s.RunHook(HBefore)
	add := make([]Handle, 0)
	s.Add(add...)
	s.RunHook(HAfter)

	return s
}

type allocators struct {
	has map[string]Allocator
}

//
func (a *allocators) Get(k string) Allocator {
	if g, ok := a.has[k]; ok {
		return g.New()
	}
	return nil
}

//
func (a *allocators) Set(c Allocator) {
	a.has[c.Tag()] = c
}

//
var Allocators *allocators

func init() {
	Allocators = &allocators{make(map[string]Allocator, 0)}
	Allocators.Set(newAllocator("emp", RFNone, EMPAllocate))
	Allocators.Set(newAllocator("rsp", RFConfFile, RSPAllocate))
	Allocators.Set(newAllocator("bge", RFConfFile, BGEAllocate))
	Allocators.Set(newAllocator("edf", RFDirectory, EDFAllocate))
}
