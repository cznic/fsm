// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fsm

import (
	"fmt"
	"testing"
)

func ExampleNFA_Powerset() {
	// See http://en.wikipedia.org/wiki/Powerset_construction#Example
	n := NewNFA()
	s1, s2, s3, s4 := n.NewState(), n.NewState(), n.NewState(), n.NewState()
	s1.NewEdge(0, s2)
	s1.NewEdge(Epsilon, s3)
	s2.NewEdge(1, s2)
	s2.NewEdge(1, s4)
	s3.IsAccepting = true
	s3.NewEdge(0, s4)
	s3.NewEdge(Epsilon, s2)
	s4.IsAccepting = true
	s4.NewEdge(0, s3)
	fmt.Printf("NFA\n%v\nDFA\n%v\nDFA with a dead state (as seen in the linked article)\n%v", n, n.Powerset(false), n.Powerset(true))

	// Output:
	// NFA
	// ->[0]
	// 	ε -> [2]
	// 	0 -> [1]
	// [1]
	// 	1 -> [1] [3]
	// [[2]]
	// 	ε -> [1]
	// 	0 -> [3]
	// [[3]]
	// 	0 -> [2]
	//
	// DFA
	// ->[[0]]
	// 	0 -> [1]
	// 	1 -> [1]
	// [[1]]
	// 	0 -> [2]
	// 	1 -> [1]
	// [[2]]
	// 	0 -> [3]
	// 	1 -> [1]
	// [[3]]
	// 	0 -> [2]
	//
	// DFA with a dead state (as seen in the linked article)
	// ->[[0]]
	// 	0 -> [1]
	// 	1 -> [1]
	// [[1]]
	// 	0 -> [2]
	// 	1 -> [1]
	// [[2]]
	// 	0 -> [3]
	// 	1 -> [1]
	// [[3]]
	// 	0 -> [2]
	// 	1 -> [4]
	// [4]
	// 	0 -> [4]
	// 	1 -> [4]
}

func ExampleNFA_Reverse() {
	// See http://en.wikipedia.org/wiki/Powerset_construction#Example
	n := NewNFA()
	s1, s2, s3, s4 := n.NewState(), n.NewState(), n.NewState(), n.NewState()
	s1.NewEdge(0, s2)
	s1.NewEdge(Epsilon, s3)
	s2.NewEdge(1, s2)
	s2.NewEdge(1, s4)
	s3.IsAccepting = true
	s3.NewEdge(0, s4)
	s3.NewEdge(Epsilon, s2)
	s4.IsAccepting = true
	s4.NewEdge(0, s3)
	fmt.Printf("NFA\n%v\nNFA reversed\n%v", n, n.Reverse())

	// Output:
	// NFA
	// ->[0]
	// 	ε -> [2]
	// 	0 -> [1]
	// [1]
	// 	1 -> [1] [3]
	// [[2]]
	// 	ε -> [1]
	// 	0 -> [3]
	// [[3]]
	// 	0 -> [2]
	//
	// NFA reversed
	// [[0]]
	// [1]
	// 	ε -> [2]
	// 	0 -> [0]
	// 	1 -> [1]
	// [2]
	// 	ε -> [0]
	// 	0 -> [3]
	// [3]
	// 	0 -> [2]
	// 	1 -> [1]
	// ->[4]
	// 	ε -> [2] [3]
}

func ExampleNFA_MinimalDFA() {
	n := NewNFA()
	// NFA for regexp `012|12|02`
	s0, s1, s2, s3, s4, s5 := n.NewState(), n.NewState(), n.NewState(), n.NewState(), n.NewState(), n.NewState()
	s0.NewEdge(0, s1)
	s0.NewEdge(0, s5)
	s0.NewEdge(1, s4)
	s1.NewEdge(1, s2)
	s2.NewEdge(2, s3)
	s3.IsAccepting = true
	s4.NewEdge(2, s3)
	s5.NewEdge(2, s3)
	fmt.Printf(
		"NFA\n%v\nDFA\n%v\nMinimal DFA\n%v\nMinimal DFA with a dead state\n%v",
		n, n.Powerset(false), n.MinimalDFA(false), n.MinimalDFA(true),
	)

	// Output:
	// NFA
	// ->[0]
	// 	0 -> [1] [5]
	// 	1 -> [4]
	// [1]
	// 	1 -> [2]
	// [2]
	// 	2 -> [3]
	// [[3]]
	// [4]
	// 	2 -> [3]
	// [5]
	// 	2 -> [3]
	//
	// DFA
	// ->[0]
	// 	0 -> [1]
	// 	1 -> [4]
	// [1]
	// 	1 -> [2]
	// 	2 -> [3]
	// [2]
	// 	2 -> [3]
	// [[3]]
	// [4]
	// 	2 -> [3]
	//
	// Minimal DFA
	// ->[0]
	// 	0 -> [3]
	// 	1 -> [1]
	// [1]
	// 	2 -> [2]
	// [[2]]
	// [3]
	// 	1 -> [1]
	// 	2 -> [2]
	//
	// Minimal DFA with a dead state
	// ->[0]
	// 	0 -> [3]
	// 	1 -> [1]
	// 	2 -> [4]
	// [1]
	// 	0 -> [4]
	// 	1 -> [4]
	// 	2 -> [2]
	// [[2]]
	// 	0 -> [4]
	// 	1 -> [4]
	// 	2 -> [4]
	// [3]
	// 	0 -> [4]
	// 	1 -> [1]
	// 	2 -> [2]
	// [4]
	// 	0 -> [4]
	// 	1 -> [4]
	// 	2 -> [4]
}

func TestEpsilon(t *testing.T) {
	if g, e := Epsilon, -1; g != e {
		t.Fatal(g, e)
	}
}

//TODO Tests. Even though indirectly tested by the existing client code.