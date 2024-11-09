package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PreviousHash string        `json:"previous_hash"`
	Proof        int           `json:"proof"`
	Hash         string        `json:"hash"`
}

// Transaction represents a transaction
type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int    `json:"amount"`
}

// Blockchain represents the entire blockchain
type Blockchain struct {
	Chain               []Block
	CurrentTransactions []Transaction
}

// NewBlockchain initializes a new blockchain
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Chain:               []Block{},
		CurrentTransactions: []Transaction{},
	}
}

// NewBlock creates a new block and adds it to the chain
func (bc *Blockchain) NewBlock(previousHash string) Block {
	lastBlock := bc.LastBlock()
	lastProof := 0
	if lastBlock != nil {
		lastProof = lastBlock.Proof
	}

	proof := bc.ProofOfWork(lastProof, previousHash)

	block := Block{
		Index:        len(bc.Chain) + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: bc.CurrentTransactions,
		PreviousHash: previousHash,
		Hash:         "", // This will be filled after hashing
		Proof:        proof,
	}

	block.Hash = bc.Hash(block)
	bc.Chain = append(bc.Chain, block)
	bc.CurrentTransactions = []Transaction{} // Reset current transactions
	return block
}

// NewTransaction adds a new transaction to the list of transactions
func (bc *Blockchain) NewTransaction(sender, recipient string, amount int) int {
	transaction := Transaction{Sender: sender, Recipient: recipient, Amount: amount}
	bc.CurrentTransactions = append(bc.CurrentTransactions, transaction)
	if bc.LastBlock() == nil {
		return 1
	}

	return bc.LastBlock().Index + 1
}

// Hash creates a SHA-256 hash of a Block
func (bc *Blockchain) Hash(block Block) string {
	// Convert transactions to JSON
	transactionsJSON, err := json.Marshal(block.Transactions)
	if err != nil {
		fmt.Println("Error marshaling transactions:", err)
		return ""
	}

	record := fmt.Sprintf("%d%d%s%s", block.Index, block.Timestamp, block.PreviousHash, transactionsJSON)

	hash := sha256.New()
	hash.Write([]byte(record))
	return hex.EncodeToString(hash.Sum(nil))
}

// LastBlock returns the last Block in the chain
func (bc *Blockchain) LastBlock() *Block {
	if len(bc.Chain) == 0 {
		return nil
	}
	return &bc.Chain[len(bc.Chain)-1]
}

func (bc *Blockchain) ProofOfWork(lastProof int, previousHash string) int {
	proof := 0
	for !bc.ValidProof(lastProof, proof, previousHash) {
		proof++
	}
	return proof
}

func (bc *Blockchain) ValidProof(lastProof int, proof int, previousHash string) bool {
	guess := fmt.Sprintf("%d%d%s", lastProof, proof, previousHash)
	guessHash := sha256.Sum256([]byte(guess))
	return hex.EncodeToString(guessHash[:])[:4] == "0000"
}
