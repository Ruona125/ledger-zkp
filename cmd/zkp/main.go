package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc"
	fr "github.com/consensys/gnark-crypto/ecc/bn254/fr"
	mimcfr "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/gofiber/fiber/v2"
)

// ------------------- Request/Response Structs -------------------

type commitReq struct {
	Secret string `json:"secret"` // decimal string
}
type commitResp struct {
	PublicHash string `json:"public_hash"`
}

type proveReq struct {
	Secret     string `json:"secret"`
	PublicHash string `json:"public_hash"`
}
type proveResp struct {
	ProofB64 string `json:"proof_b64"`
	VKb64    string `json:"vk_b64"`
	Public   string `json:"public_hash"`
}
type verifyReq struct {
	ProofB64   string `json:"proof_b64"`
	VKb64      string `json:"vk_b64"`
	PublicHash string `json:"public_hash"`
}

// ------------------- Helpers -------------------

// computeCommit: off-circuit MiMC(secret) → public_hash
func computeCommit(dec string) (string, error) {
	var x fr.Element
	if _, err := x.SetString(dec); err != nil {
		return "", errors.New("invalid secret (decimal)")
	}
	h := mimcfr.NewMiMC()
	b := x.Bytes()
	_, _ = h.Write(b[:])
	sum := h.Sum(nil)
	var out fr.Element
	out.SetBytes(sum)
	return out.String(), nil
}

// ------------------- Main -------------------

func main() {
	app := fiber.New()

	// Compile circuit & setup keys once
	var circuit OwnershipCircuit
	ccs, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		panic(err)
	}
	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		panic(err)
	}

	// --- Commit ---
	app.Post("/commit", func(c *fiber.Ctx) error {
		var req commitReq
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return fiber.NewError(400, err.Error())
		}
		ph, err := computeCommit(req.Secret)
		if err != nil {
			return fiber.NewError(400, err.Error())
		}
		return c.JSON(commitResp{PublicHash: ph})
	})

	// --- Prove ---
	app.Post("/prove", func(c *fiber.Ctx) error {
		var req proveReq
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return fiber.NewError(400, err.Error())
		}
		sec, ok := new(big.Int).SetString(req.Secret, 10)
		if !ok {
			return fiber.NewError(400, "invalid secret")
		}
		pub, ok := new(big.Int).SetString(req.PublicHash, 10)
		if !ok {
			return fiber.NewError(400, "invalid public_hash")
		}

		assignment := OwnershipCircuit{Secret: sec, Hash: pub}
		wit, err := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
		if err != nil {
			return err
		}

		proof, err := groth16.Prove(ccs, pk, wit)
		if err != nil {
			return err
		}

		// Serialize proof + vk using WriteTo
		var bufProof, bufVk bytes.Buffer
		if _, err := proof.WriteTo(&bufProof); err != nil {
			return err
		}
		if _, err := vk.WriteTo(&bufVk); err != nil {
			return err
		}

		return c.JSON(proveResp{
			ProofB64: base64.StdEncoding.EncodeToString(bufProof.Bytes()),
			VKb64:    base64.StdEncoding.EncodeToString(bufVk.Bytes()),
			Public:   req.PublicHash,
		})
	})

	// --- Verify ---
	app.Post("/verify", func(c *fiber.Ctx) error {
		var req verifyReq
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return fiber.NewError(400, err.Error())
		}
		pub, ok := new(big.Int).SetString(req.PublicHash, 10)
		if !ok {
			return fiber.NewError(400, "invalid public_hash")
		}

		pb, err := base64.StdEncoding.DecodeString(req.ProofB64)
		if err != nil {
			return fiber.NewError(400, "bad proof_b64")
		}
		vkb, err := base64.StdEncoding.DecodeString(req.VKb64)
		if err != nil {
			return fiber.NewError(400, "bad vk_b64")
		}

		// Deserialize proof + vk using ReadFrom
		var proof groth16.Proof
		if _, err := proof.ReadFrom(bytes.NewReader(pb)); err != nil {
			return fiber.NewError(400, "invalid proof")
		}
		var vk groth16.VerifyingKey
		if _, err := vk.ReadFrom(bytes.NewReader(vkb)); err != nil {
			return fiber.NewError(400, "invalid vk")
		}

		// Build public witness
		var pubOnly OwnershipCircuit
		pubOnly.Hash = pub
		pw, err := frontend.NewWitness(&pubOnly, ecc.BN254.ScalarField(), frontend.PublicOnly())
		if err != nil {
			return err
		}

		if err := groth16.Verify(proof, vk, pw); err != nil {
			return c.JSON(fiber.Map{"valid": false, "error": err.Error()})
		}
		return c.JSON(fiber.Map{"valid": true})
	})

	println("ZKP service running on :8082")
	if err := app.Listen(":8082"); err != nil {
		panic(err)
	}
}
