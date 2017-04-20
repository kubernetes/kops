package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"sync"

	"github.com/fsouza/go-dockerclient"
)

func (proxy *Proxy) Intercept(i interceptor, w http.ResponseWriter, r *http.Request) {
	if err := i.InterceptRequest(r); err != nil {
		switch err.(type) {
		case *docker.NoSuchContainer:
			http.Error(w, err.Error(), http.StatusNotFound)
		case *ErrNoSuchImage:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			Log.Warning("Error intercepting request: ", err)
		}
		return
	}

	conn, err := proxy.Dial()
	if err != nil {
		http.Error(w, "Could not connect to target", http.StatusInternalServerError)
		Log.Warning(err)
		return
	}
	client := httputil.NewClientConn(conn, nil)
	defer client.Close()

	resp, err := client.Do(r)
	if err != nil && err != httputil.ErrPersistEOF {
		http.Error(w, fmt.Sprintf("Could not make request to target: %v", err), http.StatusInternalServerError)
		Log.Warning("Error forwarding request: ", err)
		return
	}
	err = i.InterceptResponse(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		Log.Warning("Error intercepting response: ", err)
		return
	}

	hdr := w.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			hdr.Add(k, v)
		}
	}
	Log.Debugf("Response from target: %s %v", resp.Status, w.Header())

	if resp.Header.Get("Content-Type") == "application/vnd.docker.raw-stream" {
		doRawStream(w, resp, client)
	} else if resp.TransferEncoding != nil && resp.TransferEncoding[0] == "chunked" {
		doChunkedResponse(w, resp, client)
	} else {
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			Log.Warning(err)
		}
	}
}

func doRawStream(w http.ResponseWriter, resp *http.Response, client *httputil.ClientConn) {
	down, downBuf, up, remaining, err := hijack(w, client)
	if err != nil {
		http.Error(w, "Unable to hijack connection for raw stream mode", http.StatusInternalServerError)
		return
	}
	defer down.Close()
	defer up.Close()
	defer func() {
		if err != nil {
			Log.Warning(err)
		}
	}()

	if _, err = downBuf.Write([]byte("HTTP/1.1 " + resp.Status + "\n")); err != nil {
		return
	}
	if err = resp.Header.Write(downBuf); err != nil {
		return
	}
	if _, err = downBuf.Write([]byte("\n")); err != nil {
		return
	}
	if err = downBuf.Flush(); err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go copyStream(down, io.MultiReader(remaining, up), &wg)
	go copyStream(up, downBuf, &wg)
	wg.Wait()
}

type closeWriter interface {
	CloseWrite() error
}

func copyStream(dst io.WriteCloser, src io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	if _, err := io.Copy(dst, src); err != nil {
		Log.Warning(err)
	}
	var err error
	if c, ok := dst.(closeWriter); ok {
		err = c.CloseWrite()
	} else {
		err = dst.Close()
	}
	if err != nil {
		Log.Warningf("Error closing connection: %s", err)
	}
}

type writeFlusher interface {
	io.Writer
	http.Flusher
}

func doChunkedResponse(w http.ResponseWriter, resp *http.Response, client *httputil.ClientConn) {
	wf, ok := w.(writeFlusher)
	if !ok {
		http.Error(w, "Error forwarding chunked response body: flush not available", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	wf.Flush()

	up, remaining := client.Hijack()
	defer up.Close()

	var err error
	chunks := NewChunkedReader(io.MultiReader(remaining, up))
	for chunks.Next() && err == nil {
		_, err = io.Copy(wf, chunks.Chunk())
		wf.Flush()
	}
	if err == nil {
		err = chunks.Err()
	}
	if err != nil {
		Log.Errorf("Error forwarding chunked response body: %s", err)
	}
}

func hijack(w http.ResponseWriter, client *httputil.ClientConn) (down net.Conn, downBuf *bufio.ReadWriter, up net.Conn, remaining io.Reader, err error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		err = errors.New("Unable to cast to Hijack")
		return
	}
	down, downBuf, err = hj.Hijack()
	if err != nil {
		return
	}
	up, remaining = client.Hijack()
	return
}
