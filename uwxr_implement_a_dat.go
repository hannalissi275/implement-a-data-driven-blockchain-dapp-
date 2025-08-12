Go
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Config for dApp monitor
type Config struct {
	EthNodeURL  string `json:"ethNodeUrl"`
	ContractABI string `json:"contractAbi"`
	ContractAddr string `json:"contractAddr"`
	PollInterval int    `json:"pollInterval"` // in seconds
}

type Transaction struct {
	TxHash    common.Hash   `json:"txHash"`
	BlockNum  uint64       `json:"blockNum"`
	Timestamp int64        `json:"timestamp"`
	From      common.Address `json:"from"`
	To        common.Address `json:"to"`
	Value     *types.BigInt  `json:"value"`
}

type Monitor struct {
	client *ethclient.Client
	contract *bind.BoundContract
	pollInterval int
	transactions chan Transaction
}

func NewMonitor(cfg Config) (*Monitor, error) {
	client, err := ethclient.Dial(cfg.EthNodeURL)
	if err != nil {
		return nil, err
	}

	parsedABI, err := bind.ParseJSON(strings.NewReader(cfg.ContractABI))
	if err != nil {
		return nil, err
	}

	contractAddr := common.HexToAddress(cfg.ContractAddr)
	instance, err := bind.NewBoundContract(contractAddr, parsedABI, client, client, client)
	if err != nil {
		return nil, err
	}

	return &Monitor{
		client: client,
		contract: instance,
		pollInterval: cfg.PollInterval,
		transactions: make(chan Transaction, 10),
	}, nil
}

func (m *Monitor) Start() {
	go func() {
		for {
			header, err := m.client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Println(err)
				continue
			}

			for _, tx := range header.Transactions() {
				txHash := tx.Hash()
				txReceipt, _, err := m.client.TransactionReceipt(context.Background(), txHash)
				if err != nil {
					log.Println(err)
					continue
				}

				from, err := types.Sender(m.client, tx)
				if err != nil {
					log.Println(err)
					continue
				}

				to, err := tx.To()
				if err != nil {
					log.Println(err)
					continue
				}

				m.transactions <- Transaction{
					TxHash:    txHash,
					BlockNum:  header.Number.Int64(),
					Timestamp: int64(header.Time),
					From:      from,
					To:        to,
					Value:     tx.Value(),
				}
			}

			time.Sleep(time.Duration(m.pollInterval) * time.Second)
		}
	}()
}

func (m *Monitor) Transactions() <-chan Transaction {
	return m.transactions
}

func main() {
(cfg := Config{
		EthNodeURL:  "https://mainnet.infura.io/v3/YOUR_PROJECT_ID",
		ContractABI: `abi.json content`,
		ContractAddr: "0x Contract Address",
		PollInterval: 10,
	})

	monitor, err := NewMonitor(cfg)
	if err != nil {
		log.Fatal(err)
	}

	monitor.Start()

	for tx := range monitor.Transactions() {
		fmt.Printf("Received transaction: %+v\n", tx)
	}
}