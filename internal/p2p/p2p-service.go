package p2p

import (
    "bufio"
    "context"
    "fmt"
    "log" 
    "sync"

    libp2p "github.com/libp2p/go-libp2p"
    host "github.com/libp2p/go-libp2p/core/host"
    network "github.com/libp2p/go-libp2p/core/network"
    peer "github.com/libp2p/go-libp2p/core/peer"
    peerstore "github.com/libp2p/go-libp2p/core/peerstore"
    protocol "github.com/libp2p/go-libp2p/core/protocol"
    multiaddr "github.com/multiformats/go-multiaddr"
)

// PeerMessage wraps an inbound protocol Message with its sender ID
type PeerMessage struct {
    From peer.ID
    To peer.ID
    Msg  *Message
}

// P2PService manages libp2p networking and message delivery via channels
type P2PService struct {
    ctx        context.Context
    cancel     context.CancelFunc
    host       host.Host
    protocolID protocol.ID

    peerLock sync.RWMutex
    connPeers    map[peer.ID]peer.AddrInfo
    peers map[peer.ID]peer.AddrInfo

    Inbound  chan PeerMessage  // incoming messages from network
    Outbound chan *PeerMessage     // outgoing messages to broadcast
}

// NewP2PService constructs and configures a libp2p host listening on listenAddr
// and sets up the service, but does not start dialing peers.
func NewP2PService(parentCtx context.Context, listenAddr string, protoID string) (*P2PService, error) {
    ctx, cancel := context.WithCancel(parentCtx)
    
    h, err := libp2p.New(
        libp2p.ListenAddrStrings(listenAddr),
    )
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create libp2p host: %w", err)
    }

    svc := &P2PService{
        ctx:        ctx,
        cancel:     cancel,
        host:       h,
        protocolID: protocol.ID(protoID),
        connPeers:      make(map[peer.ID]peer.AddrInfo),
        peers:      make(map[peer.ID]peer.AddrInfo),
        Inbound:    make(chan PeerMessage, 32),
        Outbound:   make(chan *Message, 32),
    }

    // register handler for incoming streams
    h.SetStreamHandler(svc.protocolID, svc.handleStream)
    return svc, nil
}

// Start launches background tasks: dialing static peers and outbound broadcaster
func (s *P2PService) Start(staticPeers []string) {
    // Dial static peers
    for _, addrStr := range staticPeers {
        go func(a string) {
            if err := s.Connect(a); err != nil {
                // log and ignore
                fmt.Printf("[P2P] connect to %s failed: %v\n", a, err)
            }
        }(addrStr)
    }
    // Start outbound broadcaster
    go s.serveOutbound()
}

// Stop terminates the service and closes resources
func (s *P2PService) Stop() error {
    s.cancel()
    // closing host will close all listeners
    return s.host.Close()
}

// Connect adds a peer by its multiaddress, storing its AddrInfo for future use
func (s *P2PService) Connect(addrStr string) error {
    maddr, err := multiaddr.NewMultiaddr(addrStr)
    if err != nil {
        return fmt.Errorf("invalid multiaddr %s: %w", addrStr, err)
    }
    info, err := peer.AddrInfoFromP2pAddr(maddr)
    if err != nil {
        return fmt.Errorf("failed to parse AddrInfo: %w", err)
    }
    // add to peerstore
    s.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
    // track in-memory
    s.peerLock.Lock()
    s.connPeers[info.ID] = *info
    s.peerLock.Unlock()
    return nil
}

// ListPeers returns the IDs of connected peers
func (s *P2PService) ListPeers() []peer.ID {
    s.peerLock.RLock()
    defer s.peerLock.RUnlock()
    ids := make([]peer.ID, 0, len(s.peers))
    for pid := range s.peers {
        ids = append(ids, pid)
    }
    return ids
}

// serveOutbound listens on the Outbound channel and broadcasts each message
func (s *P2PService) serveOutbound() {
    for {
        select {
        case <-s.ctx.Done():
            return
        case pmsg := <-s.Outbound:
          s.handleMsg(pmsg)
            // data, err := EncodeMessage(msg)
            // if err != nil {
            //     fmt.Printf("[P2P] encode message error: %v\n", err)
            //     continue
            // }
            // s.broadcastBytes(data)
        }
    }
}

// broadcastBytes opens a fresh stream to each peer and writes the payload
func (s *P2PService) broadcastBytes(data []byte) {
    s.peerLock.RLock()
    defer s.peerLock.RUnlock()
    for _, info := range s.peers {
        go func(pi peer.AddrInfo) {
            stream, err := s.host.NewStream(s.ctx, pi.ID, s.protocolID)
            if err != nil {
                // could not open stream; skip
                return
            }
            defer stream.Close()
            // write the raw length-prefixed payload
            if _, err := stream.Write(data); err != nil {
                // skip on error
                return
            }
        }(info)
    }
}

// handleStream processes an incoming libp2p stream, decoding messages
func (s *P2PService) handleStream(stream network.Stream) {
  defer stream.Close()
  reader := bufio.NewReader(stream)
  
  msg, err := DecodeNextMessage(reader)
  if err != nil {
    // EOF or decode error ends loop
    log.Printf("Failed to decode text message %v", err)
    return
  }

  // deliver to application
  pm := PeerMessage{
    From: stream.Conn().RemotePeer(),
    To: s.host.ID(),
    Msg: msg,
  }
  switch pm.Msg.Type{
  case MsgTypeHello:
    s.handleHelloIn(&pm) 
  case MsgTypeGossip:
    s.handleGossipIn(&pm) 
  case MsgTypeGetBlock:
    s.handleGetBlockIn(&pm)  
  case MsgTypeBlock:
    s.handleBlockIn(&pm)
  }
}


func (s *P2PService) handleMsg(pmsg *PeerMessage){
  switch pmsg.Msg.Type{
    case MsgTypeGossip:
      s.broadcastMsg(pmsg)
    default:
     s.sendToPeer(pmsg) 
  }
}

func (s *P2PService) handleHelloIn(msg *PeerMessage){

}

func (s *P2PService) handleGossipIn(msg *PeerMessage){

}

func (s *P2PService) handleGetBlockIn(msg *PeerMessage){

}

func (s *P2PService) handleBlockIn(msg *PeerMessage){

}

func (s *P2PService) handleHiIn(msg *PeerMessage){

}

func (s *P2PService) sendToPeer(msg *PeerMessage){

}

func (s *P2PService) broadcastMsg(msg *PeerMessage){

}

// func (s *P2PService) handleGossipOut(msg *Message){
//
// }
//
// func (s *P2PService) handleGetBlockOut(msg *Message){
//
// }
//
// func (s *P2PService) handleBlockOut(msg *Message){
//
// }
