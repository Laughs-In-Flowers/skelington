package skelington

import (
	cr "crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Laughs-In-Flowers/xrr"
)

// A type used for switching on various error handling methods.
type ErrorHandling int

const (
	Unspecified ErrorHandling = iota
	IgnoreError
	ContinueOnError
	ExitOnError
	PanicOnError
)

// A function taking an error for specific handling.
type ErrorHandler func(error)

var openError = xrr.Xrror("unable to find or open file %s, provided %s").Out

func exist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModeDir|0755)
	}
}

func open(path string) (*os.File, error) {
	p := filepath.Clean(path)
	dir, name := filepath.Split(p)
	var fp string
	var err error
	switch dir {
	case "":
		fp, err = filepath.Abs(name)
	default:
		exist(dir)
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

// A struct for a tag containing an integer order and a string value.
type Tag struct {
	Order int
	Value string
}

// An interface for managing an abstraction of a path with a variety of dimensions.
type Pather interface {
	Key() string
	Path() string
	SetPath(string)
	GetTag() *Tag
}

type pather struct {
	key, path string
}

func newPather(key, path string) *pather {
	return &pather{key, path}
}

// The string key of the pather.
func (p *pather) Key() string {
	return p.key
}

// The string path of the pather.
func (p *pather) Path() string {
	return p.path
}

// Sets the pather path with the provided string.
func (p *pather) SetPath(path string) {
	p.path = path
}

// The tag of the pather.
func (p *pather) GetTag() *Tag {
	return &Tag{0, p.path}
}

var (
	DefaultSequencePatternString string = "([0-9A-Za-z]+)-of-([0-9A-Za-z]+)"
	DefaultSequenceNumericalFmt  string = "%d-of-%d"
)

// A struct for managing a specific point within a sequence containing integers
// for number and count.
type Sequence struct {
	Number, Count int
}

// The string value for the given sequence.
func (s *Sequence) String() string {
	return fmt.Sprintf(DefaultSequenceNumericalFmt, s.Number, s.Count)
}

// A 16 byte universally unique identifier
type UUID [16]byte

var halfbyte2hexchar = []byte{
	48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 97, 98, 99, 100, 101, 102,
}

// A string value for the given UUID.
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

func uuid() (UUID, error) {
	u := UUID{}

	_, err := cr.Read(u[:])
	if err != nil {
		return u, err
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return u, nil
}

func uuidString() string {
	u, err := uuid()
	if err != nil {
		return err.Error()
	}
	return u.String()
}
