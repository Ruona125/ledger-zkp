package main

import (
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/hash/mimc"
)

type OwnershipCircuit struct {
	Secret frontend.Variable `gnark:",public=false"`
	Hash   frontend.Variable `gnark:",public"` // MiMC(secret)
}

func (c *OwnershipCircuit) Define(api frontend.API) error {
	h, _ := mimc.NewMiMC(api)
	h.Write(c.Secret)
	api.AssertIsEqual(h.Sum(), c.Hash)
	return nil
}
