package api

import (
	"blockchain-service/internal/models"
	"blockchain-service/internal/p2p"

	"github.com/gofiber/fiber/v2"
)

type NodeAPIHandler struct {
    Node *p2p.BlockchainNode
}

func (h *NodeAPIHandler) UploadHash(c *fiber.Ctx) error{
	var blockDataAPI models.BlockDataAPI 
	if err := c.BodyParser(blockDataAPI); err != nil{
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}
	
	blockData, err := blockDataAPI.ToBlockData()
	if err != nil{
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}

	block, err := h.Node.AddBlockAPI(blockData)

	if err != nil{
		return c.SendStatus(500)
	}

	blockAPI := models.FromBlock(block)

	return c.Status(201).JSON(blockAPI)
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
