package main

import (
	"blockchain-service/internal/api"
	"blockchain-service/internal/blockchain"
	"blockchain-service/internal/p2p"
	"blockchain-service/internal/utils"
	"context"
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

const(
	peersPath = "peers.json"
)

func main() {
	nodeIdx := flag.Int("nodeIdx", 0, "Node index")
	fiberPort := flag.Int("fiberPort", 3100, "API Port")
	
	err := godotenv.Load()
	if err != nil{
		log.Panicf("Failed to load environment variables")
	}

	flag.Parse()

	peers, err := utils.LoadPeers(peersPath)
	if err != nil{
		log.Panicf("Error loading peers: %v", err)
	}
	if *nodeIdx >= len(peers){
		log.Panicf("Node at index %d does not exist", *nodeIdx)
	}

	peer := peers[*nodeIdx]


	blockchain := blockchain.InitBlockChain(*nodeIdx)	
	ctx := context.Background() 

	node, err := p2p.NewBlockchainNode(
		ctx,
		"bc/1.0.0",
		peer,
		blockchain,
	)

	if err != nil{
		log.Panic(err)
	}

	pdfHandler := &api.NodeAPIHandler{
		Node: node,
	}
	
	hostPeers := make([]utils.PeerInfo, 0)
	for i, p := range peers{
		if i == *nodeIdx{
			continue 
		}
		hostPeers = append(hostPeers, p)
	}
	go node.Run(hostPeers)


	app := fiber.New()

	app.Get("/hello-world", func(c *fiber.Ctx) error {
		return c.SendString("Hello World!")
	})

	app.Post("/upload", pdfHandler.UploadHash)
	app.Get("/list", pdfHandler.GetBlocks)
	app.Get("/verify", pdfHandler.VerifyHash)
	
	app.Listen(os.Getenv("BASE_URL") + ":" + strconv.Itoa(*fiberPort))
}
