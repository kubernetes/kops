package ipam

type prime struct {
	resultChan chan<- struct{}
}

func (c *prime) Try(alloc *Allocator) bool {
	if !alloc.ring.Empty() {
		close(c.resultChan)
		return true
	}

	alloc.establishRing()

	return false
}

func (c *prime) Cancel() {
	close(c.resultChan)
}

func (c *prime) ForContainer(ident string) bool {
	return false
}
