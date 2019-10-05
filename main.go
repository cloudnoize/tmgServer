package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/cloudnoize/elport"
	locklessq "github.com/cloudnoize/locklessQ"
)

func selectDevice(str string) (int, error) {
	pa.ListDevices()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(str)
	text, _ := reader.ReadString('\n')
	devnum, err := strconv.Atoi(strings.TrimSuffix(text, "\n"))
	if err != nil {
		return 0, nil
	}
	return devnum, nil
}

func main() {

	dur := 10
	if v := os.Getenv("DURATION"); v != "" {
		dur, _ = strconv.Atoi(v)
	}

	op := "play"
	if v := os.Getenv("OP"); v != "" {
		op = v
	}
	log.Println(op)

	addr := ":8765"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	wavPath := "/home/eranl/WAV_FILES/server/"

	http.Handle("/", http.FileServer(http.Dir(wavPath)))

	log.SetFlags(log.LstdFlags | log.Llongfile)
	//16 bit
	sf := pa.SampleFormat(8)

	err := pa.Initialize()
	if err != nil {
		println("ERROR ", err.Error())
		return
	}
	desiredSR := 48000
	channels := 1
	ab := &AudioBuffer{q: locklessq.NewQint16(int32(desiredSR) * 300), sl: make([]int16, desiredSR*300), isRecord: true}
	sdone := make(chan struct{})
	done := make(chan struct{})
	recch := make(chan struct{})
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		log.Println("Got signal")
		sdone <- struct{}{}
		recch <- struct{}{}

		<-sigs
		log.Println("Got signal")
		done <- struct{}{}

	}()

	Recored(ab, sf, uint64(desiredSR), channels, recch, done)
	midi := NewMidiContext()
	if op == "udp" {
		go midi.SetNote()
		log.Println("UDP mode...")
		start := make(chan struct{})

		go midi.playMidi(recch, dur, start, sdone)
		go http.ListenAndServe(addr, GetHttpHandler(midi))
		ServeUdp(addr, ab, start, sdone)
		<-done
	} else if op == "play" {
		go midi.playMidi(recch, dur, nil, nil)
		<-done
		Play(ab, sf, uint64(desiredSR), channels, dur)

	}

	pa.Terminate()

	if v := os.Getenv("SAVE"); v != "" {
		saveWav(ab.sl, uint32(desiredSR), wavPath+GetFileName())
		log.Println("serving file...")
		go func() {
			err := http.ListenAndServe(":8766", nil)
			log.Println(err)
		}()
		<-done
	}
}
