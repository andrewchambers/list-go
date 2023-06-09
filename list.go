// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package list implements a doubly linked list.
//
// To iterate over a list (where l is a *List):
//
//	for e := l.Front(); e != nil; e = e.Next() {
//		// do something with e.Value
//	}
package list

// Element is an element of a linked list.
type Element[E any] struct {
	// Next and previous pointers in the doubly-linked list of elements.
	// To simplify the implementation, internally a list l is implemented
	// as a ring, such that &l.root is both the next element of the last
	// list element (l.Back()) and the previous element of the first list
	// element (l.Front()).
	next, prev *Element[E]

	// The list to which this element belongs.
	list *List[E]

	// The value stored with this element.
	Value E
}

// Next returns the next list element or nil.
func (e *Element[E]) Next() *Element[E] {
	if p := e.next; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// Prev returns the previous list element or nil.
func (e *Element[E]) Prev() *Element[E] {
	if p := e.prev; e.list != nil && p != &e.list.root {
		return p
	}
	return nil
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List[E any] struct {
	root  Element[E]    // sentinel list element, only &root, root.prev, and root.next are used
	len   int           // current list length excluding (this) sentinel element
	epool []*Element[E] // Element pool.
}

// Init initializes or clears list l.
func (l *List[E]) Init() *List[E] {
	l.root.next = &l.root
	l.root.prev = &l.root
	l.len = 0
	l.epool = nil
	return l
}

// New returns an initialized list.
func New[E any]() *List[E] { return new(List[E]).Init() }

// Len returns the number of elements of list l.
// The complexity is O(1).
func (l *List[E]) Len() int { return l.len }

// Front returns the first element of list l or nil if the list is empty.
func (l *List[E]) Front() *Element[E] {
	if l.len == 0 {
		return nil
	}
	return l.root.next
}

// Back returns the last element of list l or nil if the list is empty.
func (l *List[E]) Back() *Element[E] {
	if l.len == 0 {
		return nil
	}
	return l.root.prev
}

// lazyInit lazily initializes a zero List value.
func (l *List[E]) lazyInit() {
	if l.root.next == nil {
		l.Init()
	}
}

// insert inserts e after at, increments l.len, and returns e.
func (l *List[E]) insert(e, at *Element[E]) *Element[E] {
	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
	e.list = l
	l.len++
	return e
}

func (l *List[E]) poolElement(e *Element[E]) {
	const poolSize = 4
	if len(l.epool) == poolSize {
		return
	}
	l.epool = append(l.epool, e)
}

func (l *List[E]) newElement() *Element[E] {
	if len(l.epool) == 0 {
		return &Element[E]{}
	}
	e := l.epool[len(l.epool)-1]
	l.epool = l.epool[:len(l.epool)-1]
	return e
}

// insertValue is a convenience wrapper for insert(&Element{Value: v}, at) with object pooling.
func (l *List[E]) insertValue(v E, at *Element[E]) *Element[E] {
	e := l.newElement()
	e.Value = v
	return l.insert(e, at)
}

// remove removes e from its list, decrements l.len
func (l *List[E]) remove(e *Element[E]) {
	e.prev.next = e.next
	e.next.prev = e.prev
	e.next = nil // avoid memory leaks
	e.prev = nil // avoid memory leaks
	e.list = nil
	l.poolElement(e)
	l.len--
}

// move moves e to next to at.
func (l *List[E]) move(e, at *Element[E]) {
	if e == at {
		return
	}
	e.prev.next = e.next
	e.next.prev = e.prev

	e.prev = at
	e.next = at.next
	e.prev.next = e
	e.next.prev = e
}

// Remove removes e from l if e is an element of list l.
// It returns the element value e.Value.
// The element must not be nil.
// You must not use e after it has been removed as it may be reused.
func (l *List[E]) Remove(e *Element[E]) any {
	if e.list == l {
		// if e.list == l, l must have been initialized when e was inserted
		// in l or l == nil (e is a zero Element) and l.remove will crash
		l.remove(e)
	}
	return e.Value
}

// PushFront inserts a new element e with value v at the front of list l and returns e.
func (l *List[E]) PushFront(v E) *Element[E] {
	l.lazyInit()
	return l.insertValue(v, &l.root)
}

// PushBack inserts a new element e with value v at the back of list l and returns e.
func (l *List[E]) PushBack(v E) *Element[E] {
	l.lazyInit()
	return l.insertValue(v, l.root.prev)
}

// InsertBefore inserts a new element e with value v immediately before mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[E]) InsertBefore(v E, mark *Element[E]) *Element[E] {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark.prev)
}

// InsertAfter inserts a new element e with value v immediately after mark and returns e.
// If mark is not an element of l, the list is not modified.
// The mark must not be nil.
func (l *List[E]) InsertAfter(v E, mark *Element[E]) *Element[E] {
	if mark.list != l {
		return nil
	}
	// see comment in List.Remove about initialization of l
	return l.insertValue(v, mark)
}

// MoveToFront moves element e to the front of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[E]) MoveToFront(e *Element[E]) {
	if e.list != l || l.root.next == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, &l.root)
}

// MoveToBack moves element e to the back of list l.
// If e is not an element of l, the list is not modified.
// The element must not be nil.
func (l *List[E]) MoveToBack(e *Element[E]) {
	if e.list != l || l.root.prev == e {
		return
	}
	// see comment in List.Remove about initialization of l
	l.move(e, l.root.prev)
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[E]) MoveBefore(e, mark *Element[E]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark.prev)
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of l, or e == mark, the list is not modified.
// The element and mark must not be nil.
func (l *List[E]) MoveAfter(e, mark *Element[E]) {
	if e.list != l || e == mark || mark.list != l {
		return
	}
	l.move(e, mark)
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[E]) PushBackList(other *List[E]) {
	l.lazyInit()
	for i, e := other.Len(), other.Front(); i > 0; i, e = i-1, e.Next() {
		l.insertValue(e.Value, l.root.prev)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[E]) PushFrontList(other *List[E]) {
	l.lazyInit()
	for i, e := other.Len(), other.Back(); i > 0; i, e = i-1, e.Prev() {
		l.insertValue(e.Value, &l.root)
	}
}
