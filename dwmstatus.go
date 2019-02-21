package main

// #cgo LDFLAGS: -lX11
// #include <X11/Xlib.h>
import "C"

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var dpy = C.XOpenDisplay(nil)

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

	local, err := time.LoadLocation("Local")
	if err != nil {
		log.Fatal(err)
	}

	for {

		now := time.Now()

		t := now.In(local).Format("Mon Jan _2 15:04:05 2006")

		l, err := getLoadAverage("/proc/loadavg")
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("L: %s | %s", l, t)
		s := formatStatus("L: %s | %s", l, t)
		setStatus(s)
		time.Sleep(time.Second)
	}
}
