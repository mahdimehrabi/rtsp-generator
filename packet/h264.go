package packet

import (
	"errors"
	"fmt"
	"github.com/abema/go-mp4"
	"github.com/pion/rtp"
	"io"
)

var NotH264Error = errors.New("codec is not h264")

//check if its h264 using avc1
//get stsz
//extract mdat
//encapsulate each packet using stsz into nal unit
// extract sps and pps

type H264PacketGenerator struct {
	trakNum uint8 //lots of times its 0 or 1
	n       uint
}

func NewH264PacketGenerator() *H264PacketGenerator {
	return &H264PacketGenerator{}
}

func (p *H264PacketGenerator) Read(rs io.ReadSeeker) error {
	if err := p.GetTrackInfo(rs); err != nil {
		return err
	}
	return nil
}

func (p *H264PacketGenerator) GetNextPacket() (*rtp.Packet, error) {
	return nil, NotH264Error
}

func (p *H264PacketGenerator) GetTrackInfo(rs io.ReadSeeker) error {
	boxes, err := mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeHdlr()})
	if err != nil {
		fmt.Println(err.Error())
	}
	for i, box := range boxes {
		fmt.Println(box.Info.Type)
		hdlr := box.Payload.(*mp4.Hdlr)
		if hdlr.HandlerType == VideoHandlerType {
			p.trakNum = uint8(i)
			fmt.Printf("track %d is video h264\n", i)
			break
		}
		return nil
	}

	boxes, err = mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAvc1()})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("---get avc1---")
	for _, box := range boxes {
		fmt.Println(box.Info.Type)
		vse := box.Payload.(*mp4.VisualSampleEntry)
		fmt.Println(vse)
		return nil
	}
	return NotH264Error
}
