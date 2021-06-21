// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"testing"
)

func TestVMJumpOne(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		Jump{
			Skip: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpOutOfProgram(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		Jump{
			Skip: 1,
		},
		RetA{},
	})
	if errStr(err) != "cannot jump 1 instructions; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfTrueOutOfProgram(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		JumpIf{
			Cond:     JumpEqual,
			SkipTrue: 2,
		},
		RetA{},
	})
	if errStr(err) != "cannot jump 2 instructions in true case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfFalseOutOfProgram(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		JumpIf{
			Cond:      JumpEqual,
			SkipFalse: 3,
		},
		RetA{},
	})
	if errStr(err) != "cannot jump 3 instructions in false case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfXTrueOutOfProgram(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		JumpIfX{
			Cond:     JumpEqual,
			SkipTrue: 2,
		},
		RetA{},
	})
	if errStr(err) != "cannot jump 2 instructions in true case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfXFalseOutOfProgram(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		JumpIfX{
			Cond:      JumpEqual,
			SkipFalse: 3,
		},
		RetA{},
	})
	if errStr(err) != "cannot jump 3 instructions in false case; jumping past program bounds" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMJumpIfEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		JumpIf{
			Cond:     JumpEqual,
			Val:      1,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfNotEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		JumpIf{
			Cond:      JumpNotEqual,
			Val:       1,
			SkipFalse: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfGreaterThan(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		JumpIf{
			Cond:     JumpGreaterThan,
			Val:      0x00010202,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfLessThan(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		JumpIf{
			Cond:     JumpLessThan,
			Val:      0xff010203,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfGreaterOrEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		JumpIf{
			Cond:     JumpGreaterOrEqual,
			Val:      0x00010203,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfLessOrEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		JumpIf{
			Cond:     JumpLessOrEqual,
			Val:      0xff010203,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfBitsSet(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		JumpIf{
			Cond:     JumpBitsSet,
			Val:      0x1122,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfBitsNotSet(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		JumpIf{
			Cond:     JumpBitsNotSet,
			Val:      0x1221,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		LoadConstant{
			Dst: RegX,
			Val: 1,
		},
		JumpIfX{
			Cond:     JumpEqual,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXNotEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		LoadConstant{
			Dst: RegX,
			Val: 1,
		},
		JumpIfX{
			Cond:      JumpNotEqual,
			SkipFalse: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 9,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 1, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXGreaterThan(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0x00010202,
		},
		JumpIfX{
			Cond:     JumpGreaterThan,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXLessThan(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0xff010203,
		},
		JumpIfX{
			Cond:     JumpLessThan,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXGreaterOrEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0x00010203,
		},
		JumpIfX{
			Cond:     JumpGreaterOrEqual,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXLessOrEqual(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 4,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0xff010203,
		},
		JumpIfX{
			Cond:     JumpLessOrEqual,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 12,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0, 1, 2, 3,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 4, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXBitsSet(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0x1122,
		},
		JumpIfX{
			Cond:     JumpBitsSet,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMJumpIfXBitsNotSet(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		LoadAbsolute{
			Off:  8,
			Size: 2,
		},
		LoadConstant{
			Dst: RegX,
			Val: 0x1221,
		},
		JumpIfX{
			Cond:     JumpBitsNotSet,
			SkipTrue: 1,
		},
		RetConstant{
			Val: 0,
		},
		RetConstant{
			Val: 10,
		},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		0x01, 0x02,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}
