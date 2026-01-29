package handlers

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jacklau/prism/internal/api/sse"
)

// SSEHandler creates an HTTP handler for SSE connections
func SSEHandler(sseService *sse.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user from auth middleware
		userID, ok := c.Locals("userID").(string)
		if !ok || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Register the client
		client, err := sseService.RegisterClient(userID, c)
		if err != nil {
			log.Printf("Failed to register SSE client: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to establish SSE connection",
			})
		}

		// Create a context that will be cancelled when the request is done
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			sseService.UnregisterClient(client.ID)
		}()

		// Stream events to the client
		// This blocks until the client disconnects
		return sseService.StreamToClient(ctx, client, c)
	}
}

// SSEStatusHandler returns the status of the SSE service
func SSEStatusHandler(sseService *sse.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"total_clients": sseService.GetClientCount(),
		})
	}
}
