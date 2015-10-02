package main

import (
	"github.com/rakyll/portmidi"
	"time"
)

func initMidi() {
	defineFunction("midi-open", midiOpen, "()")
	defineFunction("midi-time", midiTime, "() <number>")
	defineFunction("midi-sleep", midiSleep, "(<number>) <number>")
	defineFunction("midi-write", midiWrite, "(<number> <number> <number>)")
	defineFunction("midi-close", midiClose, "()")
}

var midiOut *portmidi.Stream

func midiOpen(argv []*LOB) (*LOB, error) {
	latency := int64(0)
	bufsize := int64(1024)
	if midiOut == nil {
		err := portmidi.Initialize()
		if err != nil {
			return nil, err
		}
		deviceID := portmidi.GetDefaultOutputDeviceId()
		out, err := portmidi.NewOutputStream(deviceID, bufsize, latency)
		if err != nil {
			return nil, err
		}
		midiOut = out
	}
	return Null, nil
}

func midiSleep(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 1 {
		ms, err := int64Value(argv[0])
		if err == nil {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return midiTime(nil)
		}
	}
	return nil, Error(ArgumentErrorKey, "midi-sleep takes a single number argument")
}

func midiAllNotesOff() {
	if midiOut != nil {
		midiOut.WriteShort(0xB0, 0x7B, 0x00)
	}
}

func midiClose(argv []*LOB) (*LOB, error) {
	if midiOut != nil {
		midiAllNotesOff()
		midiOut.Close()
		midiOut = nil
	}
	return Null, nil
}

// (midi-time) -> milliseconds
func midiTime(argc []*LOB) (*LOB, error) {
	n := int64(portmidi.Time())
	return newInt64(n), nil
}

// (midi-write 144 60 80) -> middle C note on
// (midi-write 128 60 0) -> middle C note off
func midiWrite(argv []*LOB) (*LOB, error) {
	argc := len(argv)
	if argc == 3 {
		status, err1 := int64Value(argv[0])
		data1, err2 := int64Value(argv[1])
		data2, err3 := int64Value(argv[2])
		if err1 == nil && err2 == nil && err3 == nil {
			return Null, midiOut.WriteShort(status, data1, data2)
		}
	}
	return nil, Error(ArgumentErrorKey, "midi-write takes 3 numbers")
}
