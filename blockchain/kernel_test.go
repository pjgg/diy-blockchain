package blockchain_test

import (
	"testing"
	"time"

	"diy.blockchain.org/m/blockchain"
)

// TestNewBlockchain verifies that initializing a blockchain creates an empty chain with no transactions.
func TestNewBlockchain(t *testing.T) {
	bc := blockchain.NewBlockchain()
	if len(bc.Chain) != 0 {
		t.Errorf("expected chain length to be 0, got %d", len(bc.Chain))
	}
	if len(bc.CurrentTransactions) != 0 {
		t.Errorf("expected current transactions to be empty, got %d transactions", len(bc.CurrentTransactions))
	}
}

// TestNewTransaction checks that a new transaction is added correctly.
func TestNewTransaction(t *testing.T) {
	bc := blockchain.NewBlockchain()
	index := bc.NewTransaction("Alice", "Bob", 100)
	if len(bc.CurrentTransactions) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(bc.CurrentTransactions))
	}
	if bc.CurrentTransactions[0].Sender != "Alice" ||
		bc.CurrentTransactions[0].Recipient != "Bob" ||
		bc.CurrentTransactions[0].Amount != 100 {
		t.Error("transaction details do not match expected values")
	}
	expectedIndex := 1
	if index != expectedIndex {
		t.Errorf("expected next block index to be %d, got %d", expectedIndex, index)
	}
}

// TestNewBlock verifies that a new block is created, hashed, and added to the chain correctly.
func TestNewBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.NewTransaction("Alice", "Bob", 100)

	previousHash := "0000"
	block := bc.NewBlock(previousHash)

	if len(bc.Chain) != 1 {
		t.Errorf("expected chain length to be 1, got %d", len(bc.Chain))
	}
	if len(bc.CurrentTransactions) != 0 {
		t.Errorf("expected current transactions to be empty after block creation, got %d transactions", len(bc.CurrentTransactions))
	}
	if block.Index != 1 {
		t.Errorf("expected block index to be 1, got %d", block.Index)
	}
	if block.PreviousHash != previousHash {
		t.Errorf("expected previous hash to be %s, got %s", previousHash, block.PreviousHash)
	}
	if block.Hash == "" {
		t.Error("expected block hash to be non-empty")
	}
}

// TestLastBlock ensures the last block is correctly retrieved.
func TestLastBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.NewTransaction("Alice", "Bob", 100)
	block1 := bc.NewBlock("0000")
	if bc.LastBlock().Index != block1.Index {
		t.Errorf("expected last block index to be %d, got %d", block1.Index, bc.LastBlock().Index)
	}
	bc.NewTransaction("Bob", "Charlie", 50)
	block2 := bc.NewBlock(block1.Hash)
	if bc.LastBlock().Index != block2.Index {
		t.Errorf("expected last block index to be %d, got %d", block2.Index, bc.LastBlock().Index)
	}
}

// TestHash checks that the hash of a block is generated and changes if block data changes.
func TestHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	block := blockchain.Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		Transactions: []blockchain.Transaction{{Sender: "Alice", Recipient: "Bob", Amount: 100}},
		PreviousHash: "0000",
	}

	// Generate the first hash
	hash1 := bc.Hash(block)

	// Modify a transaction to simulate data change
	block.Transactions[0].Amount = 150

	// Generate a new hash after modifying the block
	hash2 := bc.Hash(block)

	if hash1 == hash2 {
		t.Error("expected different hash for modified block data")
	}
}
