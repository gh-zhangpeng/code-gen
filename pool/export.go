package pool

// NewPool return a new pool
func NewPool(size int) Pool {
	var p pool
	p.Init(size)
	return &p
}
