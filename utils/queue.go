package utils

type Queue struct {
	len        int
	head, tail int
	q          [][]byte
}

func (q Queue) New(n int) *Queue {
	return &Queue{n, 0, 0, make([][]byte, n)}
}

func (p *Queue) Enqueue(obj []byte) bool {
	p.q = append(p.q, obj)
	return true
}

func (p *Queue) Dequeue() bool {
	p.q = p.q[1:]
	return true
}

func (p *Queue) IsEmpty() bool {

	return len(p.q) == 0
}

func (p *Queue) Items() [][]byte {

	return p.q
}
