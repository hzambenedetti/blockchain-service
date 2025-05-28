package p2p

import (
    "bytes"
    "encoding/binary"
    "encoding/json"
    "fmt"
    "io"
		
		"blockchain-service/internal/blockchain"
)

// Message types
const (
    MsgTypeHello    = "HELLO"
    MsgTypeGossip      = "GOSSIP"
    MsgTypeGetBlock = "GETBLOCK"
    MsgTypeBlock    = "BLOCK"
    MsgTypeHi    = "HI"
    MsgTypeWhat    = "WHAT"
)

// Message is the envelope for all protocol messages
type Message struct {
    Type      string   `json:"type"`
    // HELLO fields
    ID        string   `json:"id,omitempty"`       // sender node ID
    Height    uint64   `json:"height,omitempty"`   // sender chain height
    Version   string   `json:"version,omitempty"`  // protocol version
    Peers     []string `json:"peers,omitempty"`    // list of known peers (multiaddrs)
    // INV / GETBLOCK fields
    BlockHash string   `json:"blockHash,omitempty"`
    // BLOCK field
    Block     *blockchain.Block   `json:"block,omitempty"`
}


// Constructor helpers
func NewHelloMsg(id string, height uint64, version string) *Message {
    return &Message{Type: MsgTypeHello, ID: id, Height: height, Version: version}
}
func NewGossipMsg(block *blockchain.Block, height uint64) *Message {
  return &Message{Type: MsgTypeGossip, Height: height, Block: block}
}
func NewGetBlockMsg(blockHash string) *Message {
    return &Message{Type: MsgTypeGetBlock, BlockHash: blockHash}
}
func NewBlockMsg(blk *blockchain.Block) *Message {
    return &Message{Type: MsgTypeBlock, Block: blk}
}
func NewHiMsg(id string, height uint64, version string, peers []string) *Message {
    return &Message{Type: MsgTypeHi, ID: id, Height: height, Version: version, Peers: peers}
}


// EncodeMessage serializes a Message to length-prefixed JSON
func EncodeMessage(msg *Message) ([]byte, error) {
    payload, err := json.Marshal(msg)
    if err != nil {
        return nil, fmt.Errorf("json marshal: %w", err)
    }
    buf := new(bytes.Buffer)
    if err := binary.Write(buf, binary.BigEndian, uint32(len(payload))); err != nil {
        return nil, fmt.Errorf("write length: %w", err)
    }
    if _, err := buf.Write(payload); err != nil {
        return nil, fmt.Errorf("write payload: %w", err)
    }
    return buf.Bytes(), nil
}

// DecodeNextMessage reads from an io.Reader, parses the next length-prefixed Message
func DecodeNextMessage(r io.Reader) (*Message, error) {
    var length uint32
    if err := binary.Read(r, binary.BigEndian, &length); err != nil {
        return nil, fmt.Errorf("read length: %w", err)
    }
    payload := make([]byte, length)
    if _, err := io.ReadFull(r, payload); err != nil {
        return nil, fmt.Errorf("read payload: %w", err)
    }
    var msg Message
    if err := json.Unmarshal(payload, &msg); err != nil {
        return nil, fmt.Errorf("json unmarshal: %w", err)
    }
    return &msg, nil
}

// DecodeMessageBytes is a convenience wrapper for decoding from a byte slice
func DecodeMessageBytes(data []byte) (*Message, error) {
    reader := bytes.NewReader(data)
    return DecodeNextMessage(reader)
}
