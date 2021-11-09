package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/krigga/ft/packets"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if len(os.Args) < 2 {
		log.Fatalln("supply filename")
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatalln(err)
	}

	for {
		p, err := packets.ReadPacket(conn)
		if err != nil {
			log.Fatalln(err)
		}
		switch p := p.(type) {
		case *packets.SeekRequest:
			n, err := file.Seek(p.Offset, int(p.Whence))
			if err != nil {
				log.Fatalln(err)
			}
			err = packets.WritePacket(conn, &packets.SeekResponse{N: n, Ok: true})
			if err != nil {
				log.Fatalln(err)
			}
		case *packets.ReadRequest:
			d := make([]byte, p.Size)
			n, err := file.Read(d)
			if err != nil && err != io.EOF {
				log.Fatalln(err)
			}
			if n != int(p.Size) && err != io.EOF {
				log.Fatalln(fmt.Errorf("requested length (%v) not equal to read length (%v)", p.Size, n))
			}
			err = packets.WritePacket(conn, &packets.ReadResponse{Data: d})
			if err != nil {
				log.Fatalln(err)
			}
		default:
			log.Fatalln(fmt.Errorf("unexpected packet %v", p))
		}
	}
}
