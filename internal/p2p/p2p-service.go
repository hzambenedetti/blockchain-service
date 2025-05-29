package p2p

import (
	"blockchain-service/internal/utils"
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
    From *peer.AddrInfo
    To *peer.AddrInfo
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
func NewP2PService(parentCtx context.Context, hostInfo utils.PeerInfo , protoID string) (*P2PService, error) {
    ctx, cancel := context.WithCancel(parentCtx)
  	
		listenAddr, err := multiaddr.NewMultiaddr(hostInfo.Address)
		if err != nil{
			return nil, fmt.Errorf("Failed to create p2p address: %v", err)
		}

		privKey, err := utils.UnmarshalPrivateKey(hostInfo.PrivKey)
		if err != nil{
			return nil, fmt.Errorf("Failed to UnmarshalPrivateKey: %v", err)
		}

    h, err := libp2p.New(
        libp2p.ListenAddrs(listenAddr),
				libp2p.Identity(privKey),
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
        Outbound:   make(chan *PeerMessage, 32),
    }

    // register handler for incoming streams
    h.SetStreamHandler(svc.protocolID, svc.handleStream)
    return svc, nil
}

// Start launches background tasks: dialing static peers and outbound broadcaster
func (s *P2PService) Start(staticPeers []utils.PeerInfo) {
	// Dial static peers
	for _, peer := range staticPeers {
		info, err := peer.ToAddrInfo()
		if err != nil{
			fmt.Printf("Failed to parse PeerInfo into AddrInfo: %v", err)
			continue
		}
		go s.Connect(info)
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
func (s *P2PService) Connect(info *peer.AddrInfo){
	s.peerLock.Lock()
	s.peers[info.ID] = *info
	defer s.peerLock.Unlock()
	s.host.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)
	if err := s.host.Connect(s.ctx, *info); err !=nil{
		log.Printf("Could Not Connect to %s", info.ID)
		return 
	}

	s.connPeers[info.ID] = *info
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
		}
	}
}

func (s *P2PService) sendMsg(to peer.ID, msg *Message){
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	
	
	stream, err := s.host.NewStream(s.ctx, to, s.protocolID)
	if err != nil{
		log.Printf("Failed to open Stream to peer with ID %s", to)
		return 
	}
	defer stream.Close()
	
	bytes, err := EncodeMessage(msg)
	if err != nil{
		log.Printf("Failed to encode message: %v", err)
		return
	}
	
	n , err := stream.Write(bytes)
	if err != nil{
		log.Printf("Failed to write to stream: %v", err)
		return 
	}
	
	log.Printf("Sent %d bytes to %s!", n, to)
}


func (s *P2PService) broadcastMsg(msg *Message){
	s.peerLock.RLock()
	defer s.peerLock.RUnlock()
	
	data, err := EncodeMessage(msg)
	if err != nil{
		log.Printf("Failed to encode message: %v", err)
		return
	}
	
	for peerID, _ := range s.peers{
		go s.sendBytes(peerID, data)
	}
}

func (s *P2PService) sendBytes(to peer.ID, data []byte){
	stream, err := s.host.NewStream(s.ctx, to, s.protocolID)
	if err != nil{
		log.Printf("Failed to open Stream to peer with ID %s", to)
		return 
	}
	defer stream.Close()
	
	n , err := stream.Write(data)
	if err != nil{
		log.Printf("Failed to write to stream: %v", err)
		return 
	}
	
	log.Printf("Broadcasted %d bytes to %s!", n, to)
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

	fromAddrInfo, err := peer.AddrInfoFromP2pAddr(stream.Conn().RemoteMultiaddr())
	if err != nil{
		fmt.Printf("Failed to extract AddrInfo from P2PAddr: %v", err)
		return 
	}

	toAddrInfo, err := peer.AddrInfoFromP2pAddr(s.host.Addrs()[0])
	if err != nil{
		fmt.Printf("Failed to extract AddrInfo from P2PAddr: %v", err)
		return 
	}

  // deliver to application
  pm := PeerMessage{
    From: fromAddrInfo, 
    To: toAddrInfo,
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
      s.broadcastMsg(pmsg.Msg)
    default:
     s.sendMsg(pmsg.To.ID, pmsg.Msg) 
  }
}

func (s *P2PService) handleHelloIn(msg *PeerMessage){
	for _, info := range msg.Msg.Peers{
		if _, ok := s.peers[info.ID]; ok{
			continue
		}
		go s.Connect(info)
	}
	
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	if _, ok := s.peers[msg.From.ID]; !ok{
		s.host.Peerstore().AddAddr(msg.From.ID, msg.From.Addrs[0], peerstore.PermanentAddrTTL)
		s.peers[msg.From.ID] = *msg.From
	}

	peers := make([]*peer.AddrInfo, 0)
	for _, info := range s.peers{
		peers = append(peers, &info)
	}
	hiMsg := NewHiMsg("", 0, string(s.protocolID) ,peers)
	s.sendMsg(msg.To.ID, hiMsg)
}

func (s *P2PService) handleGossipIn(msg *PeerMessage){
	s.Inbound <- *msg
}

func (s *P2PService) handleGetBlockIn(msg *PeerMessage){
	s.Inbound <- *msg
}

func (s *P2PService) handleBlockIn(msg *PeerMessage){
	s.Inbound <- *msg
}

func (s *P2PService) handleHiIn(msg *PeerMessage){
	for _, info := range msg.Msg.Peers{
		if _, ok := s.peers[info.ID]; ok{
			continue
		}
		go s.Connect(info)
	}
}
