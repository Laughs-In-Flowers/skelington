package skelington

import (
	"os"
	"sort"

	"github.com/Laughs-In-Flowers/log"
)

type ConfigFn func(*skeleton) error

type Config interface {
	Order() int
	Configure(*skeleton) error
}

type config struct {
	order int
	fn    ConfigFn
}

func DefaultConfig(fn ConfigFn) Config {
	return config{50, fn}
}

func NewConfig(order int, fn ConfigFn) Config {
	return config{order, fn}
}

func (c config) Order() int {
	return c.order
}

func (c config) Configure(f *skeleton) error {
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

type Configuration interface {
	Add(...Config)
	AddFn(...ConfigFn)
	Configure() error
	Configured() bool
}

type configuration struct {
	s          *skeleton
	configured bool
	list       configList
}

func newConfiguration(s *skeleton, conf ...Config) *configuration {
	c := &configuration{
		s:    s,
		list: builtIns,
	}
	c.Add(conf...)
	return c
}

func (c *configuration) Add(conf ...Config) {
	c.list = append(c.list, conf...)
}

func (c *configuration) AddFn(fns ...ConfigFn) {
	for _, fn := range fns {
		c.list = append(c.list, DefaultConfig(fn))
	}
}

func configure(s *skeleton, conf ...Config) error {
	for _, c := range conf {
		err := c.Configure(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *configuration) Configure() error {
	sort.Sort(c.list)

	err := configure(c.s, c.list...)
	if err == nil {
		c.configured = true
	}

	return err
}

func (c *configuration) Configured() bool {
	return c.configured
}

var builtIns = []Config{
	config{1001, sLogger},
	config{1002, sRoot},
	config{1003, sFile},
	config{1004, sAllocator},
}

func sLogger(s *skeleton) error {
	if s.Logger == nil {
		l := log.New(os.Stdout, log.LInfo, log.DefaultNullFormatter())
		log.Current = l
		s.Logger = l
	}
	return nil
}

func SkeletonLogger(k string) Config {
	return NewConfig(2000,
		func(s *skeleton) error {
			switch k {
			case "stdout", "text":
				s.SwapFormatter(log.GetFormatter("skelington_text"))
			}
			return nil
		})
}

var RooterConfigurationError = Xrror("No %s is set for this skeleton instance.").Out

func sRoot(s *skeleton) error {
	if s.root == nil {
		return RooterConfigurationError("ROOT")
	}
	return nil
}

func SkeletonRoot(path string) Config {
	return NewConfig(50,
		func(s *skeleton) error {
			r := newPather("root", path)
			s.root = r
			return nil
		})
}

func sFile(s *skeleton) error {
	if s.file == nil {
		return RooterConfigurationError("FILE")
	}
	return nil
}

func SkeletonFile(path string) Config {
	return NewConfig(50,
		func(s *skeleton) error {
			r := newPather("file", path)
			s.file = r
			return nil
		})
}

var AllocatorConfigurationError = Xrror("No allocator is set for this skeleton instance.")

func sAllocator(s *skeleton) error {
	if s.Allocator == nil {
		return AllocatorConfigurationError
	}
	return nil
}

func SkeletonAllocator(k string) Config {
	return NewConfig(50,
		func(s *skeleton) error {
			a := Allocators.Get(k)
			s.Allocator = a
			return nil
		})
}

func logConfiguration(l log.Logger) {
	//logger
	//root
	//file
	//allocator
}
