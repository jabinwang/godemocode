package workerpool

const minCapacity = 16


type Deque interface {
	 Front() interface{}
	 Back() interface{}
	 PushBack(elem interface{})
	 PushFront(elem interface{})
	 PopFront() interface{}
	 PopBack() interface{}
	 Clear()
	 Len() int
}

type deque struct {
	buf []interface{}
	head int
	tail int
	count int
	minCap int
}

func New(size int) Deque {
	minCap := minCapacity
	for minCap < size {
		minCap <<= 1
	}
	var buf []interface{}
	buf = make([]interface{}, minCap)
	q := &deque{
		buf: buf,
	}
	return q
}

func (q *deque) Len() int {
	return q.count
}

func (q *deque) Front() interface{} {
	if q.count <= 0{
		panic("deque: Front() called when empty")
	}
	return q.buf[q.head]
}

func (q *deque) Back() interface{} {
	if q.count <= 0 {
		panic("deque: Back() called when empty")
	}
	return q.buf[q.prev(q.tail)]
}

func (q *deque) PushBack(elem interface{}) {
	q.growIfFull()
	q.buf[q.tail] = elem
	q.tail = q.next(q.tail)
	q.count++
}

func (q *deque) PushFront(elem interface{}) {
	q.growIfFull()
	q.head = q.prev(q.head)
	q.buf[q.head] = elem
	q.count++
}

func (q *deque) PopFront() interface{} {
	if q.count <=0 {
		panic("deque PopFront() called on empty queue")
	}
	ret := q.buf[q.head]
	q.buf[q.head] = nil
	q.head = q.next(q.head)
	q.count--
	q.shrinkIfExcess()
	return ret
}

func (q *deque) PopBack() interface{} {
	if q.count <= 0 {
		panic("deque: PopBack() called on empty queue")
	}
	q.tail = q.prev(q.tail)
	ret := q.buf[q.tail]
	q.buf[q.tail] = nil
	q.count--
	q.shrinkIfExcess()
	return ret
}

func (q *deque) growIfFull() {
	if q.count != len(q.buf) {
		return
	}
	if len(q.buf) == 0 {
		if q.minCap == 0 {
			q.minCap = minCapacity
		}
		q.buf = make([]interface{}, q.minCap)
		return
	}
	q.resize()
}

func (q *deque) resize() {
	newBuf := make([]interface{}, q.count << 1)
	if q.tail > q.head {
		copy(newBuf, q.buf[q.head:q.tail])
	} else {
		n := copy(newBuf, q.buf[q.head:])
		copy(newBuf[n:], q.buf[:q.tail])
	}
	q.head = 0
	q.tail = q.count
	q.buf = newBuf
}

func (q *deque) Clear() {
	modBits := len(q.buf) - 1
	for h:= q.head; h != q.tail; h = (h+1) & modBits {
		q.buf[h] = nil
	}
	q.head = 0
	q.tail = 0
	q.count = 0
}

func (q *deque) prev(i int) int {
	return (i-1) &(len(q.buf) -1)
}
func (q *deque) next(i int) int {
	return (i+1) &(len(q.buf) - 1)
}

func (q *deque) shrinkIfExcess() {
	if len(q.buf) > q.minCap && (q.count << 2) == len(q.buf) {
		q.resize()
	}
}


