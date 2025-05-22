package blockchain

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "math/big"
		"os"
)

type EthereumClient struct{
	Client *ethclient.Client
}

func NewEthereumClient(nodeURL string) (*EthereumClient, error) {
    client, err := ethclient.Dial(nodeURL)
    if err != nil {
        return nil, err
    }
    return &EthereumClient{Client: client}, nil
}

func (ec *EthereumClient) StoreHash(hash string) (string, error) {
    if hash == "" {
        return "", fmt.Errorf("hash cannot be empty")
    }

    //Convert private key to ECDSA format
    privateKey, err := crypto.HexToECDSA(os.Getenv("PRIVATE_KEY"))
    if err != nil {
        return "", fmt.Errorf("invalid private key: %v", err)
    }

    //Get public address from private key
    publicKey := privateKey.Public()
    publicAddress := crypto.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey))

    //Get the nonce (transaction count) for the account
    nonce, err := ec.Client.PendingNonceAt(context.Background(), publicAddress)
    if err != nil {
        return "", fmt.Errorf("failed to get nonce: %v", err)
    }

    //Get current gas price
    gasPrice, err := ec.Client.SuggestGasPrice(context.Background())
    if err != nil {
        return "", fmt.Errorf("failed to get gas price: %v", err)
    }

    //Create the transaction
    tx := types.NewTransaction(
        nonce,
        common.Address{}, 
        big.NewInt(0),    
        100000,          
        gasPrice,        
        []byte(hash), 
    )

    //Sign the transaction
    chainID, err := ec.Client.ChainID(context.Background())
    if err != nil {
        return "", fmt.Errorf("failed to get chain ID: %v", err)
    }

    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        return "", fmt.Errorf("failed to sign transaction: %v", err)
    }

    //Send the transaction
    err = ec.Client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return "", fmt.Errorf("failed to send transaction: %v", err)
    }

    //Return the transaction hash
    return signedTx.Hash().Hex(), nil
}

func (ec *EthereumClient) VerifyHash(txHashHex, expectedHash string) (bool, error) {
    txHash := common.HexToHash(txHashHex)
    tx, _, err := ec.Client.TransactionByHash(context.Background(), txHash)
    if err != nil {
        return false, fmt.Errorf("transaction not found: %v", err)
    }
    return string(tx.Data()) == expectedHash, nil
}
