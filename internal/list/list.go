package list

type list struct {
	Min, Max   int
	Prev, Next *list
}

func insert(l *list, v int) *list {
	li := l
	for li != nil {
		if li.contains(v) {
			return l
		}
		if v < li.Min {
			if li.prevGapContains(v) {
				li.insertBefore(v)
				return l
			}
			li = li.Prev
			continue
		}
		if v > li.Max {
			if li.nextGapContains(v) {
				li.insertAfter(v)
				return l
			}
			li = li.Next
			continue
		}
	}
	return &list{Min: v, Max: v}
}

func (l *list) contains(v int) bool {
	return l.Min <= v && l.Max >= v
}

func (l *list) prevGapContains(v int) bool {
	if l.Prev != nil {
		return l.Prev.Max < v && v < l.Min
	}
	return v < l.Min
}

func (l *list) nextGapContains(v int) bool {
	if l.Next != nil {
		return l.Max < v && v < l.Next.Min
	}
	return l.Max < v
}

func merge(li *list) {
	cur := li
	for cur.Next != nil && cur.Max+1 == cur.Next.Min {
		mergeList(cur, cur.Next)
	}
	for cur.Prev != nil && cur.Min-1 == cur.Prev.Max {
		mergeList(cur.Prev, cur)
	}
}

func mergeList(cur *list, next *list) {
	if cur != nil && next != nil {
		if cur.Max+1 == next.Min {
			cur.Max = next.Max
			if next.Next != nil {
				next.Next.Prev = cur
			}
			cur.Next = next.Next
		}
	}

}

func insertBetween(prev *list, next *list, li *list) {
	li.Next = next
	li.Prev = prev
	if prev != nil {
		prev.Next = li
	}
	if next != nil {
		next.Prev = li
	}

}
func (l *list) insertBefore(v int) {
	li := &list{
		Min: v,
		Max: v,
	}
	insertBetween(l.Prev, l, li)
	merge(li)
}
func (l *list) insertAfter(v int) {
	li := &list{
		Min: v,
		Max: v,
	}
	insertBetween(l, l.Next, li)
	merge(li)
}
