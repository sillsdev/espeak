//go:generate build/make_js.bash
// +build linux

// Package espeak is a wrapper around espeak-ng that works both natively and in gopherjs with the same API.
// espeak-ng is an open source text to speech library that has over one hundred voices and languages and
// supports speech synthesis markup language (SSML).
package espeak // import "gopkg.in/BenLubar/espeak.v2"

import (
	"errors"
	"sync"
	"time"
)

// Error is the error type from espeak-ng.
type Error struct {
	Code    uint32 // Code associated with this error type in the espeak-ng C API.
	Message string // Message intended to be read by humans.
}

// Error implements the error interface.
func (err *Error) Error() string {
	return "espeak: " + err.Message
}

// SampleRate returns the number of samples per second in audio generated by this package.
func SampleRate() int {
	return getSampleRate()
}

var lock sync.Mutex

// Context contains the current state of text to speech data. Multiple Contexts may exist simultaneously,
// but each Context should only be accessed from one goroutine at a time. The zero value of a Context
// is empty with default values for rate, volume, pitch, and tone.
type Context struct {
	// Samples is a slice of audio samples in PCM format. Use the WriteTo method on the context to
	// encode Samples as a wav file.
	Samples []int16
	// Events are generated along with Samples and contain information about placement of words and
	// sentences, which may be useful, for example, when generating real time subtitles.
	Events []*SynthEvent

	rate   int // words per minute, 80 to 450; default 175
	volume int // percentage of normal volume, min 0; default 100
	pitch  int // base pitch, 0 to 100; default 50
	tone   int // pitch range, 0 to 100; 0 is monotone; default 50
	// TODO: punctuation?
	// TODO: capitals?
	// TODO: word gap?

	voice struct {
		name     string
		language string
		gender   Gender
		age      uint8
		variant  uint8
	}

	isInit bool
}

func (ctx *Context) init() {
	if ctx.isInit {
		return
	}

	ctx.isInit = true
	ctx.rate = 175
	ctx.volume = 100
	ctx.pitch = 50
	ctx.tone = 50
}

// Rate returns the current speed of speech in words per minute.
//
// The default rate is 175 words per minute.
func (ctx *Context) Rate() int {
	ctx.init()

	return ctx.rate
}

// Volume returns the current loudness of speech as a percent of the default volume.
func (ctx *Context) Volume() int {
	ctx.init()

	return ctx.volume
}

// Pitch returns the highness or lowness of the voice.
//
// The default pitch for the voice is represented by 50. Higher numbers are higher pitch.
func (ctx *Context) Pitch() int {
	ctx.init()

	return ctx.pitch
}

// Range returns the pitch range of speech.
//
// The default tone is 50. A tone of 0 is a monotonic voice.
func (ctx *Context) Range() int {
	ctx.init()

	return ctx.tone
}

// SetRate changes the speed of speech for future Synthesize calls to the given number of words per minute.
//
// The number of words per minute must be between 80 and 450, inclusive.
func (ctx *Context) SetRate(wpm int) {
	if wpm < 80 || wpm > 450 {
		panic("espeak: Context.SetRate: wpm must be between 80 and 450")
	}

	ctx.init()

	ctx.rate = wpm
}

// SetVolume changes the loudness of the voice for future Synthesize calls to a percentage of the default.
//
// The percentage must not be negative. Percentages over 100 may cause distortion or clipping.
func (ctx *Context) SetVolume(percentage int) {
	if percentage < 0 {
		panic("espeak: Context.SetVolume: percentage must not be negative")
	}

	ctx.init()

	ctx.volume = percentage
}

// SetPitch changes the highness or lowness of the voice for future Synthesize calls.
//
// Allowed values range from 0 (very low) to 100 (very high), with the original pitch for the voice being 50.
func (ctx *Context) SetPitch(pitch int) {
	if pitch < 0 || pitch > 100 {
		panic("espeak: Context.SetPitch: pitch must be between 0 and 100")
	}

	ctx.init()

	ctx.pitch = pitch
}

// SetRange changes the pitch range of the voice for future Synthesize calls.
//
// Allowed values range from 0 (monotone) to 100 (sing-songy), with the original range for the voice being 50.
func (ctx *Context) SetRange(tone int) {
	if tone < 0 || tone > 100 {
		panic("espeak: Context.SetRange: tone must be between 0 and 100")
	}

	ctx.init()

	ctx.tone = tone
}

