package ism

import (
	"encoding/xml"
)

type SmoothStreamingMedia struct {
	MajorVersion    uint16 `xml:"MajorVersion,attr,omitempty"`
	MinorVersion    uint16 `xml:"MinorVersion,attr,omitempty"`
	TimeScale       uint64 `xml:"TimeScale,attr,omitempty"`
	Duration        uint64 `xml:"Duration,attr,omitempty"`
	IsLive          bool   `xml:"IsLive,attr,omitempty"`
	LookaheadCount  uint32 `xml:"LookaheadCount,attr,omitempty"`
	DVRWindowLength uint64 `xml:"DVRWindowLength,attr,omitempty"`

	StreamIndexes []StreamIndex `xml:"StreamIndex,omitempty"`
	Protection    []Protection  `xml:"Protection,omitempty"`
}

func Unmarshal(ism []byte) (*SmoothStreamingMedia, error) {
	var ssm SmoothStreamingMedia
	err := xml.Unmarshal(ism, &ssm)
	if err != nil {
		return nil, err
	}
	if ssm.TimeScale == 0 {
		ssm.TimeScale = 10000000
	}
	return &ssm, nil
}

func (ism SmoothStreamingMedia) Marshal() ([]byte, error) {
	return xml.Marshal(ism)
}

type StreamIndex struct {
	Type             string `xml:"Type,attr,omitempty"`
	NumChunks        uint32 `xml:"Chunks,attr,omitempty"`
	URL              string `xml:"Url,attr,omitempty"`
	NumQualityLevels uint16 `xml:"QualityLevels,attr,omitempty"`
	MaxWidth         uint16 `xml:"MaxWidth,attr,omitempty"`
	MaxHeight        uint16 `xml:"MaxHeight,attr,omitempty"`
	DisplayWidth     uint16 `xml:"DisplayWidth,attr,omitempty"`
	DisplayHeight    uint16 `xml:"DisplayHeight,attr,omitempty"`

	Tracks []Track    `xml:"QualityLevel,omitempty"`
	Chunks []Fragment `xml:"c,omitempty"`
}

type Protection struct {
	ProtectionHeader []ProtectionHeader `xml:"ProtectionHeader,omitempty"`
}

type ProtectionHeader struct {
	SystemID string `xml:"SystemID,attr,omitempty"`

	Pro []byte `xml:",innerxml"`
}

type Track struct {
	Index              uint   `xml:"Index,attr,omitempty"`
	Bitrate            uint64 `xml:"Bitrate,attr,omitempty"`
	BufferTime         uint64 `xml:"BufferTime,attr,omitempty"`
	NominalBitrate     uint   `xml:"NominalBitrate,attr,omitempty"`
	HardwareProfile    string `xml:"HardwareProfile,attr,omitempty"`
	CodecPrivateData   string `xml:"CodecPrivateData,attr,omitempty"`
	MaxHeight          uint   `xml:"MaxHeight,attr,omitempty"`
	MaxWidth           uint   `xml:"MaxWidth,attr,omitempty"`
	SamplingRate       uint   `xml:"SamplingRate,attr,omitempty"`
	Channels           uint   `xml:"Channels,attr,omitempty"`
	BitsPerSample      uint64 `xml:"BitsPerSample,attr,omitempty"`
	PacketSize         uint64 `xml:"PacketSize,attr,omitempty"`
	AudioTag           string `xml:"AudioTag,attr,omitempty"`
	FourCC             string `xml:"FourCC,attr,omitempty"`
	NALUnitLengthField uint64 `xml:"NALUnitLengthField,attr,omitempty"`
}

type Fragment struct {
	D uint64 `xml:"d,attr,omitempty"`
	T uint64 `xml:"t,attr,omitempty"`
	N uint64 `xml:"n,attr,omitempty"`
	R uint64 `xml:"r,attr,omitempty"`
}
