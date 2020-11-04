package mpd

import (
	"encoding/xml"
)

// MPD represents root XML element for parse.
type MPD struct {
	XMLName                    xml.Name `xml:"MPD"`
	XMLNS                      string   `xml:"xmlns,attr"`
	Type                       string   `xml:"type,attr"`
	MinimumUpdatePeriod        string   `xml:"minimumUpdatePeriod,attr"`
	AvailabilityStartTime      string   `xml:"availabilityStartTime,attr"`
	MediaPresentationDuration  string   `xml:"mediaPresentationDuration,attr"`
	MinBufferTime              string   `xml:"minBufferTime,attr"`
	SuggestedPresentationDelay string   `xml:"suggestedPresentationDelay,attr"`
	TimeShiftBufferDepth       string   `xml:"timeShiftBufferDepth,attr"`
	PublishTime                string   `xml:"publishTime,attr"`
	Profiles                   string   `xml:"profiles,attr"`
	XSI                        string   `xml:"xsi,attr,omitempty"`
	SCTE35                     string   `xml:"scte35,attr,omitempty"`
	XSISchemaLocation          string   `xml:"schemaLocation,attr"`
	ID                         string   `xml:"id,attr"`

	Periods []Period `xml:"Period,omitempty"`
}

func Unmarshal(ism []byte) (*MPD, error) {
	var ssm MPD
	err := xml.Unmarshal(ism, &ssm)
	if err != nil {
		return nil, err
	}
	return &ssm, nil
}

func (ism *MPD) Marshal() ([]byte, error) {
	return xml.Marshal(ism)
}

// Period represents XSD's PeriodType.
type Period struct {
	Start    string `xml:"start,attr"`
	ID       string `xml:"id,attr"`
	Duration string `xml:"duration,attr"`

	AdaptationSets []AdaptationSet `xml:"AdaptationSet,omitempty"`
}

// AdaptationSet represents XSD's AdaptationSetType.
type AdaptationSet struct {
	MimeType                string `xml:"mimeType,attr"`
	SegmentAlignment        bool   `xml:"segmentAlignment,attr"`
	StartWithSAP            uint64 `xml:"startWithSAP,attr"`
	BitstreamSwitching      bool   `xml:"bitstreamSwitching,attr"`
	SubsegmentAlignment     bool   `xml:"subsegmentAlignment,attr"`
	SubsegmentStartsWithSAP uint64 `xml:"subsegmentStartsWithSAP,attr"`
	Lang                    string `xml:"lang,attr"`
	Codecs                  string `xml:"codecs,attr"`

	Role              *Role              `xml:"Role,omitempty"`
	EssentialProperty *EssentialProperty `xml:"EssentialProperty,omitempty"`
	Representations   []Representation   `xml:"Representation,omitempty"`
	SegmentTemplate   SegmentTemplate    `xml:"SegmentTemplate,omitempty"`
}

type Role struct {
	SchemeIDURI string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

type EssentialProperty struct {
	SchemeIDURI string `xml:"schemeIdUri,attr"`
	Value       string `xml:"value,attr"`
}

// Representation represents XSD's RepresentationType.
type Representation struct {
	ID                string `xml:"id,attr"`
	Width             uint64 `xml:"width,attr"`
	Height            uint64 `xml:"height,attr"`
	SAR               string `xml:"sar,attr"`
	FrameRate         string `xml:"frameRate,attr"`
	Bandwidth         uint64 `xml:"bandwidth,attr"`
	AudioSamplingRate string `xml:"audioSamplingRate,attr"`
	Codecs            string `xml:"codecs,attr"`

	ContentProtections []Descriptor `xml:"ContentProtection,omitempty"`
}

// Descriptor represents XSD's DescriptorType.
type Descriptor struct {
	SchemeIDURI    string `xml:"schemeIdUri,attr"`
	Value          string `xml:"value,attr,omitempty"`
	CencDefaultKID string `xml:"default_KID,attr,omitempty"`
	Cenc           string `xml:"cenc,attr,omitempty"`
	Pssh           Pssh   `xml:"pssh"`
}

// Pssh represents XSD's CencPsshType .
type Pssh struct {
	Cenc  string `xml:"cenc,attr"`
	Value string `xml:",chardata"`
}

// SegmentTemplate represents XSD's SegmentTemplateType.
type SegmentTemplate struct {
	Timescale              uint64 `xml:"timescale,attr"`
	Media                  string `xml:"media,attr"`
	Initialization         string `xml:"initialization,attr"`
	StartNumber            uint64 `xml:"startNumber,attr"`
	PresentationTimeOffset uint64 `xml:"presentationTimeOffset,attr"`

	SegmentTimeline SegmentTimeline `xml:"SegmentTimeline,omitempty"`
}

// SegmentTimeline represents XSD's SegmentTimelineType
type SegmentTimeline struct {
	S []S `xml:"S,omitempty"`
}

// S represents XSD's SegmentTimeline's inner S elements.
type S struct {
	T uint64 `xml:"t,attr"`
	D uint64 `xml:"d,attr"`
	R int64  `xml:"r,attr"`
}
