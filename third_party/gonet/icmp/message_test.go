// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icmp_test

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"golang.org/x/net/internal/iana"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

func TestMarshalAndParseMessage(t *testing.T) {
	fn := func(t *testing.T, proto int, tms []Message) {
		var pshs [][]byte
		switch proto {
		case iana.ProtocolICMP:
			pshs = [][]byte{nil}
		case iana.ProtocolIPv6ICMP:
			pshs = [][]byte{
				IPv6PseudoHeader(net.ParseIP("fe80::1"), net.ParseIP("ff02::1")),
				nil,
			}
		}
		for i, tm := range tms {
			for _, psh := range pshs {
				b, err := tm.Marshal(psh)
				if err != nil {
					t.Fatalf("#%d: %v", i, err)
				}
				m, err := ParseMessage(proto, b)
				if err != nil {
					t.Fatalf("#%d: %v", i, err)
				}
				if m.Type != tm.Type || m.Code != tm.Code {
					t.Errorf("#%d: got %#v; want %#v", i, m, &tm)
					continue
				}
				if !reflect.DeepEqual(m.Body, tm.Body) {
					t.Errorf("#%d: got %#v; want %#v", i, m.Body, tm.Body)
					continue
				}
			}
		}
	}

	t.Run("IPv4", func(t *testing.T) {
		fn(t, iana.ProtocolICMP,
			[]Message{
				{
					Type: ipv4.ICMPTypeDestinationUnreachable, Code: 15,
					Body: &DstUnreach{
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv4.ICMPTypeTimeExceeded, Code: 1,
					Body: &TimeExceeded{
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv4.ICMPTypeParameterProblem, Code: 2,
					Body: &ParamProb{
						Pointer: 8,
						Data:    []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv4.ICMPTypeEcho, Code: 0,
					Body: &Echo{
						ID: 1, Seq: 2,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
				{
					Type: ipv4.ICMPTypeExtendedEchoRequest, Code: 0,
					Body: &ExtendedEchoRequest{
						ID: 1, Seq: 2,
						Extensions: []Extension{
							&InterfaceIdent{
								Class: 3,
								Type:  1,
								Name:  "en101",
							},
						},
					},
				},
				{
					Type: ipv4.ICMPTypeExtendedEchoReply, Code: 0,
					Body: &ExtendedEchoReply{
						State: 4 /* Delay */, Active: true, IPv4: true,
					},
				},
			})
	})
	t.Run("IPv6", func(t *testing.T) {
		fn(t, iana.ProtocolIPv6ICMP,
			[]Message{
				{
					Type: ipv6.ICMPTypeDestinationUnreachable, Code: 6,
					Body: &DstUnreach{
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv6.ICMPTypePacketTooBig, Code: 0,
					Body: &PacketTooBig{
						MTU:  1<<16 - 1,
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv6.ICMPTypeTimeExceeded, Code: 1,
					Body: &TimeExceeded{
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv6.ICMPTypeParameterProblem, Code: 2,
					Body: &ParamProb{
						Pointer: 8,
						Data:    []byte("ERROR-INVOKING-PACKET"),
					},
				},
				{
					Type: ipv6.ICMPTypeEchoRequest, Code: 0,
					Body: &Echo{
						ID: 1, Seq: 2,
						Data: []byte("HELLO-R-U-THERE"),
					},
				},
				{
					Type: ipv6.ICMPTypeExtendedEchoRequest, Code: 0,
					Body: &ExtendedEchoRequest{
						ID: 1, Seq: 2,
						Extensions: []Extension{
							&InterfaceIdent{
								Class: 3,
								Type:  2,
								Index: 911,
							},
						},
					},
				},
				{
					Type: ipv6.ICMPTypeExtendedEchoReply, Code: 0,
					Body: &ExtendedEchoReply{
						State: 5 /* Probe */, Active: true, IPv6: true,
					},
				},
			})
	})
}

func TestMarshalAndParseRawMessage(t *testing.T) {
	t.Run("RawBody", func(t *testing.T) {
		for i, tt := range []struct {
			m               Message
			wire            []byte
			parseShouldFail bool
		}{
			{ // Nil body
				m: Message{
					Type: ipv4.ICMPTypeDestinationUnreachable, Code: 127,
				},
				wire: []byte{
					0x03, 0x7f, 0xfc, 0x80,
				},
				parseShouldFail: true,
			},
			{ // Empty body
				m: Message{
					Type: ipv6.ICMPTypeDestinationUnreachable, Code: 128,
					Body: &RawBody{},
				},
				wire: []byte{
					0x01, 0x80, 0x00, 0x00,
				},
				parseShouldFail: true,
			},
			{ // Crafted body
				m: Message{
					Type: ipv6.ICMPTypeDuplicateAddressConfirmation, Code: 129,
					Body: &RawBody{
						Data: []byte{0xca, 0xfe},
					},
				},
				wire: []byte{
					0x9e, 0x81, 0x00, 0x00,
					0xca, 0xfe,
				},
				parseShouldFail: false,
			},
		} {
			b, err := tt.m.Marshal(nil)
			if err != nil {
				t.Errorf("#%d: %v", i, err)
				continue
			}
			if !bytes.Equal(b, tt.wire) {
				t.Errorf("#%d: got %#v; want %#v", i, b, tt.wire)
				continue
			}
			m, err := ParseMessage(tt.m.Type.Protocol(), b)
			if err != nil != tt.parseShouldFail {
				t.Errorf("#%d: got %v, %v", i, m, err)
				continue
			}
			if tt.parseShouldFail {
				continue
			}
			if m.Type != tt.m.Type || m.Code != tt.m.Code {
				t.Errorf("#%d: got %v; want %v", i, m, tt.m)
				continue
			}
			if !bytes.Equal(m.Body.(*RawBody).Data, tt.m.Body.(*RawBody).Data) {
				t.Errorf("#%d: got %#v; want %#v", i, m.Body, tt.m.Body)
				continue
			}
		}
	})
	t.Run("RawExtension", func(t *testing.T) {
		for i, tt := range []struct {
			m    Message
			wire []byte
		}{
			{ // Unaligned data and nil extension
				m: Message{
					Type: ipv6.ICMPTypeDestinationUnreachable, Code: 130,
					Body: &DstUnreach{
						Data: []byte("ERROR-INVOKING-PACKET"),
					},
				},
				wire: []byte{
					0x01, 0x82, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00,
					'E', 'R', 'R', 'O',
					'R', '-', 'I', 'N',
					'V', 'O', 'K', 'I',
					'N', 'G', '-', 'P',
					'A', 'C', 'K', 'E',
					'T',
				},
			},
			{ // Unaligned data and empty extension
				m: Message{
					Type: ipv6.ICMPTypeDestinationUnreachable, Code: 131,
					Body: &DstUnreach{
						Data: []byte("ERROR-INVOKING-PACKET"),
						Extensions: []Extension{
							&RawExtension{},
						},
					},
				},
				wire: []byte{
					0x01, 0x83, 0x00, 0x00,
					0x02, 0x00, 0x00, 0x00,
					'E', 'R', 'R', 'O',
					'R', '-', 'I', 'N',
					'V', 'O', 'K', 'I',
					'N', 'G', '-', 'P',
					'A', 'C', 'K', 'E',
					'T',
					0x20, 0x00, 0xdf, 0xff,
				},
			},
			{ // Nil extension
				m: Message{
					Type: ipv6.ICMPTypeExtendedEchoRequest, Code: 132,
					Body: &ExtendedEchoRequest{
						ID: 1, Seq: 2, Local: true,
					},
				},
				wire: []byte{
					0xa0, 0x84, 0x00, 0x00,
					0x00, 0x01, 0x02, 0x01,
				},
			},
			{ // Empty extension
				m: Message{
					Type: ipv6.ICMPTypeExtendedEchoRequest, Code: 133,
					Body: &ExtendedEchoRequest{
						ID: 1, Seq: 2, Local: true,
						Extensions: []Extension{
							&RawExtension{},
						},
					},
				},
				wire: []byte{
					0xa0, 0x85, 0x00, 0x00,
					0x00, 0x01, 0x02, 0x01,
					0x20, 0x00, 0xdf, 0xff,
				},
			},
			{ // Crafted extension
				m: Message{
					Type: ipv6.ICMPTypeExtendedEchoRequest, Code: 134,
					Body: &ExtendedEchoRequest{
						ID: 1, Seq: 2, Local: true,
						Extensions: []Extension{
							&RawExtension{
								Data: []byte("CRAFTED"),
							},
						},
					},
				},
				wire: []byte{
					0xa0, 0x86, 0x00, 0x00,
					0x00, 0x01, 0x02, 0x01,
					0x20, 0x00, 0xc3, 0x21,
					'C', 'R', 'A', 'F',
					'T', 'E', 'D',
				},
			},
		} {
			b, err := tt.m.Marshal(nil)
			if err != nil {
				t.Errorf("#%d: %v", i, err)
				continue
			}
			if !bytes.Equal(b, tt.wire) {
				t.Errorf("#%d: got %#v; want %#v", i, b, tt.wire)
				continue
			}
			m, err := ParseMessage(tt.m.Type.Protocol(), b)
			if err != nil {
				t.Errorf("#%d: %v", i, err)
				continue
			}
			if m.Type != tt.m.Type || m.Code != tt.m.Code {
				t.Errorf("#%d: got %v; want %v", i, m, tt.m)
				continue
			}
		}
	})
}
