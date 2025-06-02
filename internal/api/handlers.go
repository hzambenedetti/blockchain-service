package api

import (
	"blockchain-service/internal/blockchain"
	"blockchain-service/internal/p2p"

	"github.com/gofiber/fiber/v2"
)

type NodeAPIHandler struct {
    Node *p2p.BlockchainNode
}

func (h *NodeAPIHandler) UploadHash(c *fiber.Ctx) error{
	var blockData blockchain.BlockData 
	if err := c.BodyParser(blockData); err != nil{
		return c.SendStatus(401)
	}

	block, err := h.Node.AddBlockAPI(&blockData)

	if err != nil{
		return c.SendStatus(500)
	}

	return c.Status(201).JSON(block)
}

func (h *NodeAPIHandler) VerifyHash(c *fiber.Ctx) error{
	//TODO
	return nil
}

func (h *NodeAPIHandler) GetBlocks(c *fiber.Ctx) error{
	blocks, err := h.Node.ListBlocksAPI()
	if err != nil{
		c.SendStatus(500)
	}

	return c.Status(fiber.StatusOK).JSON(blocks)
}
