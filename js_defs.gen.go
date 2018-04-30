// +build js
// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs js_defs.go

package espeak

const outputModeSynchronous = 0x1
const sOK = 0x0
const posCharacter = 0x1

const espeakCHARS_UTF8 = 0x1
const espeakSSML = 0x10

const espeakRATE = 0x1
const espeakVOLUME = 0x2
const espeakPITCH = 0x3
const espeakRANGE = 0x4

const espeakEVENT_LIST_TERMINATED = 0x0
const espeakEVENT_WORD = 0x1
const espeakEVENT_SENTENCE = 0x2
const espeakEVENT_MARK = 0x3
const espeakEVENT_PLAY = 0x4
const espeakEVENT_END = 0x5
const espeakEVENT_MSG_TERMINATED = 0x6
const espeakEVENT_PHONEME = 0x7
const espeakEVENT_SAMPLERATE = 0x8

const eventSize = 0x24
const eventTypeOffset = 0x0
const eventTextPositionOffset = 0x8
const eventLengthOffset = 0xc
const eventAudioPositionOffset = 0x10
const eventNumberOffset = 0x1c
const eventNameOffset = 0x1c
const eventStringOffset = 0x1c

const voiceSize = 0x18
const voiceNameOffset = 0x0
const voiceLanguagesOffset = 0x4
const voiceIdentifierOffset = 0x8
const voiceGenderOffset = 0xc
const voiceAgeOffset = 0xd
const voiceVariantOffset = 0xe