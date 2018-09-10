package mesh

import (
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testConn struct {
	io.Writer
	io.Reader
}

func (testConn) SetDeadline(t time.Time) error {
	return nil
}

func (testConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (testConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func connPair() (protocolIntroConn, protocolIntroConn) {
	a := testConn{}
	b := testConn{}
	a.Reader, b.Writer = io.Pipe()
	b.Reader, a.Writer = io.Pipe()
	return &a, &b
}

func doIntro(t *testing.T, params protocolIntroParams) <-chan protocolIntroResults {
	ch := make(chan protocolIntroResults, 1)
	go func() {
		res, err := params.doIntro()
		require.Nil(t, err)
		ch <- res
	}()
	return ch
}

func doProtocolIntro(t *testing.T, aver, bver byte, password []byte) byte {
	aconn, bconn := connPair()
	aresch := doIntro(t, protocolIntroParams{
		MinVersion: ProtocolMinVersion,
		MaxVersion: aver,
		Features:   map[string]string{"Name": "A"},
		Conn:       aconn,
		Outbound:   true,
		Password:   password,
	})
	bresch := doIntro(t, protocolIntroParams{
		MinVersion: ProtocolMinVersion,
		MaxVersion: bver,
		Features:   map[string]string{"Name": "B"},
		Conn:       bconn,
		Outbound:   false,
		Password:   password,
	})
	ares := <-aresch
	bres := <-bresch

	// Check that features were conveyed
	require.Equal(t, "B", ares.Features["Name"])
	require.Equal(t, "A", bres.Features["Name"])

	// Check that Senders and Receivers work
	go func() {
		require.Nil(t, ares.Sender.Send([]byte("Hello from A")))
		require.Nil(t, bres.Sender.Send([]byte("Hello from B")))
	}()

	data, err := bres.Receiver.Receive()
	require.Nil(t, err)
	require.Equal(t, "Hello from A", string(data))

	data, err = ares.Receiver.Receive()
	require.Nil(t, err)
	require.Equal(t, "Hello from B", string(data))

	require.Equal(t, ares.Version, bres.Version)
	return ares.Version
}

func TestProtocolIntro(t *testing.T) {
	require.Equal(t, 2, int(doProtocolIntro(t, 2, 2, nil)))
	require.Equal(t, 2, int(doProtocolIntro(t, 2, 2, []byte("sekr1t"))))
	require.Equal(t, 1, int(doProtocolIntro(t, 1, 2, nil)))
	require.Equal(t, 1, int(doProtocolIntro(t, 1, 2, []byte("pa55"))))
	require.Equal(t, 1, int(doProtocolIntro(t, 2, 1, nil)))
	require.Equal(t, 1, int(doProtocolIntro(t, 2, 1, []byte("w0rd"))))
}
