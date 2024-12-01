package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"diy.blockchain.org/m/logger"
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
	Nodes               map[string]bool
}

// NewBlockchain initializes a new blockchain
func NewBlockchain() *Blockchain {
	genesisBlock := Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{}, // Initial empty transactions
		PreviousHash: "0000",
		Proof:        100, // A valid proof for the genesis block
		Hash:         "",  // Hash will be computed later
	}

	bc := &Blockchain{
		Chain:               []Block{},
		CurrentTransactions: []Transaction{},
		Nodes:               make(map[string]bool),
	}

	// Compute the hash for the genesis block and add it to the chain
	genesisBlock.Hash = bc.Hash(genesisBlock)
	bc.Chain = append(bc.Chain, genesisBlock)

	return bc
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

// ValidChain checks if a given blockchain is valid
func (bc *Blockchain) ValidChain(chain []Block) bool {
	// Validate genesis block separately
	if len(chain) == 0 {
		logger.Errorf("Chain is empty")
		return false
	}
	genesisBlock := chain[0]
	if genesisBlock.Hash != bc.Hash(genesisBlock) {
		logger.Errorf("Genesis block hash mismatch: expected %s, got %s", bc.Hash(genesisBlock), genesisBlock.Hash)
		return false
	}
	logger.Infof("Genesis block validated: %s", genesisBlock.Hash)

	// Validate subsequent blocks
	for i := 1; i < len(chain); i++ {
		block := chain[i]
		prevBlock := chain[i-1]

		if block.PreviousHash != prevBlock.Hash {
			logger.Errorf("Block %d has incorrect previous hash: expected %s, got %s", i, prevBlock.Hash, block.PreviousHash)
			return false
		}
		if block.Hash != bc.Hash(block) {
			logger.Errorf("Block %d has incorrect hash: expected %s, got %s", i, bc.Hash(block), block.Hash)
			return false
		}
		if !bc.ValidProof(prevBlock.Proof, block.Proof, block.PreviousHash) {
			logger.Errorf("Block %d has invalid proof of work", i)
			return false
		}
		logger.Infof("Block %d validated: %s", i, block.Hash)
	}
	return true
}

// RegisterNode adds a new node to the list of nodes
func (bc *Blockchain) RegisterNode(address string) {
	bc.Nodes[address] = true
}

// ResolveConflicts is our Consensus Algorithm
func (bc *Blockchain) ResolveConflicts() bool {
	var newChain []Block
	maxLength := len(bc.Chain)

	for node := range bc.Nodes {
		// Fetch the chain from the node
		response, err := http.Get(fmt.Sprintf("http://%s/chain", node))
		if err != nil || response.StatusCode != http.StatusOK {
			// If there is an error, skip this node
			continue
		}

		defer response.Body.Close()
		var result struct {
			Length int     `json:"length"`
			Chain  []Block `json:"chain"`
		}
		err = json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			// If we can't decode the chain, skip this node
			continue
		}

		// Log the received chain length and verify chain validity
		logger.Infof("Received chain from node %s with length: %d", node, result.Length)

		// Verify if the chain is valid and longer than the current one
		if result.Length > maxLength {
			logger.Infof("Chain is longer than current chain. Verifying validity...")
			if bc.ValidChain(result.Chain) {
				// Found a longer valid chain, replace the current chain
				maxLength = result.Length
				newChain = result.Chain
				logger.Infof("New longer valid chain found, replacing current chain.")
			} else {
				logger.Infof("Received chain is invalid. Skipping replacement.")
			}
		}
	}

	// If a new chain was found, replace the current chain
	if len(newChain) > 0 {
		logger.Infof("Replacing chain with new chain of length %d", len(newChain))
		bc.Chain = newChain
		return true
	}

	logger.Infof("No valid longer chain found. No replacement made.")
	return false
}
