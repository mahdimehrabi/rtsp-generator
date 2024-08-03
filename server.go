package main

import (
	"fmt"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/pion/rtp"
	"log"
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

func (sh *serverHandler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	log.Printf("session closed (%v)", ctx.Error)

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if sh.stream != nil && ctx.Session == sh.publisher {
		sh.stream.Close()
		sh.stream = nil
	}
}

func (sh *serverHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("session closed")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}
	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

func (sh *serverHandler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("announce request")
	sh.mutex.Lock()
	defer sh.mutex.Unlock()
	if sh.stream == nil {
		sh.stream.Close()
		sh.publisher.Close()
	}

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
func (sh *serverHandler) OnPlay(_ *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("play request")

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a RECORD request.
func (sh *serverHandler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	fmt.Printf("record request")

	ctx.Session.OnPacketRTPAny(func(media *description.Media, format format.Format, packet *rtp.Packet) {
		sh.stream.WritePacketRTP(media, packet)
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

func main() {
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

	// start server and wait until a fatal error
	log.Printf("server is ready")
	panic(h.s.StartAndWait())
}
