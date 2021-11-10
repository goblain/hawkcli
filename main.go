package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	hawk "github.com/tent/hawk-go"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	viper.SetEnvPrefix("hawk")
	viper.BindEnv("id")
	viper.BindEnv("key")
	viper.BindEnv("app")

	headerCmd := &cobra.Command{
		Use:   "header",
		Short: "Hawk HTTP Authentication helper",
		Run: func(cmd *cobra.Command, args []string) {
			if auth.Nonce == "" {
				setNonce()
			}
			url, _ := url.Parse(auth.RequestURI)
			hostParsed := strings.Split(url.Host, ":")
			auth.Host = hostParsed[0]
			query := ""
			if url.RawQuery != "" {
				query = "/" + url.RawQuery
			}
			auth.RequestURI = url.Path + query
			auth.Credentials.Hash = sha256.New
			auth.Timestamp = hawk.Now().Add(10 * time.Second)
			if len(hostParsed) > 1 {
				auth.Port = hostParsed[1]
			} else {
				if url.Scheme == "https" {
					auth.Port = "443"
				} else {
					auth.Port = "80"
				}
			}
			fmt.Println(auth.RequestHeader())
		},
	}
	headerCmd.PersistentFlags().StringVarP(&auth.RequestURI, "uri", "u", "", "request uri (required)")
	headerCmd.PersistentFlags().StringVarP(&auth.Method, "method", "m", "GET", "request method")
	headerCmd.PersistentFlags().StringVarP(&auth.Nonce, "nonce", "n", "", "Hawk Auth Nonce")
	headerCmd.PersistentFlags().StringVarP(&auth.Credentials.ID, "id", "i", "", "Hawk Auth ID")
	headerCmd.PersistentFlags().StringVarP(&auth.Credentials.Key, "key", "k", "", "Hawk Auth Key")
	headerCmd.PersistentFlags().StringVarP(&auth.Credentials.App, "app", "a", "", "Hawk Auth App")

	curlCmd := &cobra.Command{
		Use:                "curl",
		Short:              "Wrap curl command with hawk header",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			curlURL, err := extractUrl(os.Args[2:])
			if err != nil {
				panic(err)
			}
			if auth.Nonce == "" {
				setNonce()
			}
			url, err := url.Parse(curlURL)
			if err != nil {
				panic(err)
			}
			hostParsed := strings.Split(url.Host, ":")
			auth.Host = hostParsed[0]
			query := ""
			if url.RawQuery != "" {
				query = "/" + url.RawQuery
			}
			auth.RequestURI = url.Path + query
			auth.Credentials.Hash = sha256.New
			auth.Timestamp = hawk.Now().Add(10 * time.Second)

			auth.Method = extractMethod(os.Args[2:])
			auth.Credentials.ID = viper.Get("id").(string)
			auth.Credentials.Key = viper.Get("key").(string)
			auth.Credentials.App = viper.Get("app").(string)

			if len(hostParsed) > 1 {
				auth.Port = hostParsed[1]
			} else {
				if url.Scheme == "https" {
					auth.Port = "443"
				} else {
					auth.Port = "80"
				}
			}

			curlCmd := exec.Command("curl", "-H", "Authorization: "+auth.RequestHeader())
			for _, v := range os.Args[2:] {
				curlCmd.Args = append(curlCmd.Args, v)
			}
			curlCmd.Stdout = os.Stdout
			curlCmd.Stderr = os.Stderr
			if err := curlCmd.Run(); err != nil {
				panic(err)
			}
		},
	}
	rootCmd.AddCommand(headerCmd)
	rootCmd.AddCommand(curlCmd)
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

type CurlFlags struct {
	Flags []CurlFlag
}

type CurlFlag struct {
	Flag     string
	HasValue bool
}

func NewCurlFlags() (*CurlFlags, error) {
	cmd := exec.Command("curl", "--help")
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("Command '%s' failed with: %w", cmd.String(), err)
	}

	f := &CurlFlags{Flags: []CurlFlag{}}

	re := regexp.MustCompile(`(-\w)?,?\s*(--[\w\d\.]+)\s*(\[.*\]|\<.*\>)?`)

	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		match := re.FindStringSubmatch(scanner.Text())
		if len(match) > 0 {
			hasValue := false
			if match[3] != "" {
				hasValue = true
			}
			if match[1] != "" {
				f.Flags = append(f.Flags, CurlFlag{Flag: match[1], HasValue: hasValue})
			}
			if match[2] != "" {
				f.Flags = append(f.Flags, CurlFlag{Flag: match[2], HasValue: hasValue})
			}
		}
	}
	return f, nil
}

func (cf *CurlFlags) Get(flag string) *CurlFlag {
	for _, flagi := range cf.Flags {
		if flagi.Flag == flag {
			return &flagi
		}
	}
	return nil
}

func extractMethod(args []string) string {
	method := "GET"
	for k, v := range args {
		if v == "-X" || v == "--request" {
			method = args[k+1]
			break
		}
	}
	return method
}

func extractUrl(args []string) (string, error) {
	flags, err := NewCurlFlags()
	if err != nil {
		return "", err
	}
	var flag *CurlFlag
	var filtered []string
	for _, v := range args {
		if strings.HasPrefix(v, "-") {
			flag = flags.Get(v)
		} else {
			if flag != nil && flag.HasValue {
				flag = nil
				continue
			}
			filtered = append(filtered, v)
			flag = nil
		}
	}
	if len(filtered) == 1 {
		return filtered[0], nil
	}
	return "", fmt.Errorf("URL not identified in %v", filtered)
}
