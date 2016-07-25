package fi

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"
)

// A test to make sure that fmt.Sprintf on a big.Int is the same as Text
func TestBigInt_Format(t *testing.T) {
	rnd := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	var limit big.Int
	limit.Lsh(big.NewInt(1), 100)
	for i := 1; i < 100; i++ {
		var r big.Int
		r.Rand(rnd, &limit)
		s1 := r.String()
		s2 := r.Text(10)

		fmt.Printf("%s\n", s1)
		if s1 != s2 {
			t.Fatalf("%s not the same as %s", s1, s2)
		}
	}
}
