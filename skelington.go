package skelington

// Creates new Skeleton instance from provided Config
func New(cnf ...Config) (*Skeleton, error) {
	p, perr := newProcessor(cnf...)
	if perr != nil {
		return nil, perr
	}
	s := p.Process()
	return s, nil
}
