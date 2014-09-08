package main

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"code.google.com/p/gompd/mpd"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var dpy = C.XOpenDisplay(nil)

func getBatteryPercentage(path string) (perc int, err error) {
	energy_now, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_now", path))
	if err != nil {
		return -1, err
	}
	energy_full, err := ioutil.ReadFile(fmt.Sprintf("%s/energy_full", path))
	if err != nil {
		return -1, err
	}
	var enow, efull int
	fmt.Sscanf(string(energy_now), "%d", &enow)
	fmt.Sscanf(string(energy_full), "%d", &efull)
	return enow * 100 / efull, nil
}

func getLoadAverage(file string) (lavg string, err error) {
	loadavg, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
		return "Couldn't read loadavg", err
	}
	return strings.Join(strings.Fields(string(loadavg))[:3], " "), nil
}

func setStatus(s *C.char) {
	C.XStoreName(dpy, C.XDefaultRootWindow(dpy), s)
	C.XSync(dpy, 1)
}

func formatStatus(format string, args ...interface{}) *C.char {
	status := fmt.Sprintf(format, args...)
	return C.CString(status)
}

func main() {
	if dpy == nil {
		log.Fatal("Can't open display")
	}
	mpdClient, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Fatal(err)
	}
	for {
		t := time.Now().Format("15:04")
		b, err := getBatteryPercentage("/sys/class/power_supply/BAT1")
		if err != nil {
			log.Fatal(err)
		}
		status, err := mpdClient.Status()
		if err != nil {
			log.Fatal(err)
		}
		var np string
		if status["state"] == "play" || status["state"] == "pause" {
			attrs, err := mpdClient.CurrentSong()
			if err != nil {
				log.Fatal(err)
			}
			np = fmt.Sprintf("\x1b[1;31m%s\x1b[0m \x1b[1;37mby\x1b[0m \x1b[1;33m%s\x1b[0m \x1b[1;37mfrom\x1b[0m \x1b[1;34m%s\x1b[0m \x1b[1;30m::\x1b[0m ", attrs["Title"], attrs["Artist"], attrs["Album"])
		} else {
			np = ""
		}
		s := formatStatus("%s\x1b[1;32m%d%%\x1b[0m \x1b[1;30m::\x1b[0m %s", np, b, t)
		setStatus(s)
		time.Sleep(time.Second)
	}
}
