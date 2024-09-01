package main

import (
	"client-rtsp/packet"
	"fmt"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
	"log"
	"os"
	"sync"
)

type serverHandler struct {
	s         *gortsplib.Server
	mutex     sync.Mutex
	stream    *gortsplib.ServerStream
	publisher *gortsplib.ServerSession
}

// called when a connection is opened.
func (sh *serverHandler) OnConnOpen(ctx *gortsplib.ServerHandlerOnConnOpenCtx) {
	log.Printf("conn opened")
}

// called when a connection is closed.
func (sh *serverHandler) OnConnClose(ctx *gortsplib.ServerHandlerOnConnCloseCtx) {
	log.Printf("conn closed (%v)", ctx.Error)
}

// called when a session is opened.
func (sh *serverHandler) OnSessionOpen(ctx *gortsplib.ServerHandlerOnSessionOpenCtx) {
	log.Printf("session opened")
}

// called when a session is closed.
func (sh *serverHandler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	log.Printf("session closed")

	//sh.mutex.Lock()
	//defer sh.mutex.Unlock()

	// if the session is the publisher,
	// close the stream and disconnect any reader.
	//if sh.stream != nil && ctx.Session == sh.publisher {
	//	sh.stream.Close()
	//	sh.stream = nil
	//}
}

// called when receiving a DESCRIBE request.
func (sh *serverHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("describe request")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	// send medias that are being published to the client
	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

// called when receiving an ANNOUNCE request.
func (sh *serverHandler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("announce request")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// disconnect existing publisher
	if sh.stream != nil {
		sh.stream.Close()
		sh.publisher.Close()
	}

	// create the stream and save the publisher
	sh.stream = gortsplib.NewServerStream(sh.s, ctx.Description)
	sh.publisher = ctx.Session

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a SETUP request.
func (sh *serverHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("setup request")

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

// called when receiving a PLAY request.
func (sh *serverHandler) OnPlay(ctx *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("play request")

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a RECORD request.
func (sh *serverHandler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	log.Printf("record request")

	// called when receiving a RTP packet
	ctx.Session.OnPacketRTPAny(func(medi *description.Media, forma format.Format, pkt *rtp.Packet) {
		// route the RTP packet to all readers
		sh.stream.WritePacketRTP(medi, pkt)
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

func main() {
	go func() {
		h264 := packet.NewH264PacketGenerator()
		f, err := os.Open("sample.mp4")
		if err != nil {
			log.Fatal(err.Error())
		}
		defer f.Close()
		err = h264.Read(f)
		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	// configure the server
	h := &serverHandler{}
	h.s = &gortsplib.Server{
		Handler:           h,
		RTSPAddress:       ":8554",
		UDPRTPAddress:     ":8000",
		UDPRTCPAddress:    ":8001",
		MulticastIPRange:  "224.1.0.0/16",
		MulticastRTPPort:  8002,
		MulticastRTCPPort: 8003,
	}

	go func() {
		h264 := packet.NewH264PacketGenerator()
		f, err := os.Open("sample.mp4")
		if err != nil {
			log.Fatal(err.Error())
		}
		defer f.Close()
		err = h264.Read(f)
		if err != nil {
			log.Fatal(err.Error())
		}
		url, err := base.ParseURL("rtsp://localhost/stream")
		if err != nil {
			log.Fatal(err.Error())
		}
		media := description.Media{
			ID:   "1",
			Type: description.MediaTypeVideo,
			Formats: []format.Format{&format.H264{
				PayloadTyp: packet.PayloadTypeH264,
				PPS:        h264.PPS,
				SPS:        h264.SPS,
			}},
		}
		session := &description.Session{
			BaseURL:   url,
			Title:     "My Stream",
			FECGroups: make([]description.SessionFECGroup, 0),
			Medias: []*description.Media{
				&media,
			},
		}
		h.stream = gortsplib.NewServerStream(h.s, session)
		for {
			pkt, err := h264.GetNextPacket()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(pkt)
			err = h.stream.WritePacketRTP(&media, pkt)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	// start server and wait until a fatal error
	log.Printf("server is ready")
	panic(h.s.StartAndWait())
}
