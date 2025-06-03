package api

import (
	"blockchain-service/internal/models"
	"blockchain-service/internal/p2p"
	"encoding/hex"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

type NodeAPIHandler struct {
    Node *p2p.BlockchainNode
}

func (h *NodeAPIHandler) UploadHash(c *fiber.Ctx) error{
	var blockDataAPI models.BlockDataAPI 
	if err := c.BodyParser(&blockDataAPI); err != nil{
		log.Errorf("Failed to parse body to BlockDataAPI type: %v", err)
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}
	
	blockData, err := blockDataAPI.ToBlockData()
	if err != nil{
		log.Errorf("Failed to convert BlockDataAPI to BlockData: %v", err)
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}

	block, err := h.Node.AddBlockAPI(blockData)

	if err != nil{
		log.Errorf("Failed to add block to blockchain %v", err)
		return c.SendStatus(500)
	}

	blockAPI := models.FromBlock(block)

	return c.Status(201).JSON(blockAPI)
}

func (h *NodeAPIHandler) VerifyHash(c *fiber.Ctx) error{
	hash := c.Query("hash", "")
	if hash == ""{
		log.Errorf("Cannot verify the existance of empty hash")
		return c.SendStatus(fiber.ErrBadRequest.Code)
	}
	
	hashBytes, err := hex.DecodeString(hash)
	if err != nil{
		log.Errorf("Failed to convert provided string to bytes: %v", err)
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"message": "The provided hash is invalid",
		})
	}

	containsHash := h.Node.ContainsFileHashAPI(hashBytes)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"result": containsHash,
	})
}

func (h *NodeAPIHandler) GetBlocks(c *fiber.Ctx) error{
	blocks, err := h.Node.ListBlocksAPI()
	if err != nil{
		c.SendStatus(500)
	}

	return c.Status(fiber.StatusOK).JSON(blocks)
}
