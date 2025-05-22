package main

import(
	"github.com/gofiber/fiber/v2"
	"blockchain-service/internal/blockchain"
	"blockchain-service/internal/api"
	"os"
)

func main() {
	ethClient, err := blockchain.NewEthereumClient(os.Getenv("NODE_URL"))
	if err != nil{
		return
	}	

	pdfHandler := &api.PDFHandler{
		EthClient: ethClient,	
	}

	app := fiber.New()

	app.Get("/hello-world", func(c *fiber.Ctx) error {
		return c.SendString("Hello World!")
	})

	app.Post("/upload", pdfHandler.UploadPDF)
	app.Post("/verify", pdfHandler.VerifyPDF)

	app.Listen(":3000")
}
