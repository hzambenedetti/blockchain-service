// package p2p
package p2p

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"blockchain-service/internal/blockchain"
	"blockchain-service/internal/utils"

)

// BlockchainNode ties together the P2P service and the blockchain logic
type BlockchainNode struct {
    ctx         context.Context
    cancel      context.CancelFunc
    chain       *blockchain.BlockChain
    p2p         *P2PService
    inbound     chan PeerMessage
    outbound    chan *PeerMessage
    version     string
}

// NewBlockchainNode constructs a new node with given parameters
func NewBlockchainNode(
    parentCtx context.Context,
    version string,
    listenAddr utils.PeerInfo,
    chain *blockchain.BlockChain,
) (*BlockchainNode, error) {
    ctx, cancel := context.WithCancel(parentCtx)
    // instantiate P2P service
    p2pSvc, err := NewP2PService(ctx, listenAddr, version)
    if err != nil {
        cancel()
        return nil, fmt.Errorf("failed to create P2P service: %w", err)
    }
    node := &BlockchainNode{
        ctx:         ctx,
        cancel:      cancel,
        chain:       chain,
        p2p:         p2pSvc,
        inbound:     p2pSvc.Inbound,
        outbound:    p2pSvc.Outbound,
        version:     version,
    }
    return node, nil
}

// Run starts the P2P service and enters the main event loop
func (n *BlockchainNode) Run(staticPeers []utils.PeerInfo) error {
    n.p2p.Start(staticPeers)
    for {
        select {
        case <-n.ctx.Done():
            return n.ctx.Err()
        case pm := <-n.inbound:
            n.handlePeerMessage(pm)
        }
    }
}

// Stop gracefully stops the node and underlying P2P service
func (n *BlockchainNode) Stop() error {
    n.cancel()
    return n.p2p.Stop()
}

// handlePeerMessage processes an incoming protocol message
func (n *BlockchainNode) handlePeerMessage(pm PeerMessage) {
    switch pm.Msg.Type {
    case MsgTypeGossip:
      n.handleGossip(&pm)
    case MsgTypeGetBlock:
      n.handleGetBlock(&pm)
    case MsgTypeBlock:
      n.handleBlock(&pm)
    default:
      n.fallbackHandler(&pm)
    }
}

func (n *BlockchainNode) handleGossip(pmsg *PeerMessage){
  n.handleBlock(pmsg)
}

func (n *BlockchainNode) handleGetBlock(pmsg *PeerMessage){
  bh := pmsg.Msg.BlockHash
  blk := n.chain.GetBlockByHash([]byte(bh))

  newMsg := PeerMessage{
    From: pmsg.To,
    To: pmsg.From,
    Msg: NewBlockMsg(blk),
  }
  n.outbound <- &newMsg 
}

func (n *BlockchainNode) handleBlock(pmsg *PeerMessage){
  block := pmsg.Msg.Block
  pow := blockchain.NewProof(block)
  if !pow.Validate(){
    log.Printf("Invalid block POW")
    return
  }
  if !bytes.Equal(block.PrevHash, n.chain.LastHash){
    log.Printf("block PrevHash != chain.LastHash")
    return 
  }

  n.chain.InsertBlock(block)

}
func (n *BlockchainNode) fallbackHandler(msg *PeerMessage){

}


// Status returns basic node status (height and peer count)
func (n *BlockchainNode) Status() (uint64, int) {
    return n.chain.Height(), len(n.p2p.ListPeers())
}

func (n *BlockchainNode) AddBlockAPI(data *blockchain.BlockData) (*blockchain.Block, error){ 
	block := n.chain.CreateInsertBlock(data)
  n.outbound <- &PeerMessage{Msg: NewGossipMsg(block, n.chain.Height())}
	return block, nil
}

func (n *BlockchainNode) ListBlocksAPI() ([]*blockchain.Block, error){
	return n.chain.ListBlocks(), nil	
}

func (n *BlockchainNode) ContainsFileHashAPI(hash []byte) bool{ 
	return n.chain.ContainsFileHash(hash) 
}
