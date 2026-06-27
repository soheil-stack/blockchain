// Package core
package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Transaction struct {
	ChainID uint64         `json:"chainID"`
	Nonce   uint64         `json:"nonce"`
	From    common.Address `json:"from"`
	To      common.Address `json:"to"`
	Value   uint64         `json:"value"`
	Tip     uint64         `json:"tip"`
	Data    []byte         `json:"data"`

	V *big.Int `json:"v"`
	R *big.Int `json:"r"`
	S *big.Int `json:"s"`

	GasPrice uint64 `json:"gasPrice"`
	GasUnits uint64 `json:"gasUnits"`
}

func NewTransaction(chainID uint64, nonce uint64, from, to common.Address, value, tip uint64, data []byte) *Transaction {
	return &Transaction{
		ChainID: chainID,
		Nonce:   nonce,
		From:    from,
		To:      to,
		Value:   value,
		Tip:     tip,
		Data:    data,
	}
}

func (tx *Transaction) SigHash() common.Hash {
	buf := new(bytes.Buffer)

	_ = binary.Write(buf, binary.BigEndian, tx.ChainID)
	_ = binary.Write(buf, binary.BigEndian, tx.Nonce)
	buf.Write(tx.From.Bytes())
	buf.Write(tx.To.Bytes())
	_ = binary.Write(buf, binary.BigEndian, tx.Value)
	_ = binary.Write(buf, binary.BigEndian, tx.Tip)
	buf.Write(tx.Data)

	data := buf.Bytes()
	stamp := fmt.Sprintf("\x19Soheil Signed Message:\n%d", len(data))

	hash := crypto.Keccak256Hash([]byte(stamp), data)

	return hash
}

func (tx *Transaction) Sign(prv *ecdsa.PrivateKey) error {
	txHash := tx.SigHash()

	sig, err := crypto.Sign(txHash[:], prv)
	if err != nil {
		return err
	}

	err = tx.SetSignature(sig)
	if err != nil {
		return err
	}

	return nil
}

func (tx *Transaction) SetSignature(sig []byte) error {
	if len(sig) != 65 {
		return errors.New("invalid signature length")
	}

	tx.R = new(big.Int).SetBytes(sig[:32])
	tx.S = new(big.Int).SetBytes(sig[32:64])
	tx.V = new(big.Int).SetBytes([]byte{sig[64]})

	return nil
}

func (tx *Transaction) Signature() ([]byte, error) {
	if tx.R == nil || tx.S == nil || tx.V == nil {
		return nil, errors.New("signature is nil")
	}

	sig := make([]byte, 65)
	tx.R.FillBytes(sig[:32])
	tx.S.FillBytes(sig[32:64])
	tx.V.FillBytes(sig[64:])

	return sig, nil
}

func (tx *Transaction) Verify(chainID uint64) error {
	signature, err := tx.Signature()
	if err != nil {
		return err
	}

	if tx.ChainID != chainID {
		return errors.New("chainID is invalid")
	}

	if tx.From == tx.To {
		return errors.New("sending money to yourself")
	}

	if !crypto.ValidateSignatureValues(byte(tx.V.Uint64()), tx.R, tx.S, false) {
		return errors.New("signature is invalid")
	}

	txHash := tx.SigHash()

	pubkey, err := crypto.SigToPub(txHash[:], signature)
	if err != nil {
		return err
	}

	if tx.From != crypto.PubkeyToAddress(*pubkey) {
		return errors.New("signature address doesn't match from address")
	}

	return nil
}

func (tx *Transaction) Hash() [32]byte {
	data, err := json.Marshal(tx)
	if err != nil {
		return common.Hash{}
	}

	return sha256.Sum256(data)
}
