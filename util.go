package skelington

import (
	cr "crypto/rand"
	"fmt"
	"os"
	"path/filepath"
)

//
type ErrorHandling int

const (
	Unspecified ErrorHandling = iota
	IgnoreError
	ContinueOnError
	ExitOnError
	PanicOnError
)

//
type ErrorHandler func(error)

type xrror struct {
	base string
	vals []interface{}
}

//
func (x *xrror) Error() string {
	return fmt.Sprintf("%s", fmt.Sprintf(x.base, x.vals...))
}

//
func (x *xrror) Out(vals ...interface{}) *xrror {
	x.vals = vals
	return x
}

//
func Xrror(base string) *xrror {
	return &xrror{base: base}
}

var openError = Xrror("unable to find or open file %s, provided %s").Out

//
func Exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

//
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

//
type Tag struct {
	Order int
	Value string
}

//
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

//
func (p *pather) Key() string {
	return p.key
}

//
func (p *pather) Path() string {
	return p.path
}

//
func (p *pather) SetPath(path string) {
	p.path = path
}

//
func (p *pather) Tag() *Tag {
	return &Tag{0, p.path}
}

//
type Sequence struct {
	number, count int
}

//
func (s *Sequence) String() string {
	return fmt.Sprintf("%d-of-%d", s.number, s.count)
}

//
type UUID [16]byte

var halfbyte2hexchar = []byte{
	48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 97, 98, 99, 100, 101, 102,
}

//
func (u UUID) String() string {
	b := [36]byte{}

	for i, n := range []int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34,
	} {
		b[n] = halfbyte2hexchar[(u[i]>>4)&0x0f]
		b[n+1] = halfbyte2hexchar[u[i]&0x0f]
	}

	b[8] = '-'
	b[13] = '-'
	b[18] = '-'
	b[23] = '-'

	return string(b[:])
}

//
func V4() (UUID, error) {
	u := UUID{}

	_, err := cr.Read(u[:])
	if err != nil {
		return u, err
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return u, nil
}

//
func V4Quick() string {
	u, err := V4()
	if err != nil {
		return err.Error()
	}
	return u.String()
}
