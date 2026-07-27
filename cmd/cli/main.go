// Package main implements the LiquidLink CLI utility.
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Commands:\n")
		fmt.Fprintf(os.Stderr, "    wallet-gen      Generate a new Ed25519 wallet keypair and address\n")
		fmt.Fprintf(os.Stderr, "    balance         Query balance of a specific address or local wallet\n")
		fmt.Fprintf(os.Stderr, "    transfer        Create, sign and transmit a transaction\n")
		fmt.Fprintf(os.Stderr, "    channel-open    Open a new payment channel with a peer\n")
		fmt.Fprintf(os.Stderr, "    bridge-request  Request a crypto-to-fiat bridge settlement\n")
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "wallet-gen":
		fmt.Println("Generating wallet... (stub)")
		// TODO: Call wallet.Generator
	case "balance":
		fmt.Println("Querying balance... (stub)")
		// TODO: Hit Node API / Ledger status
	case "transfer":
		fmt.Println("Initiating transfer... (stub)")
		// TODO: Perform signing and POST to Node API
	case "channel-open":
		fmt.Println("Opening payment channel... (stub)")
		// TODO: Call pcn.Manager via REST API
	case "bridge-request":
		fmt.Println("Creating bridge request... (stub)")
		// TODO: Call receipt.Bridge via REST API
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}
