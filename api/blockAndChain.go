package api

import (
	"encoding/json"
	"net/http"
	"sync"

	"diy.blockchain.org/m/blockchain"
)

// TODO: this should be persisted in a database
var bc = blockchain.NewBlockchain()

type (
	ErrorDto struct {
		Error string `json:"error"`
	}

	BlockAndChainHandler struct {
	}

	RestBlockAndChain interface {
		NewTransaction() func(http.ResponseWriter, *http.Request)
		MineBlock() func(http.ResponseWriter, *http.Request)
		GetChain() func(http.ResponseWriter, *http.Request)
		RegisterNodes() func(http.ResponseWriter, *http.Request)
		ResolveConflicts() func(http.ResponseWriter, *http.Request)
	}
)

var onceBlockAndChainHandler sync.Once
var instanceBlockAndChainHandler *BlockAndChainHandler

func BlockAndChainHandlerInstance() RestBlockAndChain {
	onceBlockAndChainHandler.Do(func() {
		instanceBlockAndChainHandler = &BlockAndChainHandler{}
	})
	return instanceBlockAndChainHandler
}

func (nt *BlockAndChainHandler) NewTransaction() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var txn blockchain.Transaction
		err := json.NewDecoder(r.Body).Decode(&txn)
		if err != nil {
			http.Error(w, "Invalid transaction data", http.StatusBadRequest)
			return
		}

		index := bc.NewTransaction(txn.Sender, txn.Recipient, txn.Amount)
		response := map[string]interface{}{
			"message":     "Transaction will be added to Block",
			"block_index": index,
		}
		RespondWithJSON(w, http.StatusCreated, response)
	}
}

func (nt *BlockAndChainHandler) MineBlock() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var previousHash string
		if lastBlock := bc.LastBlock(); lastBlock != nil {
			previousHash = lastBlock.Hash
		} else {
			previousHash = "0" // Genesis block case
		}

		// Mine the block with the pending transactions
		newBlock := bc.NewBlock(previousHash)
		response := map[string]interface{}{
			"message": "New Block Forged",
			"block":   newBlock,
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

func (nt *BlockAndChainHandler) GetChain() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		response := map[string]interface{}{
			"chain":  bc.Chain,
			"length": len(bc.Chain),
		}
		RespondWithJSON(w, http.StatusOK, response)
	}
}

func (nt *BlockAndChainHandler) RegisterNodes() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload struct {
			Nodes []string `json:"nodes"`
		}

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil || len(payload.Nodes) == 0 {
			http.Error(w, "Invalid request payload. Please supply a valid list of nodes.", http.StatusBadRequest)
			return
		}

		for _, node := range payload.Nodes {
			bc.RegisterNode(node)
		}

		response := map[string]interface{}{
			"message":     "New nodes have been added",
			"total_nodes": bc.Nodes,
		}
		RespondWithJSON(w, http.StatusCreated, response)
	}
}

func (nt *BlockAndChainHandler) ResolveConflicts() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		replaced := bc.ResolveConflicts()

		var response map[string]interface{}
		if replaced {
			response = map[string]interface{}{
				"message":   "Our chain was replaced",
				"new_chain": bc.Chain,
			}
		} else {
			response = map[string]interface{}{
				"message": "Our chain is authoritative",
				"chain":   bc.Chain,
			}
		}

		RespondWithJSON(w, http.StatusOK, response)
	}
}
