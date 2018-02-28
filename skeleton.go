package skelington

import "strings"

// A struct generated to specification containing a flat list of Handle, hook
// functionality, and statistics.
type Skeleton struct {
	Has  []Handle
	Stat Statistics
	Hooks
}

func newSkeleton() *Skeleton {
	s := &Skeleton{
		make([]Handle, 0), nil, nil,
	}
	s.Hooks = newHooks(s)
	return s
}

// Adds any number of Handle instance to Skeleton instance, if Always is true,
// all hooks will be run after Handle is added.
func (s *Skeleton) Add(nhs ...Handle) error {
	preErr := s.RunHook(HPre)
	if preErr != nil {
		return preErr
	}
	s.Has = append(s.Has, nhs...)
	postErr := s.RunHook(HPost)
	return postErr
}

// A function taking a Skeleton instance and returning an error.
type SkeletonHook func(*Skeleton) error

type HookTiming int

const (
	HBefore HookTiming = iota
	HAfter
	HPre
	HPost
)

// An interface for hooks to be used by a Skeleton. Provides for setting hooks
// before & after adding all handles, as well as hooks run pre and post individual
// handle addition.
type Hooks interface {
	AddHook(HookTiming, ...SkeletonHook)
	RunHook(HookTiming) error
}

type hooks struct {
	s      *Skeleton
	before []SkeletonHook
	after  []SkeletonHook
	pre    []SkeletonHook
	post   []SkeletonHook
}

func newHooks(s *Skeleton) *hooks {
	return &hooks{
		s,
		make([]SkeletonHook, 0),
		make([]SkeletonHook, 0),
		make([]SkeletonHook, 0),
		make([]SkeletonHook, 0),
	}
}

func (h *hooks) AddHook(t HookTiming, sh ...SkeletonHook) {
	switch t {
	case HBefore:
		h.before = append(h.before, sh...)
	case HAfter:
		h.after = append(h.after, sh...)
	case HPre:
		h.pre = append(h.pre, sh...)
	case HPost:
		h.post = append(h.post, sh...)
	}
}

func hookRun(h []SkeletonHook, s *Skeleton) error {
	for _, fn := range h {
		err := fn(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *hooks) RunHook(t HookTiming) error {
	var err error
	switch t {
	case HBefore:
		err = hookRun(h.before, h.s)
	case HAfter:
		err = hookRun(h.after, h.s)
	case HPre:
		err = hookRun(h.pre, h.s)
	case HPost:
		err = hookRun(h.post, h.s)
	}
	return err
}

// A hook that sequences all Handle through categorization and sorting.
func SkeletonSequence(s *Skeleton) error {
	return sortBy(s, categorize(s))
}

func categorize(s *Skeleton) []string {
	c := make(map[string]*Tag)
	for _, v := range s.Has {
		t := v.Unit()
		c[t.Value] = t
	}
	var ret []string
	for k, _ := range c {
		ret = append(ret, k)
	}
	return ret
}

func sortBy(s *Skeleton, categories []string) error {
	sort := make(map[string][]Handle)
	for _, k := range categories {
		sort[k] = make([]Handle, 0)
	}
	add := func(hn Handle) {
		u := hn.Unit()
		sort[u.Value] = append(sort[u.Value], hn)
	}
	for _, hn := range s.Has {
		add(hn)
	}
	sequenceCategories(sort)
	var nh []Handle
	for _, v := range sort {
		nh = append(nh, v...)
	}
	s.Has = nh
	return nil
}

func sequenceCategories(m map[string][]Handle) {
	for _, v := range m {
		for i, vv := range v {
			vv.SetSequence(&Sequence{i + 1, len(v)})
		}
	}
}

// Sets the statistics hook for the skeleton, taking two array of StatFunc. The first are run
// once per handle, the second run on every invocation of a per handle once function.
// Statistics are not run until this hook is set manually.
func SkeletonCallsStatistics(s *Skeleton, once []StatFunc, every []StatFunc) {
	var stats = newStats()
	for _, h := range s.Has {
		t := h.Unit()
		tag := strings.ToUpper(t.Value)
		once = append(once, func(m map[string]int) {
			stats.d[tag] = stats.d[tag] + 1
		})
	}
	stats.onceFn = once

	s.AddHook(
		HPost,
		func(s *Skeleton) error {
			stats.once()
			s.Stat = stats.d
			return nil
		})
}

// A map of string key to int values for Skeleton statistics.
type Statistics map[string]int

type stats struct {
	d       map[string]int
	onceFn  []StatFunc
	everyFn []StatFunc
}

func newStats() *stats {
	return &stats{
		d:       make(map[string]int),
		everyFn: everyFuncs,
	}
}

func (s *stats) once() {
	for _, fn := range s.onceFn {
		fn(s.d)
		s.every()
	}
}

func (s *stats) every() {
	for _, fn := range s.everyFn {
		fn(s.d)
	}
}

// A statistics function taking a map of string key to int values.
type StatFunc func(map[string]int)

var everyFuncs = []StatFunc{
	func(m map[string]int) { m["TOTAL"] = m["TOTAL"] + 1 },
}

// Sets a hook that sets HandleFunc per handle across the entire skeleton.
func SkeletonCallsHandle(s *Skeleton, hf ...HandleFunc) {
	for _, h := range s.Has {
		h.SetCall(hf...)
	}
	s.AddHook(
		HPost,
		func(s *Skeleton) error {
			var err error
			for _, h := range s.Has {
				err = h.Call()
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
}
