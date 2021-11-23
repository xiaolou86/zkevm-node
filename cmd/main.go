package main

import (
	"os"
	"os/signal"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hermeznetwork/hermez-core/aggregator"
	"github.com/hermeznetwork/hermez-core/config"
	"github.com/hermeznetwork/hermez-core/etherman"
	"github.com/hermeznetwork/hermez-core/jsonrpc"
	"github.com/hermeznetwork/hermez-core/log"
	"github.com/hermeznetwork/hermez-core/mocks"
	"github.com/hermeznetwork/hermez-core/sequencer"
	"github.com/hermeznetwork/hermez-core/synchronizer"
)

func main() {
	c := config.Load()
	setupLog(c.Log)
	go runJSONRpcServer(c.RPC)
	go runSequencer(c.Sequencer)
	go runAggregator(c.Aggregator)
	waitSignal()
}

func setupLog(c log.Config) {
	log.Init(c)
}

func runJSONRpcServer(c jsonrpc.Config) {
	p := mocks.NewPool()
	s := mocks.NewState()

	jsonrpc.NewServer(c, p, s).Start()
}

func runSequencer(c sequencer.Config) {
	p := mocks.NewPool()
	s := mocks.NewState()
	e, err := etherman.NewEtherman(c.Etherman)
	if err != nil {
		log.Fatal(err)
	}
	sy, err := synchronizer.NewSynchronizer(e, s)
	if err != nil {
		log.Fatal(err)
	}
	sequencer.NewSequencer(c, p, s, e, sy)
}

func runAggregator(c aggregator.Config) {
	s := mocks.NewState()
	bp := s.NewBatchProcessor(common.Hash{}, false)
	e, err := etherman.NewEtherman(c.Etherman)
	if err != nil {
		log.Fatal(err)
	}
	sy, err := synchronizer.NewSynchronizer(e, s)
	if err != nil {
		log.Fatal(err)
	}
	pc := aggregator.NewProverClient()
	aggregator.NewAggregator(c, s, bp, e, sy, pc)
}

func waitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			log.Info("terminating application gracefully...")
			os.Exit(0)
		}
	}
}