package safe

import "testing"

// go test -v ./src/common/safe/

func Test_LazyFree0(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 0)
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Get()
	q.Get()
	for i := 0; i < 3; i++ {
		if q.writer[i] == nil {
			t.Errorf("data unexpectedly freed")
		}
	}
}
func Test_LazyFree1(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 1)
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Get()
	q.Get()
	for i := 0; i < 3; i++ {
		if q.writer[i] != nil {
			t.Errorf("data unexpectedly not freed")
		}
	}
}
func Test_LazyFree6(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 6)
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Put(make([]byte, 16))
	q.Get()
	q.Get()
	for i := 0; i < 3; i++ {
		if q.writer[i] == nil {
			t.Errorf("data unexpectedly freed")
		}
	}
	q.Get()
	q.Get()
	for i := 0; i < 3; i++ {
		if q.writer[i] == nil {
			t.Errorf("data unexpectedly freed")
		}
	}
	q.Get()
	q.Get()
	for i := 0; i < 3; i++ {
		if q.writer[i] != nil {
			t.Errorf("data not freed at the expected freeCycle")
		}
	}
}
func Test_MultiQueue(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 0)
	for i := 0; i < 5; i++ {
		ok, stopped := q.Put(233)
		if !ok || stopped {
			t.Errorf("failed to add new entry")
		}
		if q.stop {
			t.Errorf("stopped too early")
		}
	}
	v := q.Get()
	if len(v) != 5 {
		t.Errorf("failed to get all entries")
	}
	ok, stopped := q.Put(234)
	if ok {
		t.Errorf("entry added to paused queue")
	}
	if stopped {
		t.Errorf("entry queue unexpectedly stopped")
	}
}
func Test_MultiQueue_Close(t *testing.T) {
	{
		q := MultiQueue{}
		q.Init(5, 0)
		if q.stop {
			t.Errorf("entry queue is stopped by default")
		}
		for i := 0; i < 5; i++ {
			ok, stopped := q.Put(233)
			if !ok || stopped {
				t.Errorf("failed to add new entry")
			}
			if q.stop {
				t.Errorf("stopped too early")
			}
		}
		ok, _ := q.Put(234)
		if ok {
			t.Errorf("not expect to add more")
		}
	}
	{
		q := MultiQueue{}
		q.Init(5, 0)
		q.Close()
		if !q.stop {
			t.Errorf("entry queue is not marked as stopped")
		}
		if q.wpos != 0 {
			t.Errorf("wpos %d, want 0", q.wpos)
		}
		ok, stopped := q.Put(235)
		if ok {
			t.Errorf("not expect to add more")
		}
		if !stopped {
			t.Errorf("stopped flag is not returned")
		}
	}
}
func Test_MultiQueue_Add(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 0)
	for i := uint32(0); i < 5; i++ {
		if ok, stop := q.Put(uint64(i + 1)); !ok || stop {
			t.Errorf("failed to add new entry")
		}
		if q.wpos != i+1 {
			t.Errorf("wpos %d, want %d", q.wpos, i+1)
		}
		if q.writer[i].(uint64) != uint64(i+1) {
			t.Errorf("index %d, want %d", q.writer[i].(uint64), uint64(i+1))
		}
	}
}
func Test_MultiQueue_Get(t *testing.T) {
	q := MultiQueue{}
	q.Init(5, 0)
	for i := 0; i < 3; i++ {
		if ok, stop := q.Put(uint64(i + 1)); !ok || stop {
			t.Errorf("failed to add new entry")
		}
	}
	r := q.Get()
	if len(r) != 3 {
		t.Errorf("len %d, want %d", len(r), 3)
	}
	if q.wpos != 0 {
		t.Errorf("wpos %d, want %d", q.wpos, 0)
	}
	// check whether we can keep adding entries as long as we keep getting
	// previously written entries.
	{
		expectedIndex := uint64(1)
		q := MultiQueue{}
		q.Init(5, 0)
		for i := 0; i < 1000; i++ {
			ok, stopped := q.Put(uint64(i + 1))
			if !ok || stopped {
				t.Errorf("failed to add new entry")
			}
			if q.wpos == uint32(len(q.writer)) {
				r := q.Get()
				if len(r) != 5 {
					t.Errorf("len %d, want %d", len(r), 5)
				}
				for _, e := range r {
					if e.(uint64) != expectedIndex {
						t.Errorf("index %d, expected %d", e.(uint64), expectedIndex)
					}
					expectedIndex++
				}
			}
		}
	}
}
