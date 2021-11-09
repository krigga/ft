package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/krigga/ft/packets"
)

type handler struct {
	conn net.Conn
}

type fileDownload struct {
	conn net.Conn
}

func (fd *fileDownload) Read(d []byte) (int, error) {
	err := packets.WritePacket(fd.conn, &packets.ReadRequest{Size: int64(len(d))})
	if err != nil {
		return 0, err
	}
	p, err := packets.ReadPacket(fd.conn)
	if err != nil {
		return 0, err
	}
	switch p := p.(type) {
	case *packets.ReadResponse:
		if len(d) != len(p.Data) {
			return 0, fmt.Errorf("requested length (%v) not equal to received length (%v)", len(d), len(p.Data))
		}
		return copy(d, p.Data), nil
	default:
		return 0, fmt.Errorf("unexpected packet: %v", p)
	}
}

func (fd *fileDownload) Seek(offset int64, whence int) (int64, error) {
	err := packets.WritePacket(fd.conn, &packets.SeekRequest{Offset: offset, Whence: int64(whence)})
	if err != nil {
		return 0, err
	}
	p, err := packets.ReadPacket(fd.conn)
	if err != nil {
		return 0, err
	}
	switch p := p.(type) {
	case *packets.SeekResponse:
		if !p.Ok {
			err = errors.New("client seek errored")
		}
		return p.N, err
	default:
		return 0, fmt.Errorf("unexpected packet: %v", p)
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.conn == nil {
		r.Response.StatusCode = 404
		fmt.Fprint(w, "not found")
	} else {
		fname := "test.txt"
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", fname))
		http.ServeContent(w, r, fname, time.Time{}, &fileDownload{
			conn: h.conn,
		})
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	h := &handler{}
	http.Handle("/", h)
	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", nil)
		if err != nil {
			log.Fatalln(err)
		}
	}()
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	h.conn = conn
	select {}
}
