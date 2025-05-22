package api

import (
    "blockchain-service/internal/blockchain"
    "github.com/gofiber/fiber/v2"
)

type PDFHandler struct {
    EthClient *blockchain.EthereumClient
}

func (h *PDFHandler) UploadPDF(c *fiber.Ctx) error{
	//TODO
	return nil
}

func (h *PDFHandler) VerifyPDF(c *fiber.Ctx) error{
	//TODO
	return nil
}
