package skelington

import (
	"bytes"

	yaml "gopkg.in/yaml.v2"
)

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

func ReadFromFile(path string) (*Level, error) {
	f, err := Open(path)
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

func ReadFromDirectory(path string) (*Level, error) {
	return nil, nil
}

func notate(p *Level, ls []*Level, d int) {
	d = d + 1
	p.depth = d
	for _, child := range ls {
		child.parent = p
		notate(child, child.Levels, d)
	}
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

//func (lv *Level) Tagged() string {
//	return strings.Join(tagged(lv), ",")
//}

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

func (lv *Level) CloneMultiple(n int) []*Level {
	ret := make([]*Level, 0)
	for i := 1; i <= n; i = i + 1 {
		cloned := lv.Clone()
		ret = append(ret, cloned)
	}
	return ret
}

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

func (lv *Level) Offset(tag string) *Level {
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
