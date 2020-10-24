package voice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/skwair/harmony/internal/payload"
	"github.com/skwair/harmony/log"
	"github.com/skwair/harmony/version"
	"go.uber.org/atomic"
	"nhooyr.io/websocket"
)

// Connect establishes a new voice connection with the provided information. It will automatically try to
// reconnect and resume if network failures occur. Further StateUpdate and ServerUpdate should be forwarded
// to the returned connection using its SetState and UpdateServer method. Not doing so will likely result in
// out-of-sync voice connections that can be in incoherent state or that do not handle voice server update
// events.
// This connection should be closed by calling its Close method when no longer needed.
func Connect(ctx context.Context, state *StateUpdate, server *ServerUpdate, opts ...ConnectionOption) (*Connection, error) {
	if state.ChannelID == nil {
		return nil, errors.New("could not establish voice connection: channel ID in given state is nil")
	}

	vc := &Connection{
		Send:                 make(chan []byte),
		Recv:                 make(chan *AudioPacket),
		payloads:             make(chan *payload.Payload),
		error:                make(chan error),
		stop:                 make(chan struct{}),
		state:                &state.State,
		logger:               log.NewStd(os.Stderr, log.LevelError),
		lastHeartbeatAck:     atomic.NewInt64(0),
		udpHeartbeatSequence: atomic.NewUint64(0),
		lastUDPHeartbeatAck:  atomic.NewInt64(0),
		connected:            atomic.NewBool(false),
		connecting:           atomic.NewBool(false),
		reconnecting:         atomic.NewBool(false),
	}

	vc.ctx, vc.cancel = context.WithCancel(context.Background())

	for _, opt := range opts {
		opt(vc)
	}

	if err := vc.connect(ctx, server); err != nil {
		return nil, err
	}

	return vc, nil
}

// connect performs the complete voice connection handshake to the given voice server.
func (vc *Connection) connect(ctx context.Context, server *ServerUpdate) error {
	// This is used to notify the event handler that some
	// specific payloads should be sent through to vc.payloads
	// while we are connecting to the voice server.
	vc.connecting.Store(true)
	defer vc.connecting.Store(false)

	vc.token = server.Token

	// Start by opening the voice websocket connection.
	var err error
	vc.endpoint = fmt.Sprintf("wss://%s?v=%s", strings.TrimSuffix(server.Endpoint, ":80"), version.Voice())
	vc.logger.Debugf("connecting to voice server: %s", vc.endpoint)
	vc.conn, _, err = websocket.Dial(ctx, vc.endpoint, nil)
	if err != nil {
		return err
	}
	// From now on, if any error occurs during the rest of the
	// voice connection process, we should close the underlying
	// websocket so we can try to reconnect.
	defer func() {
		if err != nil {
			_ = vc.conn.Close(websocket.StatusInternalError, "failed to establish voice connection")
			vc.connected.Store(false)
			close(vc.stop)
			vc.cancel()
		}
	}()

	vc.wg.Add(1)
	go vc.listenAndHandlePayloads()

	vc.wg.Add(1)
	go vc.wait()

	// The voice server should first send us a Hello packet defining the heartbeat
	// interval when we connect to the websocket.
	p := <-vc.payloads
	if p.Op != voiceOpcodeHello {
		return fmt.Errorf("expected Opcode 8 Hello; got Opcode %d", p.Op)
	}

	var h struct {
		V                 int     `json:"v"`
		HeartbeatInterval float64 `json:"heartbeat_interval"`
	}
	if err = json.Unmarshal(p.D, &h); err != nil {
		return err
	}
	// NOTE: do not start heartbeating before sending the identify payload
	// to the voice server, else it will close the connection.

	// Identify on the websocket connection. This is the first payload we must sent to the server.
	i := &voiceIdentify{
		ServerID:  vc.State().GuildID,
		UserID:    vc.State().UserID,
		SessionID: vc.State().SessionID,
		Token:     vc.token,
	}
	vc.logger.Debug("identifying to the voice server")
	if err = vc.sendPayload(ctx, voiceOpcodeIdentify, i); err != nil {
		return err
	}

	// Now that we sent the identify payload, we can start heartbeating.
	vc.wg.Add(1)
	go vc.heartbeat(time.Duration(h.HeartbeatInterval) * time.Millisecond)

	// A Ready payload should be sent after we identified.
	p = <-vc.payloads
	if p.Op != voiceOpcodeReady {
		return fmt.Errorf("expected Opcode 2 Ready; got Opcode %d", p.Op)
	}

	var vr voiceReady
	if err = json.Unmarshal(p.D, &vr); err != nil {
		return err
	}
	vc.ssrc = vr.SSRC

	// We should now be able to open the voice UDP connection.
	host := fmt.Sprintf("%s:%d", vr.IP, vr.Port)
	vc.logger.Debug("resolving voice connection UDP endpoint")
	vc.dataEndpoint, err = net.ResolveUDPAddr("udp", host)
	if err != nil {
		return err
	}

	vc.logger.Debugf("dialing voice connection endpoint: %s", host)
	vc.udpConn, err = net.DialUDP("udp", nil, vc.dataEndpoint)
	if err != nil {
		return err
	}
	// From now on, close the UDP connection if any error occurs.
	defer func() {
		if err != nil {
			_ = vc.udpConn.Close()
		}
	}()

	// IP discovery.
	vc.logger.Debug("starting IP discovery")
	ip, port, err := ipDiscovery(vc.udpConn, vc.ssrc)
	if err != nil {
		return err
	}
	vc.logger.Debugf("IP discovery result: %s:%d", ip, port)

	// Start heartbeating on the UDP connection.
	vc.wg.Add(1)
	go vc.udpHeartbeat(5 * time.Second)

	sp := &selectProtocol{
		Protocol: "udp",
		Data: &selectProtocolData{
			Address: ip,
			Port:    port,
			Mode:    "xsalsa20_poly1305",
		},
	}
	if err = vc.sendPayload(ctx, voiceOpcodeSelectProtocol, sp); err != nil {
		return err
	}

	// Now we should receive a Session Description packet.
	p = <-vc.payloads
	if p.Op != voiceOpcodeSessionDescription {
		return fmt.Errorf("expected Opcode 4 Session Description; got Opcode %d", p.Op)
	}

	var sd sessionDescription
	if err = json.Unmarshal(p.D, &sd); err != nil {
		return err
	}

	copy(vc.secret[:], sd.SecretKey[0:32])

	vc.wg.Add(3) // opusReceiver starts an additional goroutine.
	vc.opusReadinessWG.Add(2)
	go vc.opusReceiver()
	go vc.opusSender()

	// Making sure Opus receiver and sender are started.
	vc.opusReadinessWG.Wait()

	if err = vc.sendSilenceFrame(); err != nil {
		return err
	}

	vc.connected.Store(true)

	vc.logger.Debug("connected to voice server")
	return nil
}

