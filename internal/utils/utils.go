package utils

import (
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

type PeerInfo struct {
	ID      string `json:"id"`
	PrivKey string `json:"privKey"` // base64 encoded
	Address string `json:"address"` // multiaddress
}

func (p *PeerInfo) ToPeerID() (peer.ID, error){
	peerInfo, err := p.ToAddrInfo() 
	if err != nil{
		return "", nil 
	}

	return peerInfo.ID, nil
}

func (p *PeerInfo) ToAddrInfo() (*peer.AddrInfo, error){
	peerAddr , err := multiaddr.NewMultiaddr(p.Address + "/p2p/" + p.ID)
	if err != nil{
		return &peer.AddrInfo{}, err
	}
	
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil{
		return peerInfo, err 
	}
	return peerInfo, nil
}

// LoadPeers loads peer information from the given JSON file.
func LoadPeers(filename string) ([]PeerInfo, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var peers []PeerInfo
	if err := json.Unmarshal(data, &peers); err != nil {
		return nil, err
	}
	return peers, nil
}

// UnmarshalPrivateKey decodes a base64-encoded private key.
func UnmarshalPrivateKey(encoded string) (crypto.PrivKey, error) {
	privBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	return crypto.UnmarshalPrivateKey(privBytes)
}

