package skelington

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"

	yaml "gopkg.in/yaml.v2"
)

// A recursive structure used as a tool for exploring and creating handles.
type Level struct {
	parent   *Level
	depth    int
	sequence *Sequence
	Tag      string
	Leaf     bool
	Relative bool
	Number   int
	Percent  float64
	Actual   int
	Levels   []*Level
}

func emptyLevel(tag string) *Level {
	return &Level{
		Tag:    tag,
		Levels: make([]*Level, 0),
	}
}

// Provided a path string, will attempt to read a yaml file there, creating a new
// Level instance and an error.
func ReadFromFile(path string) (*Level, error) {
	var f *os.File
	var err error
	f, err = open(path)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer(make([]byte, 0, bytes.MinRead))
	_, err = b.ReadFrom(f)
	if err != nil {
		return nil, err
	}
	raw := b.Bytes()
	lv := &Level{}
	err = yaml.Unmarshal(raw, lv)
	if err != nil {
		return nil, err
	}
	notate(lv, lv.Levels, 0)
	return lv, nil
}

func notate(p *Level, ls []*Level, d int) {
	d = d + 1
	p.depth = d
	for _, child := range ls {
		child.parent = p
		notate(child, child.Levels, d)
	}
}

// Provided a path string, will attempt to read from that directory to create a
// new Level instance and an error.
func ReadFromDirectory(path string, offset string) (*Level, error) {
	var seq *regexp.Regexp
	var err error
	seq, err = regexp.Compile(offset)
	if err != nil {
		return nil, err
	}
	lv := emptyLevel(path)
	lv, err = walk(lv, path, seq)
	if err != nil {
		return nil, err
	}
	notate(lv, lv.Levels, 0)
	return lv, nil
}

func walk(lv *Level, path string, seq *regexp.Regexp) (*Level, error) {
	flat := make(map[string]*Level)
	flat[filepath.Base(path)] = lv

	resErr := filepath.Walk(path, func(p string, f os.FileInfo, e error) error {
		var base string = filepath.Base(p)
		var err error
		var fl *os.File
		fl, err = os.Open(p)
		if err != nil {
			return err
		}
		var dirs []string
		dirs, err = fl.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, dir := range dirs {
			switch {
			case seq.MatchString(dir):
				if par, ok := flat[base]; ok {
					par.Number = par.Number + 1
				}
			default:
				clv := emptyLevel(dir)
				if par, ok := flat[base]; ok {
					clv.parent = par
					par.Levels = append(par.Levels, clv)
				}
				flat[dir] = clv
			}
		}
		return err
	})

	return lv, resErr
}

func reverse(in []string) []string {
	for i, j := 0, len(in)-1; i < j; i, j = i+1, j-1 {
		in[i], in[j] = in[j], in[i]
	}
	return in
}

func gather(lv *Level, tags []string) []string {
	tags = append(tags, lv.Tag)
	if lv.parent != nil {
		tags = gather(lv.parent, tags)
	}
	return tags
}

func tagged(lv *Level) []string {
	return reverse(gather(lv, []string{}))
}

// Return an array of Tags as the Level family.
func (lv *Level) Family() []*Tag {
	ret := make([]*Tag, 0)
	tgd := tagged(lv)
	ll := len(tgd)
	for k, v := range tgd {
		if k != ll-1 {
			t := &Tag{k + 1, v}
			ret = append(ret, t)
		}
	}
	return ret
}

// Return the Level unit Tag.
func (lv *Level) Unit() *Tag {
	t := &Tag{}
	tgd := tagged(lv)
	ll := len(tgd)
	for k, v := range tgd {
		if k == ll-1 {
			t.Order = k + 1
			t.Value = v
		}
	}
	return t
}

// Returns a clone of the Level.
func (lv *Level) Clone() *Level {
	parent := lv.parent
	var children []*Level
	for _, v := range lv.Levels {
		children = append(children, v.Clone())
	}
	nl := *lv
	ret := &nl
	ret.parent = parent
	ret.Levels = children
	return ret
}

// Clone the Level to the provided number.
func (lv *Level) CloneMultiple(n int) []*Level {
	ret := make([]*Level, 0)
	for i := 1; i <= n; i = i + 1 {
		cloned := lv.Clone()
		ret = append(ret, cloned)
	}
	return ret
}

// Apply the provided function to the Level, and propagate to all child Level.
func (lv *Level) Iter(fn func(*Level)) {
	fn(lv)
	for _, l := range lv.Levels {
		l.Iter(fn)
	}
}

func branch(lv *Level) {
	for _, v := range lv.Levels {
		if isLeaf(v) {
			add := v.CloneMultiple(v.Number - 1)
			lv.Levels = append(lv.Levels, add...)
		}
	}
}

func flatten(lv *Level) []*Level {
	f := &flat{}
	lv.Iter(f.flatten)
	return f.has
}

type flat struct {
	has []*Level
}

func (f *flat) flatten(l *Level) {
	if isLeaf(l) {
		f.has = append(f.has, l)
	}
}

func offset(lv *Level, tag string) *Level {
	var ret *Level
	fn := func(ll *Level) {
		if ll.Tag == tag {
			ret = ll
		}
	}
	lv.Iter(fn)
	return ret
}

func isLeaf(lv *Level) bool {
	if lv.Leaf {
		return true
	}
	return isAbsoluteLeaf(lv)
}

func isAbsoluteLeaf(lv *Level) bool {
	if len(lv.Levels) > 0 {
		return false
	}
	return true
}
