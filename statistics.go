package skelington

import (
	"strings"
)

// An interface for statistics.
type Statistic interface {
	Get(string) []StatFunc
	Set(string, ...StatFunc)
	Run(*Skelington, string) error
	Report() map[string]int
	Reported(map[string]int)
	Reset()
}

// A statistics function taking a map of string key to int values.
type StatFunc func(*Skelington, map[string]int) error

type stat struct {
	d map[string]int
	m map[string][]StatFunc
}

func newStat(s *Skelington, fn map[string][]StatFunc) *stat {
	d := make(map[string]int)
	d["TOTAL"] = 0

	st := &stat{
		d: d,
		m: newFuncs(fn),
	}

	var totalStop, partStop bool = false, false

	s.AddHook(HBefore,
		func(s *Skelington) error {
			return st.Run(s, "before")
		})

	s.AddHook(HPre,
		func(s *Skelington) error {
			return st.Run(s, "pre")
		})

	s.AddHook(HPost,
		func(s *Skelington) error {
			if !totalStop {
				d["TOTAL"] = d["TOTAL"] + len(s.Has)
			}
			return nil
		},
		func(s *Skelington) error {
			return st.Run(s, "post")
		},
	)

	s.AddHook(HAfter,
		func(s *Skelington) error {
			handleTag(s, d, func(tag string, m map[string]int) {
				m[tag] = 0
			})
			return nil
		},
		func(s *Skelington) error {
			if !partStop {
				handleTag(s, d, func(tag string, m map[string]int) {
					m[tag] = m[tag] + 1
				})
			}
			return nil
		},
		func(s *Skelington) error {
			return s.Run(s, "after")
		},
		func(s *Skelington) error {
			totalStop = true
			partStop = true
			return nil
		},
	)

	return st
}

func newFuncs(in map[string][]StatFunc) map[string][]StatFunc {
	var k = []string{"before", "pre", "post", "after", "every"}
	out := make(map[string][]StatFunc)
	for _, kk := range k {
		out[kk] = make([]StatFunc, 0)
	}
	for nk, nv := range in {
		out[nk] = append(out[nk], nv...)
	}
	return out
}

func handleTag(s *Skelington, m map[string]int, fn func(string, map[string]int)) {
	for _, h := range s.Has {
		t := h.Unit()
		tag := strings.ToUpper(t.Value)
		fn(tag, m)
	}
}

//
func (s *stat) Get(k string) []StatFunc {
	if r, ok := s.m[k]; ok {
		return r
	}
	return nil
}

//
func (s *stat) Set(k string, fn ...StatFunc) {
	if v, ok := s.m[k]; ok {
		v = append(v, fn...)
		s.m[k] = v
	}
}

//
func (s *stat) Run(sk *Skelington, k string) error {
	var err error = nil
	if r := s.Get(k); r != nil {
		for _, fn := range r {
			err = fn(sk, s.d)
			if err != nil {
				return err
			}
			err = s.every(sk)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func (s *stat) every(sk *Skelington) error {
	var err error
	if e := s.Get("every"); e != nil {
		for _, efn := range e {
			err = efn(sk, s.d)
			if err != nil {
				return err
			}
		}
	}
	return err
}

//
func (s *stat) Report() map[string]int {
	return s.d
}

//
func (s *stat) Reported(d map[string]int) {
	s.d = d
}

//
func (s *stat) Reset() {
	for k, _ := range s.d {
		delete(s.d, k)
	}
	s.d["TOTAL"] = 0
}
