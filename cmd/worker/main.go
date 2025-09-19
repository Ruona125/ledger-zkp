package main

import (
	"context"
	"encoding/json"
	"log"

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
	svc := ledger.NewService(database)

	_, err = bus.Subscribe(events.SubjectTxCreated, func(data []byte) {
		var evt events.TxCreated
		if err := json.Unmarshal(data, &evt); err != nil {
			log.Println("bad event:", err)
			return
		}
		entryID := util.RandID()
		if err := svc.ApplyTx(context.Background(), evt.AccountID, evt.IdempotencyKey, entryID, evt.Amount); err != nil {
			log.Println("apply error:", err)
			return
		}
		log.Printf("applied %s amount=%d to account=%s\n", evt.EventID, evt.Amount, evt.AccountID)
	})
	if err != nil {
		panic(err)
	}

	log.Println("worker running…")
	select {}
}
