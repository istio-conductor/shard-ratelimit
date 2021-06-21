// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dnsmessage_test

import (
	"fmt"
	"net"
	"strings"
)

func mustNewName(name string) Name {
	n, err := NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}

func ExampleParser() {
	msg := Message{
		Header: Header{Response: true, Authoritative: true},
		Questions: []Question{
			{
				Name:  mustNewName("foo.bar.example.com."),
				Type:  TypeA,
				Class: ClassINET,
			},
			{
				Name:  mustNewName("bar.example.com."),
				Type:  TypeA,
				Class: ClassINET,
			},
		},
		Answers: []Resource{
			{
				Header: ResourceHeader{
					Name:  mustNewName("foo.bar.example.com."),
					Type:  TypeA,
					Class: ClassINET,
				},
				Body: &AResource{A: [4]byte{127, 0, 0, 1}},
			},
			{
				Header: ResourceHeader{
					Name:  mustNewName("bar.example.com."),
					Type:  TypeA,
					Class: ClassINET,
				},
				Body: &AResource{A: [4]byte{127, 0, 0, 2}},
			},
		},
	}

	buf, err := msg.Pack()
	if err != nil {
		panic(err)
	}

	wantName := "bar.example.com."

	var p Parser
	if _, err := p.Start(buf); err != nil {
		panic(err)
	}

	for {
		q, err := p.Question()
		if err == ErrSectionDone {
			break
		}
		if err != nil {
			panic(err)
		}

		if q.Name.String() != wantName {
			continue
		}

		fmt.Println("Found question for name", wantName)
		if err := p.SkipAllQuestions(); err != nil {
			panic(err)
		}
		break
	}

	var gotIPs []net.IP
	for {
		h, err := p.AnswerHeader()
		if err == ErrSectionDone {
			break
		}
		if err != nil {
			panic(err)
		}

		if (h.Type != TypeA && h.Type != TypeAAAA) || h.Class != ClassINET {
			continue
		}

		if !strings.EqualFold(h.Name.String(), wantName) {
			if err := p.SkipAnswer(); err != nil {
				panic(err)
			}
			continue
		}

		switch h.Type {
		case TypeA:
			r, err := p.AResource()
			if err != nil {
				panic(err)
			}
			gotIPs = append(gotIPs, r.A[:])
		case TypeAAAA:
			r, err := p.AAAAResource()
			if err != nil {
				panic(err)
			}
			gotIPs = append(gotIPs, r.AAAA[:])
		}
	}

	fmt.Printf("Found A/AAAA records for name %s: %v\n", wantName, gotIPs)

	// Output:
	// Found question for name bar.example.com.
	// Found A/AAAA records for name bar.example.com.: [127.0.0.2]
}
