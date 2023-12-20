package blockchain

import (
	"crypto/sha256"
	"testing"
)

func TestNewBlock(t *testing.T) {
	// Arrange
	t1 := Transaction{
		Inputs: []Input{
			{Hash: [32]byte{}, ID: 0},
			{Hash: [32]byte{}, ID: 1},
		},
		Outputs: []Output{
			{PKH: [20]byte{}, Amount: 10},
			{PKH: [20]byte{}, Amount: 3},
		},
		Sign:      [32]byte{},
		PublicKey: [32]byte{},
	}
	t2 := Transaction{
		Inputs: []Input{
			{Hash: [32]byte{}, ID: 0},
			{Hash: [32]byte{}, ID: 1},
		},
		Outputs: []Output{
			{PKH: [20]byte{}, Amount: 12},
			{PKH: [20]byte{}, Amount: 3},
		},
		Sign:      [32]byte{},
		PublicKey: [32]byte{},
	}
	t3 := Transaction{
		Inputs: []Input{
			{Hash: [32]byte{}, ID: 0},
			{Hash: [32]byte{}, ID: 1},
		},
		Outputs: []Output{
			{PKH: [20]byte{}, Amount: 10},
			{PKH: [20]byte{}, Amount: 6},
		},
		Sign:      [32]byte{},
		PublicKey: [32]byte{},
	}
	t1h := t1.Hash()
	t2h := t2.Hash()
	t3h := t3.Hash()

	// Act
	s1 := sha256.Sum256(append(t1h[:], t2h[:]...))
	s2 := sha256.Sum256(append(t3h[:], t3h[:]...))
	s3 := sha256.Sum256(append(s1[:], s2[:]...))
	b := NewBlock([32]byte{}, &[]Transaction{t1, t2, t3})

	// Assert
	if b.Root != s3 {
		t.Error("Wrong merkle root.")
	}
}
