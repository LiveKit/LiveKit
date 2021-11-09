package rtc

import (
	"errors"
	"net"

	"github.com/livekit/livekit-server/pkg/sfu/buffer"
	"github.com/pion/ice/v2"
	"github.com/pion/webrtc/v3"

	"github.com/livekit/livekit-server/pkg/config"
	serverlogger "github.com/livekit/livekit-server/pkg/logger"
)

const (
	minUDPBufferSize     = 5_000_000
	defaultUDPBufferSize = 16_777_216
)

type WebRTCConfig struct {
	Configuration  webrtc.Configuration
	SettingEngine  webrtc.SettingEngine
	Receiver       ReceiverConfig
	BufferFactory  *buffer.Factory
	UDPMux         ice.UDPMux
	UDPMuxConn     *net.UDPConn
	TCPMuxListener *net.TCPListener
}

type ReceiverConfig struct {
	packetBufferSize int
	maxBitrate       uint64
}

// number of packets to buffer up
const readBufferSize = 50

func NewWebRTCConfig(conf *config.Config, externalIP string) (*WebRTCConfig, error) {
	rtcConf := conf.RTC
	c := webrtc.Configuration{
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	}
	s := webrtc.SettingEngine{
		LoggerFactory: serverlogger.LoggerFactory(),
	}

	if externalIP != "" {
		s.SetNAT1To1IPs([]string{externalIP}, webrtc.ICECandidateTypeHost)
	}

	if rtcConf.PacketBufferSize == 0 {
		rtcConf.PacketBufferSize = 500
	}

	var udpMux *ice.UDPMuxDefault
	var udpMuxConn *net.UDPConn
	var err error
	networkTypes := make([]webrtc.NetworkType, 0, 4)

	if !rtcConf.ForceTCP {
		networkTypes = append(networkTypes,
			webrtc.NetworkTypeUDP4,
		)
		if rtcConf.ICEPortRangeStart != 0 && rtcConf.ICEPortRangeEnd != 0 {
			if err := s.SetEphemeralUDPPortRange(uint16(rtcConf.ICEPortRangeStart), uint16(rtcConf.ICEPortRangeEnd)); err != nil {
				return nil, err
			}
		} else if rtcConf.UDPPort != 0 {
			udpMuxConn, err = net.ListenUDP("udp4", &net.UDPAddr{
				Port: int(rtcConf.UDPPort),
			})
			if err != nil {
				return nil, err
			}
			_ = udpMuxConn.SetReadBuffer(defaultUDPBufferSize)
			_ = udpMuxConn.SetWriteBuffer(defaultUDPBufferSize)

			udpMux = ice.NewUDPMuxDefault(ice.UDPMuxParams{
				Logger:  s.LoggerFactory.NewLogger("udp_mux"),
				UDPConn: udpMuxConn,
			})
			s.SetICEUDPMux(udpMux)
			if !conf.Development {
				checkUDPReadBuffer()
			}
		}
	}

	// use TCP mux when it's set
	var tcpListener *net.TCPListener
	if rtcConf.TCPPort != 0 {
		networkTypes = append(networkTypes,
			webrtc.NetworkTypeTCP4,
		)
		tcpListener, err = net.ListenTCP("tcp4", &net.TCPAddr{
			Port: int(rtcConf.TCPPort),
		})
		if err != nil {
			return nil, err
		}

		tcpMux := webrtc.NewICETCPMux(
			s.LoggerFactory.NewLogger("tcp_mux"),
			tcpListener,
			readBufferSize,
		)
		s.SetICETCPMux(tcpMux)
	}

	if len(networkTypes) == 0 {
		return nil, errors.New("TCP is forced but not configured")
	}
	s.SetNetworkTypes(networkTypes)

	return &WebRTCConfig{
		Configuration: c,
		SettingEngine: s,
		Receiver: ReceiverConfig{
			packetBufferSize: rtcConf.PacketBufferSize,
			maxBitrate:       rtcConf.MaxBitrate,
		},
		UDPMux:         udpMux,
		UDPMuxConn:     udpMuxConn,
		TCPMuxListener: tcpListener,
	}, nil
}

func (c *WebRTCConfig) SetBufferFactory(factory *buffer.Factory) {
	c.BufferFactory = factory
	c.SettingEngine.BufferFactory = factory.GetOrNew
}
