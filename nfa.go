// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*

Package fsm provides some utilities for messing with finite state machines.

Dead DFA State

"A state is a dead state if it is not an accepting state and has no out-going
transitions except to itself."[7]

Passing withDeadState == true to some methods in this package makes the
produced DFAs "complete". For many practical purposes the dead state is not
needed and all the additional edges to it are only a waste of memory.

Note: Negative symbol values are reserved for internal purposes.

TODO

Implement ε edges having other than the default priority (Epsilon == -1). This
is needed for regexp based recognizers/tokenizers like golex[8].

Links

Referenced from elsewhere:

  [1]: http://en.wikipedia.org/wiki/Finite-state_machine
  [2]: http://en.wikipedia.org/wiki/Nondeterministic_finite_automaton
  [3]: http://en.wikipedia.org/wiki/Powerset_construction
  [4]: http://en.wikipedia.org/wiki/Nondeterministic_finite_automaton_with_%CE%B5-moves
  [5]: http://en.wikipedia.org/wiki/DFA_minimization
  [6]: http://en.wikipedia.org/wiki/Janusz_Brzozowski_%28computer_scientist%29
  [7]: http://www.cs.odu.edu/~toida/nerzic/390teched/regular/fa/min-fa.html
  [8]: http://godoc.org/github.com/cznic/golex

*/
package fsm

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cznic/mathutil"
	"github.com/cznic/strutil"
)

// Epsilon is a symbol value representing an ε edge (with no priority).
const Epsilon = -1

type closure map[*State]struct{}

func (c closure) id() string {
	a := make([]int, 0, len(c))
	for s := range c {
		a = append(a, s.Id())
	}
	sort.Ints(a)
	return fmt.Sprint(a)
}

// Exclude removes state s from the closure.
func (c closure) Exclude(s *State) {
	delete(c, s)
}

// Has returns whether s is in the closure.
func (c closure) Has(s *State) (ok bool) {
	_, ok = c[s]
	return
}

// Include adds s to the closure.
func (c closure) Include(s *State) {
	c[s] = struct{}{}
}

// List returns a slice of all states in the closure.
func (c closure) List() (r []*State) {
	r = make([]*State, len(c))
	i := 0
	for state := range c {
		r[i] = state
		i++
	}
	return
}

// -------------------------------------------------------------------- Closure

// Closure is a set of states.
type Closure struct {
	closure
}

// NewClosure returns a newly created Closure.
func NewClosure() Closure {
	return Closure{closure{}}
}

// ------------------------------------------------------------------------ NFA

// NFA is a nondeterministic finite automaton [2].
type NFA struct {
	s2i   map[*State]int
	i2s   map[int]*State
	start *State
}

// NewNFA return a new, empty NFA.
func NewNFA() *NFA {
	return &NFA{s2i: map[*State]int{}, i2s: map[int]*State{}}
}

func (n *NFA) id(s *State) int {
	if id, ok := n.s2i[s]; ok {
		return id
	}

	i := n.Len()
	n.s2i[s] = i
	n.i2s[i] = s
	return i
}

// Len returns the number of NFA's states.
func (n *NFA) Len() int {
	return len(n.s2i)
}

// List returns a slice of all NFA's states.
func (n *NFA) List() (r []*State) {
	r = make([]*State, n.Len())
	for i, state := range n.i2s {
		r[i] = state
	}
	return
}

// MinimalDFA returns the NFA converted to a minimal DFA[5]. Dead state is
// possibly constructed if withDeadState == true.
//
// Note: Algorithm used is Brzozowski[6].
func (n *NFA) MinimalDFA(withDeadState bool) *NFA {
	return n.Reverse().Powerset(withDeadState).Reverse().Powerset(withDeadState)
}

// NewState returns a new state added to the NFA. If the NFA was empty, the new
// state becomes the start state.
func (n *NFA) NewState() *State {
	s := &State{nfa: n}
	if n.Len() == 0 {
		n.start = s
	}
	s.Id()
	return s
}

// Powerset converts[3] the NFA into a NFA without ε edges, ie. into a DFA.
// Dead state is possibly constructed if withDeadState == true.
func (n *NFA) Powerset(withDeadState bool) (out *NFA) {
	alphabetSize := 0
	out = NewNFA()
	closures := map[string]*State{}
	var f func(closure) *State

	f = func(c closure) (result *State) {
		cid := c.id()
		if s, ok := closures[cid]; ok {
			return s
		}

		result = out.NewState()
		closures[cid] = result
		transitions := transitions{}
		for cset := range c {
			result.IsAccepting = result.IsAccepting || cset.IsAccepting
			for sym, nextStates := range cset.transitions() {
				if sym < 0 { //TODO(later) implement priorities
					continue
				}

				alphabetSize = mathutil.Max(alphabetSize, sym+1)
				for nextState := range nextStates {
					for nextState := range nextState.closure() {
						transitions.newEdge(sym, true, nextState)
					}
				}
			}
		}
		for sym, closure := range transitions {
			result.NewEdge(sym, f(closure))
		}

		return
	}

	out.start = f(n.Start().closure())
	var dead *State
	if withDeadState {
		for state := range out.s2i {
			edges := state.transitions()
			for sym := 0; sym < alphabetSize; sym++ {
				if _, ok := edges[sym]; !ok {
					if dead == nil {
						dead = out.NewState()
					}
					state.NewEdge(sym, dead)
				}
			}
		}
		if dead != nil {
			for sym := 0; sym < alphabetSize; sym++ {
				dead.NewEdge(sym, dead)
			}
		}
	}
	return
}

