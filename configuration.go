package skelington

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
)

// A function taking a Processor instance for configuration and returning an error.
type ConfigFn func(*Processor) error

// An interface for a Config function that provides order and configuration functionality
// for a Processor instance.
type Config interface {
	Order() int
	Configure(*Processor) error
}

type config struct {
	order int
	fn    ConfigFn
}

// A default Config with arbitrary order of 50
func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

// Returns a new Config with the provided int order and ConfigFn.
func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

// A function returning the order of the Config instance for sorting.
func (c config) Order() int {
	return c.order
}

// A function providing the Processor to the Config instance.
func (c config) Configure(f *Processor) error {
	return c.fn(f)
}

type configList []Config

// configList Len, satisfying the Sort interface
func (c configList) Len() int {
	return len(c)
}

// configList Swap, satisfying the Sort interface
func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// configList Less, satisfying the Sort interface
func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

// An interface providing an abstraction of configuration functionality.
type Configuration interface {
	Add(...Config)
	Configure() error
	Configured() bool
}

type configuration struct {
	p          *Processor
	configured bool
	list       configList
}

func newConfiguration(p *Processor, conf ...Config) *configuration {
	c := &configuration{
		p:    p,
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

// Adds any number of Config to the configuration instance.
func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func configure(p *Processor, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(p)
		if err != nil {
			return err
		}
	}
	return nil
}

// Configures the configuraton instance by providing the Processor to all Config,
// returning any error immediately.
func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.p, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

// Returns a boolean indicating if configuration has been run.
func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1001, sLogger},
	config{1002, sRoot},
	config{1004, sError},
	config{1005, sAllocator},
}

func sLogger(p *Processor) error {
	if p.Logger == nil {
		l := log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
		log.Current = l
		p.Logger = l
	}
	return nil
}

// Sets the logger used by a string key. Both 'stdout' & 'text' change the logger
// from the default null to default initialized text logging. Any other string key
// will attempt to swap the logger to that key(which must've been set before hand
// in other intialization or exists as a default available logger with the logging
// package).
func SkelingtonLogger(k string) Config {
	return NewConfig(2000,
		func(p *Processor) error {
			switch k {
			case "stdout", "text":
				p.SwapFormatter(log.GetFormatter("skelington_text"))
			default:
				p.SwapFormatter(log.GetFormatter(k))
			}
			return nil
		})
}

var ConfigurationError = xrror("configuration error: %s").Out

func sRoot(p *Processor) error {
	if p.root == nil {
		return ConfigurationError("no ROOT specified")
	}
	return nil
}

// Sets the root used by skelington & allocator.
func SkelingtonRoot(path string) Config {
	return DefaultConfig(
		func(p *Processor) error {
			r := newPather("root", path)
			p.root = r
			return nil
		})
}

// Sets the file path to read a configuration from, if the allocator requires one.
func SkelingtonFile(path string) Config {
	return DefaultConfig(
		func(p *Processor) error {
			r := newPather("file", path)
			p.file = r
			return nil
		})
}

func sError(p *Processor) error {
	if p.errorHandler == Unspecified {
		p.errorHandler = ContinueOnError
	}
	return nil
}

// Sets the skelington error handling method by string key:
// one of 'continue', 'exit', or 'panic' with the default being 'continue'.
func SkelingtonError(err string) Config {
	return DefaultConfig(
		func(p *Processor) error {
			var perr ErrorHandling = ContinueOnError
			switch err {
			case "exit":
				perr = ExitOnError
			case "panic":
				perr = PanicOnError
			}
			p.errorHandler = perr
			return nil
		})
}

func sAllocator(p *Processor) error {
	if p.Allocator == nil {
		p.Allocator = Allocators.Get("emp")
	}
	return nil
}

// Sets an allocator for the skelington by string key.
func SkelingtonAllocator(k string) Config {
	return NewConfig(50,
		func(p *Processor) error {
			a := Allocators.Get(k)
			p.Allocator = a
			return nil
		})
}

// Provides any desired offset to the allocator.
func SkelingtonAllocationOffset(o string) Config {
	return NewConfig(50,
		func(p *Processor) error {
			if o != "" {
				p.offset = o
			}
			return nil
		})
}
