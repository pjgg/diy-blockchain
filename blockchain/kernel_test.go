package blockchain_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diy.blockchain.org/m/blockchain"
)

// TestNewBlockchain verifies that initializing a blockchain creates an empty chain with no transactions.
func TestNewBlockchain(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// The genesis block is now part of the chain
	expectedChainLength := 1
	if len(bc.Chain) != expectedChainLength {
		t.Errorf("expected chain length to be %d, got %d", expectedChainLength, len(bc.Chain))
	}

	// Current transactions should still be empty
	if len(bc.CurrentTransactions) != 0 {
		t.Errorf("expected current transactions to be empty, got %d transactions", len(bc.CurrentTransactions))
	}

	// Validate the genesis block
	genesisBlock := bc.Chain[0]
	if genesisBlock.Index != 1 {
		t.Errorf("expected genesis block index to be 1, got %d", genesisBlock.Index)
	}
	if genesisBlock.PreviousHash != "0000" {
		t.Errorf("expected genesis block previous hash to be '0000', got %s", genesisBlock.PreviousHash)
	}
	if genesisBlock.Hash == "" {
		t.Error("expected genesis block hash to be non-empty")
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

	// The expected index should match the index of the next block to be added
	expectedIndex := bc.LastBlock().Index + 1
	if index != expectedIndex {
		t.Errorf("expected next block index to be %d, got %d", expectedIndex, index)
	}
}

// TestNewBlock verifies that a new block is created, hashed, and added to the chain correctly.
func TestNewBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.NewTransaction("Alice", "Bob", 100)

	previousHash := bc.LastBlock().Hash
	block := bc.NewBlock(previousHash)

	// The chain now includes the genesis block and the new block
	expectedChainLength := 2
	if len(bc.Chain) != expectedChainLength {
		t.Errorf("expected chain length to be %d, got %d", expectedChainLength, len(bc.Chain))
	}
	if len(bc.CurrentTransactions) != 0 {
		t.Errorf("expected current transactions to be empty after block creation, got %d transactions", len(bc.CurrentTransactions))
	}
	if block.Index != expectedChainLength {
		t.Errorf("expected block index to be %d, got %d", expectedChainLength, block.Index)
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

func TestValidChain(t *testing.T) {
	bc := blockchain.NewBlockchain()

	// Create the first block
	bc.NewTransaction("Alice", "Bob", 50)
	block1 := bc.NewBlock(bc.LastBlock().Hash) // Use the correct hash of the genesis block

	// Create the second block
	bc.NewTransaction("Bob", "Charlie", 30)
	bc.NewBlock(block1.Hash) // Use the correct hash of the first block

	// Validate the entire chain
	valid := bc.ValidChain(bc.Chain)
	if !valid {
		t.Error("expected chain to be valid, got invalid")
	}

	// Tamper with the second block in the chain
	bc.Chain[1].Transactions[0].Amount = 9999 // Alter a transaction

	// Do NOT recalculate the hash for the tampered block
	// This simulates tampering without re-mining the block

	// Check again for validity
	valid = bc.ValidChain(bc.Chain)
	if valid {
		t.Error("expected chain to be invalid after tampering, got valid")
	}
}

// TestRegisterNode verifies that nodes are correctly registered.
func TestRegisterNode(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.RegisterNode("http://localhost:5001")
	bc.RegisterNode("http://localhost:5002")

	if len(bc.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(bc.Nodes))
	}

	_, exists := bc.Nodes["http://localhost:5001"]
	if !exists {
		t.Error("expected node http://localhost:5001 to be registered")
	}
	_, exists = bc.Nodes["http://localhost:5002"]
	if !exists {
		t.Error("expected node http://localhost:5002 to be registered")
	}
}

// TestResolveConflicts verifies that ResolveConflicts correctly replaces the chain if a longer valid chain is found.
func TestResolveConflicts(t *testing.T) {
	bc := blockchain.NewBlockchain()
	t.Logf("Blockchain initialized with length: %d", len(bc.Chain))

	// Create a mock chain with real transactions
	mockChain := []blockchain.Block{
		{
			Index:        1,
			Timestamp:    time.Now().Unix(),
			Transactions: []blockchain.Transaction{{Sender: "Alice", Recipient: "Bob", Amount: 10}},
			PreviousHash: "0000",
			Hash:         "",
		},
		{
			Index:        2,
			Timestamp:    time.Now().Unix(),
			Transactions: []blockchain.Transaction{{Sender: "Bob", Recipient: "Charlie", Amount: 5}},
			PreviousHash: "",
			Hash:         "",
		},
	}

	// Calculate hashes and proofs for mock chain
	mockChain[0].Hash = bc.Hash(mockChain[0]) // Hash the first block
	mockChain[0].Proof = bc.ProofOfWork(0, mockChain[0].PreviousHash)

	mockChain[1].PreviousHash = mockChain[0].Hash // Set previous hash for second block
	mockChain[1].Proof = bc.ProofOfWork(mockChain[0].Proof, mockChain[1].PreviousHash)
	mockChain[1].Hash = bc.Hash(mockChain[1]) // Hash the second block

	// Validate the proofs of the mock chain before sending it
	for i := 1; i < len(mockChain); i++ {
		previous := mockChain[i-1]
		current := &mockChain[i]
		if !bc.ValidProof(previous.Proof, current.Proof, current.PreviousHash) {
			t.Errorf("Invalid proof for block %d", current.Index)
		}
	}

	// Create a mock response from a mock server that sends the mock chain
	mockResponse := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/chain" {
			resp := map[string]interface{}{
				"length": len(mockChain),
				"chain":  mockChain,
			}
			// Log the mock response being sent
			t.Logf("Mock server sending chain with length: %d", len(mockChain))
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer mockResponse.Close()

	// Register the mock server as a node
	bc.RegisterNode(mockResponse.Listener.Addr().String())

	// Print the blockchain state before resolving conflicts
	t.Logf("Blockchain before ResolveConflicts, length: %d", len(bc.Chain))

	// Test ResolveConflicts
	replaced := bc.ResolveConflicts()

	// Assert the chain was replaced
	if !replaced {
		t.Error("expected chain to be replaced, got no replacement")
	}

	// Assert the chain length is now the same as the mock chain
	if len(bc.Chain) != len(mockChain) {
		t.Errorf("expected chain length to be %d, got %d", len(mockChain), len(bc.Chain))
	}

	// Check the actual blockchain content
	t.Logf("Blockchain after ResolveConflicts, length: %d", len(bc.Chain))
	if len(bc.Chain) > 0 {
		t.Logf("Blockchain after conflict resolution: %v", bc.Chain)
	}
}
