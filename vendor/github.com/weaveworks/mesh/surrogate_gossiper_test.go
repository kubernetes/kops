package mesh

import "testing"
import "time"
import "github.com/stretchr/testify/require"

func TestSurrogateGossiperUnicast(t *testing.T) {
	t.Skip("TODO")
}

func TestSurrogateGossiperBroadcast(t *testing.T) {
	t.Skip("TODO")
}

func TestSurrogateGossiperGossip(t *testing.T) {
	t.Skip("TODO")
}

func checkOnGossip(t *testing.T, s Gossiper, input, expected []byte) {
	r, err := s.OnGossip(input)
	require.NoError(t, err)
	if r == nil {
		if expected == nil {
			return
		}
		require.Fail(t, "Gossip result should NOT be nil, but was")
	}
	require.Equal(t, [][]byte{expected}, r.Encode())
}

func TestSurrogateGossiperOnGossip(t *testing.T) {
	myTime := time.Now()
	now = func() time.Time { return myTime }
	s := &surrogateGossiper{}
	msg := [][]byte{[]byte("test 1"), []byte("test 2"), []byte("test 3"), []byte("test 4")}
	checkOnGossip(t, s, msg[0], msg[0])
	checkOnGossip(t, s, msg[1], msg[1])
	checkOnGossip(t, s, msg[0], nil)
	checkOnGossip(t, s, msg[1], nil)
	myTime = myTime.Add(gossipInterval / 2) // Should not trigger cleardown
	checkOnGossip(t, s, msg[2], msg[2])     // Only clears out old ones on new entry
	checkOnGossip(t, s, msg[0], nil)
	checkOnGossip(t, s, msg[1], nil)
	myTime = myTime.Add(gossipInterval)
	checkOnGossip(t, s, msg[0], nil)
	checkOnGossip(t, s, msg[3], msg[3]) // Only clears out old ones on new entry
	checkOnGossip(t, s, msg[0], msg[0])
	checkOnGossip(t, s, msg[0], nil)
}

func TestSurrogateGossipDataEncode(t *testing.T) {
	t.Skip("TODO")
}

func TestSurrogateGossipDataMerge(t *testing.T) {
	t.Skip("TODO")
}