// wait waits for an error to happen while connected to the voice server
// or for a stop signal to be sent.
func (vc *Connection) wait() {
	defer vc.wg.Done()

	vc.logger.Debug("starting voice connection manager")
	defer vc.logger.Debug("stopped voice connection manager")

	var err error
	select {
	case err = <-vc.error:
		vc.onError(err)

	case <-vc.stop:
		vc.logger.Debug("disconnecting from the voice server")
		vc.onDisconnect()
	}

	close(vc.payloads)

	if vc.udpConn != nil {
		if err = vc.udpConn.Close(); err != nil {
			vc.logger.Errorf("failed to properly close voice UDP connection: %v", err)
		}
	}

	vc.cancel()
	vc.connected.Store(false)

	// If there was an error, maybe try to reconnect.
	if shouldReconnect(err) && !vc.isReconnecting() {
		vc.reconnectWithBackoff()
	}
}

// Must send some audio packets so the voice server starts to send us audio packets.
// This appears to be a bug from Discord.
func (vc *Connection) sendSilenceFrame() error {
	if err := vc.SetSpeakingMode(SpeakingModeMicrophone); err != nil {
		return err
	}

	vc.Send <- SilenceFrame

	if err := vc.SetSpeakingMode(SpeakingModeOff); err != nil {
		return err
	}

	return nil
}

// onError is called when an error occurs while the connection to
// the voice server is up. It closes the underlying websocket connection
// with a 1006 code, logs the error and finally signals to all other
// goroutines (heartbeat, listenAndHandlePayloads, etc.) to stop by
// closing the stop channel.
func (vc *Connection) onError(err error) {
	vc.logger.Errorf("voice connection error: %v", err)

	if closeErr := vc.conn.Close(websocket.StatusInternalError, "voice error"); closeErr != nil {
		vc.logger.Errorf("could not properly close voice websocket connection: %v", closeErr)
	}

	// If an error occurred before the connection is established,
	// the stop channel will already be closed, so return early.
	if !vc.isConnected() {
		return
	}

	close(vc.stop)
}

// onDisconnect is called when a normal disconnection happens (the client
// called the Close() method). It closes the underlying websocket
// connection with a 1000 code and resets the UDP heartbeat sequence.
func (vc *Connection) onDisconnect() {
	if err := vc.conn.Close(websocket.StatusNormalClosure, "disconnecting"); err != nil {
		vc.logger.Errorf("could not properly close voice websocket connection: %v", err)
	}
	vc.udpHeartbeatSequence.Store(0)
}

// Disconnect closes the voice connection.
func (vc *Connection) Close() {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	if !vc.isConnected() {
		return
	}

	close(vc.stop)
	vc.wg.Wait()
	// NOTE: maybe we should explicitly close
	// other channels here.
	close(vc.Recv)
}
