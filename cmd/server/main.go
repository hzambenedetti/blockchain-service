package main

import (
	"blockchain-service/internal/api"
	"blockchain-service/internal/blockchain"
	"blockchain-service/internal/p2p"
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
)

var staticPeers = [1]string{"localhost:3002"}

func main() {
	
	blockchain := blockchain.InitBlockChain()	
	ctx := context.Background() 

	node, err := p2p.NewBlockchainNode(
		ctx,
		"0",
		"1",
		"/ip4/0.0.0.0/udp/9000/quic-v1",
		staticPeers[:],
		blockchain,
	)

	if err != nil{
		log.Panic(err)
	}

	pdfHandler := &api.NodeAPIHandler{
		Node: node,
	}

	go node.Run()


	app := fiber.New()

	app.Get("/hello-world", func(c *fiber.Ctx) error {
		return c.SendString("Hello World!")
	})

	app.Post("/upload", pdfHandler.UploadHash)
	app.Get("/list", pdfHandler.GetBlocks)

	app.Listen(":3000")
}
