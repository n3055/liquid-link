// Package main implements the LiquidLink CLI utility.
package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/n3055/liquid-link/internal/wallet"
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
		genCmd := flag.NewFlagSet("wallet-gen", flag.ExitOnError)
		aliasFlag := genCmd.String("alias", "", "Alias to save the generated wallet under (optional)")
		dirFlag := genCmd.String("dir", "", "Directory to save the wallet (default: ~/.liquid-link/wallets)")
		
		if err := genCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			os.Exit(1)
		}

		w, err := wallet.GenerateWallet()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating wallet: %v\n", err)
			os.Exit(1)
		}

		pubKeyHex := hex.EncodeToString(w.ExportPublicKey())

		fmt.Println("New Ed25519 Wallet Generated successfully!")
		fmt.Printf("Address:    %s\n", w.Address())
		fmt.Printf("Public Key: %s\n", pubKeyHex)

		if *aliasFlag != "" {
			dir, err := getWalletDir(*dirFlag)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error getting wallet directory: %v\n", err)
				os.Exit(1)
			}

			repo, err := wallet.NewFileRepository(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating repository: %v\n", err)
				os.Exit(1)
			}

			ctx := context.Background()
			if err := repo.Save(ctx, *aliasFlag, w); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving wallet: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Wallet saved with alias '%s' at: %s\n", *aliasFlag, filepath.Join(dir, *aliasFlag+".json"))
		} else {
			// If not saving, display private key so they can keep it.
			// But note that it's sensitive, so we warn the user.
			// (Since the user wants save/load and sign/verify functions attached, displaying private key here is helpful).
			// We can get the private key by casting or we can just print a message.
			// Wait, can we extract the private key to print it?
			// Let's implement an export private key method or cast to localWallet.
			// Wait, let's see how we can display the private key. Let's check:
			if lw, ok := w.(interface{ ExportPrivateKey() []byte }); ok {
				fmt.Printf("Private Key:%s\n", hex.EncodeToString(lw.ExportPrivateKey()))
			}
			fmt.Println("Warning: Private key was not saved. Write it down or run with -alias to save to disk.")
		}

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

// ExportPrivateKey is implemented by localWallet in wallet_impl.go to allow CLI to show private keys.
// We can define that in wallet_impl.go or type assert.
func getWalletDir(customDir string) (string, error) {
	if customDir != "" {
		return customDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".liquid-link", "wallets"), nil
}
