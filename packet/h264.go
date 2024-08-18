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
	n uint
}

func NewH264PacketGenerator() *H264PacketGenerator {
	return &H264PacketGenerator{}
}

func (p *H264PacketGenerator) Read(rs io.ReadSeeker) error {

	return nil
}

func (p *H264PacketGenerator) GetNextPacket() (*rtp.Packet, error) {
	return nil, NotH264Error
}

func (p *H264PacketGenerator) CheckCodec(rs io.ReadSeeker) error {
	boxes, err := mp4.ExtractBoxWithPayload(rs, nil, mp4.BoxPath{mp4.BoxTypeMoov(), mp4.BoxTypeTrak(),
		mp4.BoxTypeMdia(), mp4.BoxTypeMinf(), mp4.BoxTypeStbl(), mp4.BoxTypeStsd(), mp4.BoxTypeAvc1()})
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("---get avc1---")
	for _, box := range boxes {
		fmt.Println(box.Info.Type)
		stsd := box.Payload.(*mp4.VisualSampleEntry)
		fmt.Println(stsd)
		fmt.Println(mp4.Stringify(box.Payload, box.Info.Context))
	}
}
