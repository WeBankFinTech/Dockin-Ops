/*
 * Copyright (C) @2021 Webank Group Holding Limited
 * <p>
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * <p>
 * http://www.apache.org/licenses/LICENSE-2.0
 * <p>
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package wsstream

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/websocket"
	"k8s.io/klog"

	"k8s.io/apimachinery/pkg/util/runtime"
)

const ChannelWebSocketProtocol = "channel.k8s.io"

const Base64ChannelWebSocketProtocol = "base64.channel.k8s.io"

type codecType int

const (
	rawCodec codecType = iota
	base64Codec
)

type ChannelType int

const (
	IgnoreChannel ChannelType = iota
	ReadChannel
	WriteChannel
	ReadWriteChannel
)

var (
	// connectionUpgradeRegex matches any Connection header value that includes upgrade
	connectionUpgradeRegex = regexp.MustCompile("(^|.*,\\s*)upgrade($|\\s*,)")
)

func IsWebSocketRequest(req *http.Request) bool {
	if !strings.EqualFold(req.Header.Get("Upgrade"), "websocket") {
		return false
	}
	return connectionUpgradeRegex.MatchString(strings.ToLower(req.Header.Get("Connection")))
}

func IgnoreReceives(ws *websocket.Conn, timeout time.Duration) {
	defer runtime.HandleCrash()
	var data []byte
	for {
		resetTimeout(ws, timeout)
		if err := websocket.Message.Receive(ws, &data); err != nil {
			return
		}
	}
}

func handshake(config *websocket.Config, req *http.Request, allowed []string) error {
	protocols := config.Protocol
	if len(protocols) == 0 {
		protocols = []string{""}
	}

	for _, protocol := range protocols {
		for _, allow := range allowed {
			if allow == protocol {
				config.Protocol = []string{protocol}
				return nil
			}
		}
	}

	return fmt.Errorf("requested protocol(s) are not supported: %v; supports %v", config.Protocol, allowed)
}

type ChannelProtocolConfig struct {
	Binary   bool
	Channels []ChannelType
}

func NewDefaultChannelProtocols(channels []ChannelType) map[string]ChannelProtocolConfig {
	return map[string]ChannelProtocolConfig{
		"":                             {Binary: true, Channels: channels},
		ChannelWebSocketProtocol:       {Binary: true, Channels: channels},
		Base64ChannelWebSocketProtocol: {Binary: false, Channels: channels},
	}
}

type Conn struct {
	protocols        map[string]ChannelProtocolConfig
	selectedProtocol string
	channels         []*websocketChannel
	codec            codecType
	ready            chan struct{}
	ws               *websocket.Conn
	timeout          time.Duration
}

func NewConn(protocols map[string]ChannelProtocolConfig) *Conn {
	return &Conn{
		ready:     make(chan struct{}),
		protocols: protocols,
	}
}

func (conn *Conn) SetIdleTimeout(duration time.Duration) {
	conn.timeout = duration
}

func (conn *Conn) Open(w http.ResponseWriter, req *http.Request) (string, []io.ReadWriteCloser, error) {
	go func() {
		defer runtime.HandleCrash()
		defer conn.Close()
		websocket.Server{Handshake: conn.handshake, Handler: conn.handle}.ServeHTTP(w, req)
	}()
	<-conn.ready
	rwc := make([]io.ReadWriteCloser, len(conn.channels))
	for i := range conn.channels {
		rwc[i] = conn.channels[i]
	}
	return conn.selectedProtocol, rwc, nil
}

func (conn *Conn) initialize(ws *websocket.Conn) {
	negotiated := ws.Config().Protocol
	conn.selectedProtocol = negotiated[0]
	p := conn.protocols[conn.selectedProtocol]
	if p.Binary {
		conn.codec = rawCodec
	} else {
		conn.codec = base64Codec
	}
	conn.ws = ws
	conn.channels = make([]*websocketChannel, len(p.Channels))
	for i, t := range p.Channels {
		switch t {
		case ReadChannel:
			conn.channels[i] = newWebsocketChannel(conn, byte(i), true, false)
		case WriteChannel:
			conn.channels[i] = newWebsocketChannel(conn, byte(i), false, true)
		case ReadWriteChannel:
			conn.channels[i] = newWebsocketChannel(conn, byte(i), true, true)
		case IgnoreChannel:
			conn.channels[i] = newWebsocketChannel(conn, byte(i), false, false)
		}
	}

	close(conn.ready)
}

func (conn *Conn) handshake(config *websocket.Config, req *http.Request) error {
	supportedProtocols := make([]string, 0, len(conn.protocols))
	for p := range conn.protocols {
		supportedProtocols = append(supportedProtocols, p)
	}
	return handshake(config, req, supportedProtocols)
}

func (conn *Conn) resetTimeout() {
	if conn.timeout > 0 {
		conn.ws.SetDeadline(time.Now().Add(conn.timeout))
	}
}

func (conn *Conn) Close() error {
	<-conn.ready
	for _, s := range conn.channels {
		s.Close()
	}
	conn.ws.Close()
	return nil
}

func (conn *Conn) handle(ws *websocket.Conn) {
	defer conn.Close()
	conn.initialize(ws)

	for {
		conn.resetTimeout()
		var data []byte
		if err := websocket.Message.Receive(ws, &data); err != nil {
			if err != io.EOF {
				klog.Errorf("Error on socket receive: %v", err)
			}
			break
		}
		if len(data) == 0 {
			continue
		}
		channel := data[0]
		if conn.codec == base64Codec {
			channel = channel - '0'
		}
		data = data[1:]
		if int(channel) >= len(conn.channels) {
			klog.V(6).Infof("Frame is targeted for a reader %d that is not valid, possible protocol error", channel)
			continue
		}
		if _, err := conn.channels[channel].DataFromSocket(data); err != nil {
			klog.Errorf("Unable to write frame to %d: %v\n%s", channel, err, string(data))
			continue
		}
	}
}

func (conn *Conn) write(num byte, data []byte) (int, error) {
	conn.resetTimeout()
	switch conn.codec {
	case rawCodec:
		frame := make([]byte, len(data)+1)
		frame[0] = num
		copy(frame[1:], data)
		if err := websocket.Message.Send(conn.ws, frame); err != nil {
			return 0, err
		}
	case base64Codec:
		frame := string('0'+num) + base64.StdEncoding.EncodeToString(data)
		if err := websocket.Message.Send(conn.ws, frame); err != nil {
			return 0, err
		}
	}
	return len(data), nil
}

type websocketChannel struct {
	conn *Conn
	num  byte
	r    io.Reader
	w    io.WriteCloser

	read, write bool
}

func newWebsocketChannel(conn *Conn, num byte, read, write bool) *websocketChannel {
	r, w := io.Pipe()
	return &websocketChannel{conn, num, r, w, read, write}
}

func (p *websocketChannel) Write(data []byte) (int, error) {
	if !p.write {
		return len(data), nil
	}
	return p.conn.write(p.num, data)
}

func (p *websocketChannel) DataFromSocket(data []byte) (int, error) {
	if !p.read {
		return len(data), nil
	}

	switch p.conn.codec {
	case rawCodec:
		return p.w.Write(data)
	case base64Codec:
		dst := make([]byte, len(data))
		n, err := base64.StdEncoding.Decode(dst, data)
		if err != nil {
			return 0, err
		}
		return p.w.Write(dst[:n])
	}
	return 0, nil
}

func (p *websocketChannel) Read(data []byte) (int, error) {
	if !p.read {
		return 0, io.EOF
	}
	return p.r.Read(data)
}

func (p *websocketChannel) Close() error {
	return p.w.Close()
}
