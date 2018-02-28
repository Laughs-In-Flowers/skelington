package skelington

import (
	"fmt"
	"os"

	"github.com/Laughs-In-Flowers/log"
)

// A core processing type to generate a Skeleton instance.
type Processor struct {
	root         Pather
	file         Pather
	errorHandler ErrorHandling
	offset       string
	Configuration
	log.Logger
	Allocator
}

func emptyProcessor() *Processor {
	s := &Processor{}
	c := newConfiguration(s)
	s.Configuration = c
	return s
}

func newProcessor(cnf ...Config) (*Processor, error) {
	s := emptyProcessor()
	s.Add(cnf...)
	err := s.Configure()
	if err != nil {
		return nil, err
	}
	return s, nil
}

//
func (p *Processor) Process() *Skeleton {
	p.Print("begin processing")
	if p.offset != "" {
		p.Printf("processing with offset %s")
	}

	var ret *Skeleton
	ret = p.Allocate(p.file, p.root, p.offset, p.manageError)

	p.Printf("finished processing")
	return ret
}

func (p *Processor) manageError(e error) {
	if e != nil {
		switch p.errorHandler {
		case ContinueOnError:
			p.Print(e)
		case ExitOnError:
			fmt.Fprintf(os.Stderr, "FATAL: %s\n", e)
			os.Exit(-1)
		case PanicOnError:
			panic(e)
		}
	}
}

func init() {
	log.SetFormatter("skelington_text", log.MakeTextFormatter("skelington"))
}
