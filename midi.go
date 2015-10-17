package main

import (
	"github.com/rakyll/portmidi"
	"sync"
	"time"
)

func initMidi() {
   defineFunctionKeyArgs("midi-open", midiOpen, NullType,
		[]*LOB{StringType, StringType, NumberType},
		[]*LOB{EmptyString,EmptyString,newInt(1024)},
		[]*LOB{intern("input:"), intern("output:"), intern("bufsize")})
	defineFunction("midi-write", midiWrite, NullType, NumberType, NumberType, NumberType)
	defineFunction("midi-close", midiClose, NullType)
	defineFunction("midi-listen", midiListen, ChannelType)
}

var midiOpened = false
var midiInDevice string
var midiOutDevice string
var midiBufsize int64
var midiBaseTime float64

var midiOut *portmidi.Stream
var midiIn *portmidi.Stream
var midiChannel chan portmidi.Event
var midiMutex = &sync.Mutex{}

func findMidiInputDevice(name string) portmidi.DeviceId {
	devcount := portmidi.CountDevices()
	for i := 0; i < devcount; i++ {
		id := portmidi.DeviceId(i)
		info := portmidi.GetDeviceInfo(id)
		if info.IsInputAvailable {
			if info.Name == name {
				return id
			}
		}
	}
	return portmidi.DeviceId(-1)
}

func findMidiOutputDevice(name string) (portmidi.DeviceId, string) {
	devcount := portmidi.CountDevices()
	for i := 0; i < devcount; i++ {
		id := portmidi.DeviceId(i)
		info := portmidi.GetDeviceInfo(id)
		if info.IsOutputAvailable {
			if info.Name == name {
				return id, info.Name
			}
		}
	}
	id := portmidi.GetDefaultOutputDeviceId()
	info := portmidi.GetDeviceInfo(id)
	return id, info.Name
}

func midiOpen(argv []*LOB) (*LOB, error) {
//	defaultInput := "USB Oxygen 8 v2"
//	defaultOutput := "IAC Driver Bus 1"
	latency := int64(10)
	if !midiOpened {
		err := portmidi.Initialize()
		if err != nil {
			return nil, err
		}
		midiOpened = true
		midiInDevice = argv[0].text
		midiOutDevice = argv[1].text
		midiBufsize = int64(argv[2].fval)

		outdev, outname := findMidiOutputDevice(midiOutDevice)
		out, err := portmidi.NewOutputStream(outdev, midiBufsize, latency)
		if err != nil {
			return nil, err
		}
		midiOut = out
		midiOutDevice = outname
		if midiInDevice != "" {
			indev := findMidiInputDevice(midiInDevice)
			if indev >= 0 {
				in, err := portmidi.NewInputStream(indev, midiBufsize)
				if err != nil {
					return nil, err
				}
				midiIn = in
			}
		}
		midiBaseTime = now()
		
	}
	result := makeStruct(4)
	if midiInDevice != "" {
		put(result, intern("input:"), newString(midiInDevice))
	}
	if midiOutDevice != "" {
		put(result, intern("output:"), newString(midiOutDevice))
	}
	put(result, intern("bufsize:"), newInt64(midiBufsize))
	return result, nil
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

func midiListen(argv []*LOB) (*LOB, error) {
	ch := Null
	midiMutex.Lock()
	if midiIn != nil {
		ch = newChannel(int(midiBufsize), "midi")
		go func(s *portmidi.Stream, ch *LOB) {
			for {
				time.Sleep(10 * time.Millisecond)
				events, err := s.Read(1024)
				if err != nil {
					continue
				}
				channel := ch.channel
				if channel != nil {
					for _, ev := range events {
						ts := (float64(ev.Timestamp) / 1000) + midiBaseTime
						st := ev.Status
						d1 := ev.Data1
						d2 := ev.Data2
						channel <- list(newFloat64(ts), newInt64(st), newInt64(d1), newInt64(d2))
					}
				}
			}
		}(midiIn, ch)
	}
	midiMutex.Unlock()
	return ch, nil
}

