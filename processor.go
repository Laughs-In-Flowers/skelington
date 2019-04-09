package skelington

import (
	"fmt"
	"os"
)

// A core processing type gathering everything required for Skelington generation.
type Processor struct {
	root         Pather
	file         Pather
	errorHandler ErrorHandling
	offset       string
	hookHolder   map[HookTiming][]SkelingtonHook
	statHolder   map[string][]StatFunc
	Allocator
}

func newProcessor(cnf ...Config) (*Processor, error) {
	p := &Processor{
		hookHolder: make(map[HookTiming][]SkelingtonHook),
		statHolder: make(map[string][]StatFunc),
	}
	c := newConfiguration(p)
	c.Add(cnf...)
	err := c.Configure()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// The core function that produces a Skelington instance from an allocation strategy.
func (p *Processor) Process() *Skelington {
	s := newSkelington(p.hookHolder, p.statHolder)
	ret := p.Allocate(s, p.file, p.root, p.offset, p.manageError)
	return ret
}

func (p *Processor) manageError(e error) {
	if e != nil {
		switch p.errorHandler {
		case ContinueOnError:
			fmt.Fprintf(os.Stdout, "%s\n", e)
		case ExitOnError:
			fmt.Fprintf(os.Stdout, "FATAL: %s\n", e)
			os.Exit(-1)
		case PanicOnError:
			panic(e)
		}
	}
}

type hookHold struct {
	when HookTiming
	fn   SkelingtonHook
}
