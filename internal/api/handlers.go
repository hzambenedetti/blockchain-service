package api

import (
	"blockchain-service/internal/blockchain"
	"crypto/sha256"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
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
  
  buffer, err := getFileContents(fileHeader)
  if err != nil{
    log.Error("Failed to read file contents")
    return c.Status(500).JSON(fiber.Map{"error": err.Error()})
  }

  fileHash := sha256.Sum256(buffer) 
  txHash, err := h.EthClient.StoreHash(string(fileHash[:]))
  if err != nil{
    log.Error("Failed to store Hash on the database")
  }

  //Store transaction details in the database 

  return c.Status(201).JSON(fiber.Map{"txHash": txHash})
}

func (h *PDFHandler) VerifyPDF(c *fiber.Ctx) error{
	//TODO
	return nil
}

func getFileContents(header *multipart.FileHeader) ([]byte, error){
  fileHandler, err  := header.Open()
  defer fileHandler.Close()

  if err != nil{
    log.Error("Failed to open file")
    return nil, err
  }
  
  buffer := make([]byte, header.Size)
  bytesRead , err := fileHandler.Read(buffer[:])
  if err != nil{
    log.Error("Failed to read file contents")
    return nil, err
  }
  log.Info("Read %d bytes from %s", bytesRead, header.Filename)

  return buffer, nil
}
