package skelington

// A struct generated to specification containing a flattened array of Handle
// and hook functionality.
type Skelington struct {
	Has []Handle
	Hooks
	Statistic
}

// Creates new Skelington instance from provided Config
func New(cnf ...Config) (*Skelington, error) {
	p, pErr := newProcessor(cnf...)
	if pErr != nil {
		return nil, pErr
	}
	s := p.Process()
	return s, nil
}

func newSkelington(
	h map[HookTiming][]SkelingtonHook,
	m map[string][]StatFunc,
) *Skelington {
	s := &Skelington{
		make([]Handle, 0), nil, nil,
	}
	s.Hooks = newHooks(s)
	for k, v := range h {
		s.AddHook(k, v...)
	}
	s.Statistic = newStat(s, m)
	return s
}

// Adds any number of Handle instance to Skelington instance.
func (s *Skelington) Add(nhs ...Handle) error {
	preErr := s.RunHook(HPre)
	if preErr != nil {
		return preErr
	}
	s.Has = append(s.Has, nhs...)
	postErr := s.RunHook(HPost)
	return postErr
}

//func (s *Skelington) Clear() {
//reset all hooks to defaults
//empty handles
//reset stats
//}

// A function taking a Skelington instance and returning an error.
type SkelingtonHook func(*Skelington) error

// A type for specifying hook timing.
type HookTiming int

const (
	HBefore HookTiming = iota // before all handles are added
	HPre                      // before any handle is added
	HPost                     // after any handle is added
	HAfter                    // after all handles are added
)

// An interface for hooks to be used by a Skelington. Provides for setting hooks
// before & after adding all handles, as well as hooks run pre and post individual
// handle addition.
type Hooks interface {
	AddHook(HookTiming, ...SkelingtonHook)
	RunHook(HookTiming) error
}

type hooks struct {
	s *Skelington
	m map[HookTiming][]SkelingtonHook
}

func newHooks(s *Skelington) *hooks {
	return &hooks{
		s,
		make(map[HookTiming][]SkelingtonHook),
	}
}

func (h *hooks) getHooks(t HookTiming) []SkelingtonHook {
	if _, exists := h.m[t]; !exists {
		h.setHooks(t, make([]SkelingtonHook, 0))
	}
	return h.m[t]
}

func (h *hooks) setHooks(t HookTiming, hs []SkelingtonHook) {
	h.m[t] = hs
}

// Adds the provided SkelingtonHook for the provided HookTiming.
func (h *hooks) AddHook(t HookTiming, sh ...SkelingtonHook) {
	hs := h.getHooks(t)
	hs = append(hs, sh...)
	h.setHooks(t, hs)
}

func hookRun(h []SkelingtonHook, s *Skelington) error {
	for _, fn := range h {
		err := fn(s)
		if err != nil {
			return err
		}
	}
	return nil
}

// Run all hooks matching the provided HookTiming.
func (h *hooks) RunHook(t HookTiming) error {
	return hookRun(h.getHooks(t), h.s)
}

// A sequencing hook.
func SkelingtonSequence(s *Skelington) error {
	return sortBy(s, categorize(s))
}

func categorize(s *Skelington) []string {
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

func sortBy(s *Skelington, categories []string) error {
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

// Sets a hook that sets HandleFunc per handle across the entire skeleton, and
// executing the Call function for all handles in the post add hook.
func SkelingtonHandleCalls(s *Skelington, hf ...HandleCall) {
	for _, h := range s.Has {
		h.SetCall(hf...)
	}
	s.AddHook(
		HPost,
		func(s *Skelington) error {
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
