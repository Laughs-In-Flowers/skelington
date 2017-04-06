package skelington

import (
	"github.com/Laughs-In-Flowers/log"
)

type Skeleton interface {
	log.Logger
	Allocator
	Process(int, string) *Handles
}

type skeleton struct {
	root Pather
	file Pather
	Configuration
	log.Logger
	Allocator
}

func New(cnf ...Config) (*skeleton, error) {
	s := &skeleton{}
	c := newConfiguration(s, cnf...)
	s.Configuration = c
	err := s.Configure()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func Default(logger, root, file, allocator string) (*skeleton, error) {
	lc := SkeletonLogger(logger)
	rc := SkeletonRoot(root)
	fc := SkeletonFile(file)
	ac := SkeletonAllocator(allocator)
	return New(lc, rc, fc, ac)
}

func (s *skeleton) Process(offset string) (*Handles, error) {
	s.Print("begin processing")
	if offset != "" {
		s.Printf("processing with offset %s")
	}
	path := s.file.Path()
	err := s.Open(path)
	if err != nil {
		s.Print(err)
		return nil, err
	}
	root := s.root.Tag()
	h, err := s.Allocate(root, offset)
	if err != nil {
		s.Print(err)
		return nil, err
	}
	s.Printf("finished processing")
	return h, nil
}

func init() {
	log.SetFormatter("skelington_text", log.MakeTextFormatter("skelington"))
}
