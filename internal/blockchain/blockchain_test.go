package blockchain

import (
	"crypto/sha256"
	"testing"
)

func TestNewBlock(t *testing.T) {
	// Arrange
	t1 := Transaction{
		input: [32]byte{},
		output: Output{
			pkh:   [20]byte{2, 3, 4},
			token: Token{[32]byte{}, make([]byte, 1024)},
		},
		sign:      [32]byte{},
		publicKey: [32]byte{},
	}
	t2 := Transaction{
		input: [32]byte{},
		output: Output{
			pkh:   [20]byte{2, 3, 4},
			token: Token{[32]byte{}, make([]byte, 1024)},
		},
		sign:      [32]byte{},
		publicKey: [32]byte{},
	}
	t3 := Transaction{
		input: [32]byte{},
		output: Output{
			pkh:   [20]byte{2, 3, 4},
			token: Token{[32]byte{}, make([]byte, 1024)},
		},
		sign:      [32]byte{},
		publicKey: [32]byte{},
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
	if b.root != s3 {
		t.Error("Wrong merkle root.")
	}
}
