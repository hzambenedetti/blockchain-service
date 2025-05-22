package api

import (
    "blockchain-service/internal/blockchain"
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/log"
    "crypto/sha256" 
)

type PDFHandler struct {
    EthClient *blockchain.EthereumClient
}

func (h *PDFHandler) UploadPDF(c *fiber.Ctx) error{
  fileHeader, err:= c.FormFile("pdf")
  if err != nil{
    log.Error("Failed to Generate file from ")
    return c.Status(400).JSON(fiber.Map{"error": err.Error()})
  }

  fileHandler, err  := fileHeader.Open()
  if err != nil{
    log.Error("Failed to open file")
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
  }
  
  buffer := make([]byte, fileHeader.Size)
  bytesRead , err := fileHandler.Read(buffer[:])
  if err != nil{
    log.Error("Failed to read file contents")
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
  }
  log.Info("Read %d bytes from %s", bytesRead, fileHeader.Filename)

  fileHash := sha256.Sum256(buffer) 
  txHash, err := h.EthClient.StoreHash(string(fileHash[:]))
  if err != nil{
    log.Error("failed")
  }

  //Store transaction details in the database 

  return c.Status(201).JSON(fiber.Map{"txHash": txHash})
}

func (h *PDFHandler) VerifyPDF(c *fiber.Ctx) error{
	//TODO
	return nil
}
