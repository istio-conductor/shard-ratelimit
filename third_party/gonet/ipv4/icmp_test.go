// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ipv4_test

import (
	"net"
	"reflect"
	"runtime"
	"testing"

	"golang.org/x/net/nettest"
)

var icmpStringTests = []struct {
	in  ICMPType
	out string
}{
	{ICMPTypeDestinationUnreachable, "destination unreachable"},

	{256, "<nil>"},
}

func TestICMPString(t *testing.T) {
	for _, tt := range icmpStringTests {
		s := tt.in.String()
		if s != tt.out {
			t.Errorf("got %s; want %s", s, tt.out)
		}
	}
}

func TestICMPFilter(t *testing.T) {
	switch runtime.GOOS {
	case "linux":
	default:
		t.Skipf("not supported on %s", runtime.GOOS)
	}

	var f ICMPFilter
	for _, toggle := range []bool{false, true} {
		f.SetAll(toggle)
		for _, typ := range []ICMPType{
			ICMPTypeDestinationUnreachable,
			ICMPTypeEchoReply,
			ICMPTypeTimeExceeded,
			ICMPTypeParameterProblem,
		} {
			f.Accept(typ)
			if f.WillBlock(typ) {
				t.Errorf("ipv4.ICMPFilter.Set(%v, false) failed", typ)
			}
			f.Block(typ)
			if !f.WillBlock(typ) {
				t.Errorf("ipv4.ICMPFilter.Set(%v, true) failed", typ)
			}
		}
	}
}

func TestSetICMPFilter(t *testing.T) {
	switch runtime.GOOS {
	case "linux":
	default:
		t.Skipf("not supported on %s", runtime.GOOS)
	}
	if !nettest.SupportsRawSocket() {
		t.Skipf("not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	c, err := net.ListenPacket("ip4:icmp", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	p := NewPacketConn(c)

	var f ICMPFilter
	f.SetAll(true)
	f.Accept(ICMPTypeEcho)
	f.Accept(ICMPTypeEchoReply)
	if err := p.SetICMPFilter(&f); err != nil {
		t.Fatal(err)
	}
	kf, err := p.ICMPFilter()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(kf, &f) {
		t.Fatalf("got %#v; want %#v", kf, f)
	}
}
