// package p2p
package p2p

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"

	"blockchain-service/internal/blockchain"
)

// BlockchainNode ties together the P2P service and the blockchain logic
type BlockchainNode struct {
    ctx         context.Context
    cancel      context.CancelFunc
    chain       *blockchain.BlockChain
    p2p         *P2PService
    inbound     chan PeerMessage
    outbound    chan *Message
    id          string
    version     string
    staticPeers []string
}

// NewBlockchainNode constructs a new node with given parameters
func NewBlockchainNode(
    parentCtx context.Context,
    id string,
    version string,
    listenAddr string,
    staticPeers []string,
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
        id:          id,
        version:     version,
        staticPeers: staticPeers,
    }
    return node, nil
}

// Run starts the P2P service and enters the main event loop
func (n *BlockchainNode) Run() error {
    // start networking
    n.p2p.Start(n.staticPeers)
    // send initial HELLO
    hello := NewHelloMessage(n.id, n.chain.Height(), n.version, n.staticPeers)
    n.outbound <- hello

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
    msg := pm.Msg
    switch msg.Type {
    case MsgTypeHello:
        // integrate new peers
        for _, addr := range msg.Peers {
            go n.p2p.Connect(addr)
        }
        // optionally respond
        // n.outbound <- NewHelloMessage(n.id, n.chain.Height(), n.version, n.staticPeers)

    case MsgTypeInv:
        bh := msg.BlockHash
        if !n.chain.ContainsBlock([]byte(bh)) {
            n.outbound <- NewGetBlockMessage(bh)
        }

    case MsgTypeGetBlock:
        bh := msg.BlockHash
        blk := n.chain.GetBlockByHash([]byte(bh))
        n.outbound <- NewBlockMessage(blk)

    case MsgTypeBlock:
        blk := msg.Block
        if blk == nil {
            return
        }
				pow := blockchain.NewProof(blk)

        // verify
        if !pow.Validate() {
            return
        }

				if !bytes.Equal(blk.PrevHash, n.chain.LastHash){
					return
				}	
        // insert
        n.chain.InsertBlock(blk)
        // announce
        n.outbound <- NewInvMessage(hex.EncodeToString(blk.Hash), n.chain.Height())

    default:
        // unknown type
    }
}

// Status returns basic node status (height and peer count)
func (n *BlockchainNode) Status() (uint64, int) {
    return n.chain.Height(), len(n.p2p.ListPeers())
}
//
// import(
// 	"blockchain-service/internal/blockchain"
// )
//
// type BlockChainNode struct{
// 	Address string
// 	ID string 
// 	Blockchain *blockchain.BlockChain
// 	Peers []string 
// }
//
// func (node *BlockChainNode) Run(){
// 	node.StartNode()
//
// 	//loop
// }
//
// func (node *BlockChainNode) StartNode(){
// 	//Start Socket
// 	
// 	//Connect to peers 
// }
//
// func (node *BlockChainNode) GossipBlock(block *blockchain.Block){
// 	//Iterate through peers sending the block 
// }
//
// func (node *BlockChainNode) ReceiveBlock(from string, block blockchain.Block){
// 	//Verify Proof of Work 
//
// 	//Verify if Block is not already in the blockchain
//
// 	//Insert block in the blockchain 
//
// 	//GossipBlock
// }
//
//
