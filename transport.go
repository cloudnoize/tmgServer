package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cloudnoize/conv"
)

func ServeUdp(addr string, ab *AudioBuffer, start chan struct{}, done chan struct{}) {
	conn, e := net.ListenPacket("udp", addr)
	if e != nil {
		println(e.Error())
		return
	}

	for {
		var b [4096]byte
		log.Println("Ready, send me signal to start")
		_, add, _ := conn.ReadFrom(b[:])
		log.Println("starting...")
		start <- struct{}{}
		log.Println("Start to stream audio,have ", ab.q.ReadAvailble(), " samples")

		var i int
		sleep := 10
		if v := os.Getenv("SLEEP"); v != "" {
			sleep, _ = strconv.Atoi(v)
		}

		chunks := 1024
		if v := os.Getenv("CHUNKS"); v != "" {
			chunks, _ = strconv.Atoi(v)
		}

		test := false
		if v := os.Getenv("TEST"); v != "" {
			test, _ = strconv.ParseBool(v)
		}

		log.Printf("test mode [%b]\n", test)

		log.Printf("will sleep %d between writes\n", sleep)
		for {
			select {
			case <-done:
				log.Println("Stopping UDP")
				conn.Close()
				return
			default:
				v, ok := ab.q.Pop()
				if !ok {
					continue
				}
				if test {
					v = 128
				}
				conv.Int16ToBytes(v, b[:], (i*2)%chunks)
				i++
				if (i*2)%chunks == 0 {
					conn.WriteTo(b[:chunks], add)
					time.Sleep(time.Duration(sleep) * time.Millisecond)
				}
			}

		}

	}
}

func GetHttpHandler(m *MidiContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snote, ok := r.URL.Query()["note"]

		if !ok || len(snote) < 1 {
			return
		}

		note, err := strconv.Atoi(snote[0])
		if err != nil {
			return
		}

		log.Println("Set note ", note)
		m.notes <- note
	}
}