// Voice is a voice supported by espeak.
type Voice struct {
	// Name for this voice (unique)
	Name string

	// Languages and priorities. Lower numbers mean this voice is more likely to be used for the language.
	Languages []Language

	// Identifier is the filename for this voice within espeak-ng-data/voices.
	Identifier string

	// Gender of voice.
	Gender Gender

	// Age in years, or 0 if not specified.
	Age uint8
}

// Language supported by a voice.
type Language struct {
	// Priority of the voice for this language. A low number indicates a more preferred voice, and
	// a higher number indicates a less preferred voice.
	Priority uint8

	// The name of the language, which may be in BCP47 format, but is not required to be.
	Name string
}

// ListVoices returns the complete list of voices supported by espeak. The returned slice is not shared,
// and callers may modify it without any side effects.
func ListVoices() []*Voice {
	lock.Lock()
	defer lock.Unlock()

	return listVoices()
}

// Gender of a voice.
type Gender uint8

// Voice genders
const (
	Unknown Gender = 0
	Male    Gender = 1
	Female  Gender = 2
	Neutral Gender = 3
)

// SetVoice sets a voice by name.
func (ctx *Context) SetVoice(name string) error {
	if name == "" {
		return errors.New("espeak: missing name in SetVoice")
	}

	return ctx.SetVoiceProperties(name, "", Unknown, 0, 0)
}

func validVoice(name, language string, gender Gender, age, variant uint8) error {
	lock.Lock()
	defer lock.Unlock()

	return setVoice(name, language, gender, age, variant)
}

// SetVoiceProperties sets the voice for future calls to Synthesize. Any or all of the arguments can be set
// to their zero values, in which case they will be ignored. Variant differentiates between multiple voices
// if more than one voice is matched by the other arguments.
func (ctx *Context) SetVoiceProperties(name, language string, gender Gender, age, variant uint8) error {
	if err := validVoice(name, language, gender, age, variant); err != nil {
		return err
	}

	ctx.init()

	ctx.voice.name = name
	ctx.voice.language = language
	ctx.voice.gender = gender
	ctx.voice.age = age
	ctx.voice.variant = variant

	return nil
}

// SynthEventType is the type of a SynthEvent.
type SynthEventType uint8

const (
	// EventWord is the start of a word.
	EventWord SynthEventType = 1

	// EventSentence is the start of a sentence.
	EventSentence SynthEventType = 2

	// EventMark is a <mark/> element in SSML.
	EventMark SynthEventType = 3

	// EventPlay is an <audio/> element in SSML.
	EventPlay SynthEventType = 4

	// EventEnd is the end of a sentence or clause.
	EventEnd SynthEventType = 5

	// EventMsgTerminated is the end of the synthesized message.
	EventMsgTerminated SynthEventType = 6

	// EventPhoneme is emitted for each phoneme if enabled.
	EventPhoneme SynthEventType = 7
)

// SynthEvent gives additional information about the generated speech.
type SynthEvent struct {
	// Type of the event.
	Type SynthEventType

	// TextPosition in characters from the start of the string. Unlike Go indexes, this starts at 1.
	TextPosition int

	// Length of the word, in characters. (for EventWord)
	Length int

	// AudioPosition is the time within the generated speech output data.
	AudioPosition time.Duration

	Number  int    // Number is used for EventWord and EventSentence
	Name    string // Name is used for EventMark and EventPlay
	Phoneme string // Phoneme is used for EventPhoneme
}

// TODO:
/*
func (ctx *Context) Synthesize(speak *ssml.Speak) error {
	ctx.init()

	text, err := xml.Marshal(speak)
	if err != nil {
		return err
	}

	return ctx.synthesize(string(text))
}
*/

// SynthesizeText converts the given text to speech.
//
// Some SSML tags are accepted. All other XML tags are ignored.
func (ctx *Context) SynthesizeText(text string) error {
	ctx.init()

	return ctx.synthesize(text)
}

func (ctx *Context) synthesize(text string) error {
	lock.Lock()
	defer lock.Unlock()

	if err := setRate(ctx.rate); err != nil {
		return err
	}

	if err := setVolume(ctx.volume); err != nil {
		return err
	}

	if err := setPitch(ctx.pitch); err != nil {
		return err
	}

	if err := setTone(ctx.tone); err != nil {
		return err
	}

	if err := setVoice(ctx.voice.name, ctx.voice.language, ctx.voice.gender, ctx.voice.age, ctx.voice.variant); err != nil {
		return err
	}

	return synthesize(text, ctx)
}
