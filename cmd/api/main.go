package main

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ruona125/ledger-zkp/internal/db"
	"github.com/ruona125/ledger-zkp/internal/events"
	"github.com/ruona125/ledger-zkp/internal/ledger"
	"github.com/ruona125/ledger-zkp/internal/util"
)

func main() {
	database := db.Open()
	bus, err := events.NewBus("nats://localhost:4222")
	if err != nil {
		panic(err)
	}

	store := ledger.NewStore(database)
	app := fiber.New()

	app.Post("/accounts", func(c *fiber.Ctx) error {
		var r struct {
			Name       string `json:"name"`
			PublicHash string `json:"public_hash"`
		}
		if err := c.BodyParser(&r); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if r.Name == "" || r.PublicHash == "" {
			return fiber.NewError(http.StatusBadRequest, "name and public_hash required")
		}
		id := util.RandID()
		if err := store.CreateAccount(context.Background(), id, r.Name, r.PublicHash); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"id": id, "public_hash": r.PublicHash})
	})

	app.Post("/transactions", func(c *fiber.Ctx) error {
		var r struct {
			AccountID      string `json:"account_id"`
			Amount         int64  `json:"amount"`
			IdempotencyKey string `json:"idempotency_key"`
		}
		if err := c.BodyParser(&r); err != nil {
			return fiber.NewError(http.StatusBadRequest, err.Error())
		}
		if r.AccountID == "" || r.Amount == 0 || r.IdempotencyKey == "" {
			return fiber.NewError(http.StatusBadRequest, "account_id, amount, idempotency_key required")
		}

		e := events.TxCreated{
			EventID:        uuid.NewString(),
			AccountID:      r.AccountID,
			Amount:         r.Amount,
			IdempotencyKey: r.IdempotencyKey,
		}
		if err := bus.Publish(events.SubjectTxCreated, e); err != nil {
			return err
		}
		return c.JSON(e)
	})

	app.Get("/balances/:id", func(c *fiber.Ctx) error {
		bal, err := store.Balance(context.Background(), c.Params("id"))
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"account_id": c.Params("id"), "balance": bal})
	})

	if err := app.Listen(":8081"); err != nil {
		panic(err)
	}
}
