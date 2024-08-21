package packet

import (
	"errors"
	"fmt"
	"github.com/abema/go-mp4"
	"github.com/pion/rtp"
	"io"
)

var ErrorCodecNotH264 = errors.New("track not found(h264)")
var ErrorTrackNotFound = errors.New("codec not found(h264)")
var ErrorSTSZNotFound = errors.New("stsz not found(h264)")

//check if its h264 using avc1
//get stsz  <------
//extract mdat
//encapsulate each packet using stsz into nal unit
// extract sps and pps

type H264PacketGenerator struct {
	trakNum int8 //lots of times its 0 or 1
	stsz    []uint32
}

func NewH264PacketGenerator() *H264PacketGenerator {
	return &H264PacketGenerator{
		trakNum: -1,
	}
}

func (p *H264PacketGenerator) Read(rs io.ReadSeeker) error {
	if err := p.GetTrackInfo(rs); err != nil {
		return err
	}
	return nil
}

func (p *H264PacketGenerator) GetNextPacket() (*rtp.Packet, error) {
	return nil, ErrorCodecNotH264
}

func (p *H264PacketGenerator) GetTrackInfo(rs io.ReadSeeker) error {
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

	boxes, err = mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAvc1()})
	if err != nil {
		return ErrorCodecNotH264
	}
	fmt.Println("---get avc1---")
	if len(boxes) < 1 {
		return ErrorCodecNotH264
	}

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
	return nil
}
