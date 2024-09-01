package packet

import (
	"errors"
	"fmt"
	"github.com/abema/go-mp4"
	"github.com/pion/rtp"
	"io"
)

var (
	ErrorCodecNotH264  = errors.New("codec is not h264")
	ErrorTrackNotFound = errors.New("track not found(h264)")
	ErrorSTSZNotFound  = errors.New("stsz not found(h264)")
)

const (
	PayloadTypeH264 = 96
	MaxPacketSize   = 1272
)

// check if its h264 using avc1
// get stsz
// extract mdat and encapsulate each packet using stsz into nal unit <------
// extract SPS and PPS
type H264PacketGenerator struct {
	trakNum       int8     //lots of times its 0 or 1
	stsz          []uint32 //frames
	frameNum      uint32   //counter to generate packets and iterate through frames
	mdat          []byte
	PPS           []byte
	SPS           []byte
	packetCounter int //number of packets generated from current frame
}

func NewH264PacketGenerator() *H264PacketGenerator {
	return &H264PacketGenerator{
		trakNum: -1,
	}
}

func (p *H264PacketGenerator) Read(rs io.ReadSeeker) error {
	if err := p.extractMetaData(rs); err != nil {
		return err
	}
	return nil
}

func (p *H264PacketGenerator) GetNextPacket() (*rtp.Packet, error) {
	if len(p.stsz) < int(p.frameNum)+1 {
		//TODO: add ability to choose video be looped to stream or io.EOF and close stream
		p.frameNum = 0
	}
	var start uint32
	if p.frameNum > 0 {
		start = p.stsz[p.frameNum-1]
	}
	end := p.stsz[p.frameNum]
	frame := make([]byte, end-start+1)
	for i := start; i <= end; i++ {
		frame[i-start] = p.mdat[i]
	}
	pktData := make([]byte, MaxPacketSize)
	frameByteIndex := p.packetCounter * MaxPacketSize
	j := 0
	for i := frameByteIndex; i <= len(frame) && j < MaxPacketSize; i++ {
		pktData[j] = frame[i]
		j++
	}
	if j < MaxPacketSize {
		// frame ended next frame
		p.frameNum++
		p.packetCounter = 0
	} else {
		//get next packet from this frame too
		p.packetCounter++
	}

	packet := &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			Padding:        false,
			Extension:      false,
			Marker:         false,
			PayloadType:    PayloadTypeH264,
			SequenceNumber: 0,
			Timestamp:      0,
			SSRC:           0,
		},
		Payload: pktData,
	}
	return packet, nil
}

func (p *H264PacketGenerator) extractMetaData(rs io.ReadSeeker) error {
	boxes, err := mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeHdlr()})
	if err != nil {
		return ErrorTrackNotFound
	}
	for i, box := range boxes {
		fmt.Println(box.Info.Type)
		hdlr := box.Payload.(*mp4.Hdlr)
		if hdlr.HandlerType == VideoHandlerType {
			p.trakNum = int8(i)
			fmt.Printf("track %d is video h264\n", i)
			break
		}
	}
	if p.trakNum == -1 {
		return ErrorTrackNotFound
	}

	fmt.Println("---get avc1---")
	boxes, err = mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAvc1(),
		mp4.BoxTypeAvcC()})
	if err != nil {
		return ErrorCodecNotH264
	}
	avcBox, ok := boxes[0].Payload.(*mp4.AVCDecoderConfiguration)
	if !ok || len(boxes) < 1 || len(avcBox.PictureParameterSets) < 1 || len(avcBox.SequenceParameterSets) < 1 {
		return ErrorCodecNotH264
	}
	p.PPS = avcBox.PictureParameterSets[0].NALUnit
	p.SPS = avcBox.SequenceParameterSets[0].NALUnit

	boxes, err = mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(),
		mp4.BoxTypeStsz()})
	if err != nil {
		return ErrorSTSZNotFound
	}
	fmt.Println("---get stsz---")
	for i, box := range boxes {
		if i == int(p.trakNum) {
			stsz := box.Payload.(*mp4.Stsz)
			p.stsz = stsz.EntrySize
		}
	}
	if len(p.stsz) < 1 {
		return ErrorSTSZNotFound
	}

	if err = p.getMdat(rs); err != nil {
		return err
	}
	return nil
}

func (p *H264PacketGenerator) getMdat(rs io.ReadSeeker) error {
	boxes, err := mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMdat()})
	if err != nil {
		return err
	}
	fmt.Println("---get mdat---")
	if len(boxes) < 1 {
		return errors.New("no mdat box")
	}
	mdat := boxes[0].Payload.(*mp4.Mdat)
	p.mdat = mdat.Data
	return nil
}
