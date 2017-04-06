package skelington

import (
	"fmt"
	"os"
	"path/filepath"
)

type xrror struct {
	base string
	vals []interface{}
}

func (x *xrror) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(x.base, x.vals...))
}

func (x *xrror) Out(vals ...interface{}) *xrror {
	x.vals = vals
	return x
}

func Xrror(base string) *xrror {
	return &xrror{base: base}
}

var openError = Xrror("unable to find or open file %s, provided %s").Out

func Exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

func Open(path string) (*os.File, error) {
	p := filepath.Clean(path)
	dir, name := filepath.Split(p)
	var fp string
	var err error
	switch dir {
	case "":
		fp, err = filepath.Abs(name)
	default:
		Exist(dir)
		fp, err = filepath.Abs(p)
	}

	if err != nil {
		return nil, err
	}

	if file, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE, 0660); err == nil {
		return file, nil
	}

	return nil, openError(fp, path)
}

type Tag struct {
	Order int
	Value string
}

type Pather interface {
	Key() string
	Path() string
	SetPath(string)
	Tag() *Tag
}

type pather struct {
	key, path string
}

func newPather(key, path string) *pather {
	return &pather{key, path}
}

func (p *pather) Key() string {
	return p.key
}

func (p *pather) Path() string {
	return p.path
}

func (p *pather) SetPath(path string) {
	p.path = path
}

func (p *pather) Tag() *Tag {
	return &Tag{0, p.path}
}

type Sequence struct {
	number, count int
}
