package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"diy.blockchain.org/m/api"
	"diy.blockchain.org/m/configuration"
	"gopkg.in/yaml.v2"
)

var serverPort int

func TestMain(m *testing.M) {
	readyCh := make(chan bool)
	serverPort = randomServerPort()

	// load test configuration
	serverConfiguration := fmt.Sprintf("http_port: \"%d\"\n", serverPort)
	yamlData := []byte(serverConfiguration)
	if err := yaml.Unmarshal(yamlData, &configuration.InstanceConfig); err != nil {
		fmt.Printf("Failed to parse config data: %v\n", err)
	}

	// start HTTP Server
	go startServer(readyCh)
	select {
	case <-readyCh:
		fmt.Println("Server is ready")
	case <-time.After(30 * time.Second):
		fmt.Println("Timeout waiting for server to be ready")
		os.Exit(1) // Exit with an error if the server didn't start in time
	}

	// Run tests
	code := m.Run()
	// TODO: teardown
	os.Exit(code)
}

func TestNewTransaction(t *testing.T) {
	payload := []byte(`{"sender": "Alice", "recipient": "Bob", "amount": 10}`)
	url := fmt.Sprintf("http://localhost:%d/transactions/new", serverPort)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to make request to /transactions/new: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if _, ok := result["block_index"]; !ok {
		t.Errorf("Expected 'block_index' in response, got %v", result)
	}
}

func TestMineBlock(t *testing.T) {
	// Add a sample transaction before mining a block
	payload := []byte(`{"sender": "Alice", "recipient": "Bob", "amount": 10}`)
	url := fmt.Sprintf("http://localhost:%d/transactions/new", serverPort)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatalf("Failed to make request to /transactions/new: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	// Now that the transaction is created, send a request to mine a new block
	url = fmt.Sprintf("http://localhost:%d/mine", serverPort)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	t.Logf("Response: %s", string(body))
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if result["message"] != "New Block Forged" {
		t.Errorf("Expected message 'New Block Forged', got %v", result["message"])
	}

	blockData, ok := result["block"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected block data in response, got %v", result["block"])
	}

	// The index of the first block should be 2 (since genesis block has index 1)
	if blockData["index"] != float64(2) {
		t.Errorf("Expected block index 2, got %v", blockData["index"])
	}
}

func TestGetChain(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%d/chain", serverPort)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("Failed to create request to /chain: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request to /chain: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %v, got %v", http.StatusOK, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}
}

func randomServerPort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(fmt.Sprintf("Failed to get random port: %v", err))
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port
}

func startServer(ch chan<- bool) {
	ctx := context.Background()
	go api.Start(ctx, &configuration.InstanceConfig)

	// Retry logic: make request every 2 seconds until success
	for {
		url := fmt.Sprintf("http://localhost:%d/health", serverPort)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			fmt.Printf("Failed to create request to /health: %v\n", err)
			ch <- false
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Failed to make request to /health: %v\n", err)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ch <- true
				return
			}
		}

		time.Sleep(2 * time.Second)
	}
}

func TestRegisterNodes(t *testing.T) {
	nodes := []string{
		"http://localhost:5001",
		"http://localhost:5002",
	}

	payload := map[string]interface{}{
		"nodes": nodes,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Send the request to register the nodes
	url := fmt.Sprintf("http://localhost:%d/nodes/register", serverPort)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("Failed to send request to /nodes/register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if result["message"] != "New nodes have been added" {
		t.Errorf("Expected message 'New nodes have been added', got %v", result["message"])
	}

	// Check the 'total_nodes' field and verify it is a map
	if totalNodes, ok := result["total_nodes"].(map[string]interface{}); ok {
		// Verify that the nodes from the payload are present in the total_nodes map
		for _, node := range nodes {
			if _, exists := totalNodes[node]; !exists {
				t.Errorf("Expected node %s to be registered, but it wasn't found", node)
			}
		}
	} else {
		t.Errorf("Expected 'total_nodes' to be a map[string]bool, got %v", result["total_nodes"])
	}
}
