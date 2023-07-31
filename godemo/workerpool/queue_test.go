package workerpool

import "testing"

func TestSimple(t *testing.T)  {
	q  := New(16)
	for i:= 0; i < minCapacity; i++ {
		q.PushBack(i)
	}

	for i:= 0; i < minCapacity; i++ {
		if q.Front() != i {
			t.Error("peek", i, "real value", q.Front())
		}
		x := q.PopFront()
		if x != i {
			t.Error("remove", i, "real value", x)
		}
	}
		q.Clear()
		for i := 0; i < minCapacity; i++ {
			q.PushFront(i)
		}
		for i := minCapacity - 1; i >= 0; i-- {
			x := q.PopFront()
			if x != i {
				t.Error("remove", i, "had value", x)
			}
		}

}