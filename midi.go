package main

import (
	"github.com/rakyll/portmidi"
	"sync"
	"time"
)

func initMidi() {
	defineFunction("midi-open", midiOpen, NullType)
	defineFunction("midi-time", midiTime, NumberType)
	defineFunction("midi-sleep", midiSleep, NumberType, NumberType)
	defineFunction("midi-write", midiWrite, NullType, NumberType, NumberType, NumberType)
	defineFunction("midi-close", midiClose, NullType)
}

var midiOut *portmidi.Stream
var midiMutex = &sync.Mutex{}

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

// (midi-time) -> seconds
func midiTime(argc []*LOB) (*LOB, error) {
	n := float64(portmidi.Time())
	return newFloat64(n / 1000.0), nil
}

func midiSleep(argv []*LOB) (*LOB, error) {
	time.Sleep(time.Duration(argv[0].fval*1000.0) * time.Millisecond)
	return midiTime(nil)
}

func midiAllNotesOff() {
	midiOut.WriteShort(0xB0, 0x7B, 0x00)
}

func midiClose(argv []*LOB) (*LOB, error) {
	midiMutex.Lock()
	if midiOut != nil {
		midiAllNotesOff()
		midiOut.Close()
		midiOut = nil
	}
	midiMutex.Unlock()
	return Null, nil
}

// (midi-write 144 60 80) -> middle C note on
// (midi-write 128 60 0) -> middle C note off
func midiWrite(argv []*LOB) (*LOB, error) {
	status := int64(argv[0].fval)
	data1 := int64(argv[1].fval)
	data2 := int64(argv[2].fval)
	midiMutex.Lock()
	var err error
	if midiOut != nil {
		err = midiOut.WriteShort(status, data1, data2)
	}
	midiMutex.Unlock()
	return Null, err
}
