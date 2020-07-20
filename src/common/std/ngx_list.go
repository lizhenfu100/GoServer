package std

type List struct {
	prev *List
	next *List
}

func (h *List) Init()       { h.prev, h.next = h, h } //第一个作哨兵
func (h *List) Empty() bool { return h == h.prev }
func (h *List) Head() *List { return h.next }
func (h *List) Last() *List { return h.prev }

func (h *List) Insert(x *List) {
	x.next = h.next
	x.next.prev = x
	x.prev = h
	h.next = x
}
func (h *List) InsertPrev(x *List) {
	x.prev = h.prev
	x.prev.next = x
	x.next = h
	h.prev = x
}
func (x *List) Remove() {
	x.next.prev = x.prev
	x.prev.next = x.next
}
func (h *List) Split(q, n *List) { //q = n.Head
	n.prev = h.prev
	n.prev.next = n
	n.next = q
	h.prev = q.prev
	h.prev.next = h
	q.prev = n
}
func (h *List) Add(n *List) {
	h.prev.next = n.next
	n.next.prev = h.prev
	h.prev = n.prev
	h.prev.next = h
}
func (h *List) Middle() *List {
	var middle *List
	if middle = h.Head(); middle == h.Last() {
		return middle
	}
	for next := h.Head(); ; {
		middle = middle.next
		if next = next.next; next == h.Last() {
			return middle
		}
		if next = next.next; next == h.Last() {
			return middle
		}
	}
}
func (h *List) Sort(cmp func(r, l *List) int) {
	var q, prev, next *List
	if q = h.Head(); q == h.Last() {
		return
	}
	for q = q.next; q != h; q = next {
		prev = q.prev
		next = q.next
		q.Remove()
		for {
			if cmp(prev, q) <= 0 {
				break
			}
			if prev = prev.prev; prev == h {
				break
			}
		}
		prev.Insert(q)
	}
}
