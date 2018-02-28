package skelington

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
)

//
type ConfigFn func(*Processor) error

//
type Config interface {
	Order() int
	Configure(*Processor) error
}

type config struct {
	order int
	fn    ConfigFn
}

//
func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

//
func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

//
func (c config) Order() int {
	return c.order
}

//
func (c config) Configure(f *Processor) error {
	return c.fn(f)
}

type configList []Config

func (c configList) Len() int {
	return len(c)
}

func (c configList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c configList) Less(i, j int) bool {
	return c[i].Order() < c[j].Order()
}

//
type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
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

//
func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

//
func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
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

//
func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.p, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

//
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

//
func SkeletonLogger(k string) Config {
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

var ConfigurationError = Xrror("configuration error: %s").Out

func sRoot(p *Processor) error {
	if p.root == nil {
		return ConfigurationError("no ROOT specified")
	}
	return nil
}

//
func SkeletonRoot(path string) Config {
	return DefaultConfig(
		func(p *Processor) error {
			r := newPather("root", path)
			p.root = r
			return nil
		})
}

//
func SkeletonFile(path string) Config {
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

//
func SkeletonError(err string) Config {
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

//
func SkeletonAllocator(k string) Config {
	return NewConfig(50,
		func(p *Processor) error {
			a := Allocators.Get(k)
			p.Allocator = a
			return nil
		})
}

//
func SkeletonAllocationOffset(o string) Config {
	return NewConfig(50,
		func(p *Processor) error {
			if o != "" {
				p.offset = o
			}
			return nil
		})
}
