// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bpf_test

import (
	"testing"
)

func TestVMStoreScratchInvalidScratchRegisterTooSmall(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		StoreScratch{
			Src: RegA,
			N:   -1,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid scratch slot -1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchInvalidScratchRegisterTooLarge(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		StoreScratch{
			Src: RegA,
			N:   16,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid scratch slot 16" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchUnknownSourceRegister(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		StoreScratch{
			Src: 100,
			N:   0,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid source register 100" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchInvalidScratchRegisterTooSmall(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		LoadScratch{
			Dst: RegX,
			N:   -1,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid scratch slot -1" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchInvalidScratchRegisterTooLarge(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		LoadScratch{
			Dst: RegX,
			N:   16,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid scratch slot 16" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMLoadScratchUnknownDestinationRegister(t *testing.T) {
	_, _, err := testVM(t, []Instruction{
		LoadScratch{
			Dst: 100,
			N:   0,
		},
		RetA{},
	})
	if errStr(err) != "assembling instruction 1: invalid target register 100" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVMStoreScratchLoadScratchOneValue(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		// Load byte 255
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		// Copy to X and store in scratch[0]
		TAX{},
		StoreScratch{
			Src: RegX,
			N:   0,
		},
		// Load byte 1
		LoadAbsolute{
			Off:  9,
			Size: 1,
		},
		// Overwrite 1 with 255 from scratch[0]
		LoadScratch{
			Dst: RegA,
			N:   0,
		},
		// Return 255
		RetA{},
	})
	if err != nil {
		t.Fatalf("failed to load BPF program: %v", err)
	}
	defer done()

	out, err := vm.Run([]byte{
		0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff,
		255, 1, 2,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 3, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}

func TestVMStoreScratchLoadScratchMultipleValues(t *testing.T) {
	vm, done, err := testVM(t, []Instruction{
		// Load byte 10
		LoadAbsolute{
			Off:  8,
			Size: 1,
		},
		// Store in scratch[0]
		StoreScratch{
			Src: RegA,
			N:   0,
		},
		// Load byte 20
		LoadAbsolute{
			Off:  9,
			Size: 1,
		},
		// Store in scratch[1]
		StoreScratch{
			Src: RegA,
			N:   1,
		},
		// Load byte 30
		LoadAbsolute{
			Off:  10,
			Size: 1,
		},
		// Store in scratch[2]
		StoreScratch{
			Src: RegA,
			N:   2,
		},
		// Load byte 1
		LoadAbsolute{
			Off:  11,
			Size: 1,
		},
		// Store in scratch[3]
		StoreScratch{
			Src: RegA,
			N:   3,
		},
		// Load in byte 10 to X
		LoadScratch{
			Dst: RegX,
			N:   0,
		},
		// Copy X -> A
		TXA{},
		// Verify value is 10
		JumpIf{
			Cond:     JumpEqual,
			Val:      10,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		RetConstant{
			Val: 0,
		},
		// Load in byte 20 to A
		LoadScratch{
			Dst: RegA,
			N:   1,
		},
		// Verify value is 20
		JumpIf{
			Cond:     JumpEqual,
			Val:      20,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		RetConstant{
			Val: 0,
		},
		// Load in byte 30 to A
		LoadScratch{
			Dst: RegA,
			N:   2,
		},
		// Verify value is 30
		JumpIf{
			Cond:     JumpEqual,
			Val:      30,
			SkipTrue: 1,
		},
		// Fail test if incorrect
		RetConstant{
			Val: 0,
		},
		// Return first two bytes on success
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
		10, 20, 30, 1,
	})
	if err != nil {
		t.Fatalf("unexpected error while running program: %v", err)
	}
	if want, got := 2, out; want != got {
		t.Fatalf("unexpected number of output bytes:\n- want: %d\n-  got: %d",
			want, got)
	}
}
