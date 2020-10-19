package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"time"

	hawk "github.com/tent/hawk-go"

	"github.com/spf13/cobra"
)

var auth hawk.Auth

func main() {
	rootCmd := &cobra.Command{
		Use:   "hawk",
		Short: "Hawk HTTP Authentication helper",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&auth.RequestURI, "uri", "u", "", "request uri (required)")
	rootCmd.PersistentFlags().StringVarP(&auth.Method, "method", "m", "GET", "request method")
	rootCmd.PersistentFlags().StringVarP(&auth.Nonce, "nonce", "n", "", "Hawk Auth Nonce")
	rootCmd.PersistentFlags().StringVarP(&auth.Credentials.ID, "id", "i", "", "Hawk Auth ID")
	rootCmd.PersistentFlags().StringVarP(&auth.Credentials.Key, "key", "k", "", "Hawk Auth Key")
	rootCmd.PersistentFlags().StringVarP(&auth.Credentials.App, "app", "a", "", "Hawk Auth App")

	headerCmd := &cobra.Command{
		Use:   "header",
		Short: "Hawk HTTP Authentication helper",
		Run: func(cmd *cobra.Command, args []string) {
			if auth.Nonce == "" {
				setNonce()
			}
			auth.Credentials.Hash = sha256.New
			auth.Timestamp = hawk.Now().Add(10 * time.Second)
			auth.Host = "123123"
			auth.Port = "443"
			fmt.Println(auth.RequestHeader())
		},
	}

	rootCmd.AddCommand(headerCmd)
	rootCmd.Execute()
}

func setNonce() {
	b := make([]byte, 8)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic(err)
	}
	auth.Nonce = base64.StdEncoding.EncodeToString(b)[:8]
}