// Reverse returns a NFA for the reverse language accepted by n.
func (n *NFA) Reverse() (out *NFA) {
	out = NewNFA()
	a := make([]*State, n.Len())
	for i := range a {
		a[i] = out.NewState()
	}

	var acceptingIds []int
	for idFrom := 0; idFrom < n.Len(); idFrom++ {
		state := n.State(idFrom)
		if state.IsAccepting {
			acceptingIds = append(acceptingIds, idFrom)
		}
		for sym, tos := range n.State(idFrom).edges {
			for to := range tos {
				a[to.Id()].NewEdge(sym, a[idFrom])
			}
		}
	}

	a[n.start.Id()].IsAccepting = true
	switch len(acceptingIds) {
	case 1:
		out.start = a[acceptingIds[0]]
	default:
		out.start = out.NewState()
		for _, id := range acceptingIds {
			out.start.NewEdge(Epsilon, a[id])
		}
	}
	return
}

// SetStart sets the NFA's start state. Passing a state from a different NFA
// will panic.
func (n *NFA) SetStart(s *State) {
	if s.nfa != n {
		panic(s)
	}

	n.start = s
}

// Start returns the NFA's start state.
func (n *NFA) Start() *State {
	return n.start
}

// State returns the NFA's state with Id() == id or nil if no such state exists.
func (n *NFA) State(id int) *State {
	return n.i2s[id]
}

// String implements fmt.Stringer for debugging, etc.
func (n *NFA) String() string {
	var b bytes.Buffer
	for i := 0; i < n.Len(); i++ {
		b.WriteString(n.i2s[i].String())
	}
	return b.String()
}

// ---------------------------------------------------------------------- State

// State is one of the NFA states.
type State struct {
	nfa         *NFA
	IsAccepting bool // Whether this state is an accepting one.
	edges       transitions
}

// Closure returns a state set consisting of s and all states reachable from s
// through ε edges, transitively.
func (s *State) Closure() (c Closure) {
	return Closure{s.closure()}
}

func (s *State) closure() (c closure) {
	c = closure{}
	var f func(*State)
	f = func(s *State) {
		if _, ok := c[s]; ok {
			return
		}

		c[s] = struct{}{}
		for s := range s.ε() {
			f(s)
		}
		return
	}
	f(s)
	return
}

func (s *State) edge(sym int) closure {
	return s.transitions().edge(sym, false)
}

// Transitions returns the symbol -> closure projection of state s.
func (s *State) Transitions() Transitions {
	return Transitions{s.transitions()}
}

func (s *State) transitions() transitions {
	if s.edges == nil {
		s.edges = transitions{}
	}
	return s.edges
}

func (s *State) ε() closure {
	return s.edge(Epsilon)
}

// Id returns the state's zero based index.
func (s *State) Id() int {
	return s.nfa.id(s)
}

// NewEdge connects state s and state next by a new edge, labeled by sym. By
// convention, passing sym == Epsilon is reserved to indicate adding of an ε
// edge.
//
//TODO implement priorities for sym < Epsilon
func (s *State) NewEdge(sym int, next *State) {
	s.transitions().newEdge(sym, true, next)
}

var (
	isAcceptingL = map[bool]string{true: "["}
	isAcceptingR = map[bool]string{true: "]"}
	isStart      = map[bool]string{true: "->"}
	isSep        = map[bool]string{true: " "}
)

// String implements fmt.Stringer for debugging, etc.
func (s *State) String() string {
	var b bytes.Buffer
	f := strutil.IndentFormatter(&b, "\t")
	f.Format("%s%s[%d]%s\n%i",
		isStart[s == s.nfa.start],
		isAcceptingL[s.IsAccepting],
		s.Id(),
		isAcceptingR[s.IsAccepting],
	)
	var syms sort.IntSlice
	for edge := range s.transitions() {
		syms = append(syms, edge)
	}
	sort.Sort(syms)
	for _, edge := range syms {
		nextSet := s.transitions()[edge]
		switch {
		case edge == Epsilon:
			f.Format("ε -> ")
		default:
			f.Format("%d -> ", edge)
		}
		isFirst := true
		for next := range nextSet {
			f.Format("%s[%d]", isSep[!isFirst], next.Id())
			isFirst = false
		}
		f.Format("\n")
	}
	return b.String()
}

// ----------------------------------------------------------------- Transitions

// Transitions maps symbols to their associated closures.
type Transitions struct {
	transitions
}

// NewTransitions returns a newly created Transitions.
func NewTransitions() Transitions {
	return Transitions{transitions{}}
}

type transitions map[int]closure

func (t transitions) edge(sym int, canCreate bool) (c closure) {
	c = t[sym]
	if c == nil {
		c = closure{}
		if canCreate {
			t[sym] = c
		}
	}
	return c
}

func (t transitions) newEdge(sym int, canCreate bool, next *State) (c closure) {
	c = t.edge(sym, canCreate)
	c[next] = struct{}{}
	return
}

// Delete removes the closure associated with sym.
func (t transitions) Delete(sym int) {
	delete(t, sym)
}

// Get returns the closure associated with sym.
func (t transitions) Get(sym int) (c Closure) {
	c.closure, _ = t[sym]
	return
}

// Set sets c as the closure associated with sym.
func (t transitions) Set(sym int, c Closure) {
	t[sym] = c.closure
}

// List returns a slice of all symbols appearing in the transitions.
func (t transitions) List() (r []int) {
	r = make([]int, len(t))
	i := 0
	for sym := range t {
		r[i] = sym
		i++
	}
	return
}
