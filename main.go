package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// CustomServer represents the server configuration
// 字段顺序与Rust版本保持一致，api和relay字段可选，没有输入则不输出
type CustomServer struct {
    Host  string `json:"host"`
    Key   string `json:"key"`
    API   string `json:"api,omitempty"`
    Relay string `json:"relay,omitempty"`
}

// reverseString reverses a string
func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}

// genName encodes a CustomServer to a reversed Base64 string
// This implementation matches the Rust version: gen_name function
func genName(lic CustomServer, privateKey ed25519.PrivateKey) (string, error) {
    // Serialize to JSON
    jsonData, err := json.Marshal(lic)
    if err != nil {
        return "", fmt.Errorf("failed to marshal JSON: %v", err)
    }
    fmt.Printf("Debug: JSON: %s\n", string(jsonData))
    
    // If private key is provided, sign the data
    var dataToEncode []byte
    if privateKey != nil {
        signedData := ed25519.Sign(privateKey, jsonData)
        dataToEncode = signedData
        fmt.Printf("Debug: Signed data (Base64): %s\n", base64.StdEncoding.EncodeToString(signedData))
    } else {
        // For compatibility with Rust version, use the JSON directly when not signing
        dataToEncode = jsonData
        fmt.Printf("Debug: Data to encode: %s\n", string(dataToEncode))
    }
    
    // Encode to Base64 (URL-safe, no padding) - matches Rust's URL_SAFE_NO_PAD
    base64Encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(dataToEncode)
    fmt.Printf("Debug: Base64: %s\n", base64Encoded)
    
    // Reverse the string - matches Rust's chars().rev().collect()
    reversed := reverseString(base64Encoded)
    fmt.Printf("Debug: Reversed: %s\n", reversed)
    
    return reversed, nil
}

// getCustomServerFromConfigString decodes a reversed Base64 string to a CustomServer
// This implementation matches the Rust version's decoding process
func getCustomServerFromConfigString(s string, publicKey ed25519.PublicKey) (CustomServer, error) {
    var lic CustomServer

    // Reverse the string - matches Rust's chars().rev().collect()
    reversed := reverseString(s)
    fmt.Printf("Debug: Reversed string: %s\n", reversed)

    // Base64 decode - matches Rust's URL_SAFE_NO_PAD.decode
    decoded, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(reversed)
    if err != nil {
        return lic, fmt.Errorf("failed to decode Base64: %v", err)
    }
    fmt.Printf("Debug: Decoded string: %s\n", string(decoded))

    // Try direct JSON parsing (unsigned)
    if err := json.Unmarshal(decoded, &lic); err == nil {
        return lic, nil
    }

    // Try signature verification (signed)
    if publicKey != nil {
        if len(decoded) < ed25519.SignatureSize {
            return lic, fmt.Errorf("decoded data too short for signature")
        }
        signature := decoded[:ed25519.SignatureSize]
        message := decoded[ed25519.SignatureSize:]
        if ed25519.Verify(publicKey, message, signature) {
            if err := json.Unmarshal(message, &lic); err != nil {
                return lic, fmt.Errorf("failed to parse verified JSON: %v", err)
            }
            fmt.Printf("Debug: Verified JSON: %s\n", string(message))
            return lic, nil
        }
        return lic, fmt.Errorf("signature verification failed")
    }

    return lic, fmt.Errorf("failed to parse JSON and no public key for verification")
}

// promptForInput collects interactive input
func promptForInput() (CustomServer, error) {
    var lic CustomServer

    fmt.Print("Enter key: ")
    key, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return lic, fmt.Errorf("failed to read key: %v", err)
    }
    lic.Key = strings.TrimSpace(string(key))
    if lic.Key == "" {
        return lic, fmt.Errorf("key is required")
    }
    fmt.Println()

    fmt.Print("Enter host: ")
    host, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return lic, fmt.Errorf("failed to read host: %v", err)
    }
    lic.Host = strings.TrimSpace(string(host))
    if lic.Host == "" {
        return lic, fmt.Errorf("host is required")
    }
    fmt.Println()

    fmt.Print("Enter api (optional, press Enter to skip): ")
    api, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return lic, fmt.Errorf("failed to read api: %v", err)
    }
    lic.API = strings.TrimSpace(string(api))
    fmt.Println()

    fmt.Print("Enter relay (optional, press Enter to skip): ")
    relay, err := term.ReadPassword(int(os.Stdin.Fd()))
    if err != nil {
        return lic, fmt.Errorf("failed to read relay: %v", err)
    }
    lic.Relay = strings.TrimSpace(string(relay))
    fmt.Println()

    return lic, nil
}

func main() {
    // Public key from Rust code
    publicKeyBytes := []byte{88, 168, 68, 104, 60, 5, 163, 198, 165, 38, 12, 85, 114, 203, 96, 163, 70, 48, 0, 131, 57, 12, 46, 129, 83, 17, 84, 193, 119, 197, 130, 103}
    if len(publicKeyBytes) != ed25519.PublicKeySize {
        fmt.Fprintf(os.Stderr, "Invalid public key length\n")
        os.Exit(1)
    }
    publicKey := ed25519.PublicKey(publicKeyBytes)

    // Private key (placeholder, set via environment variable or replace with actual key)
    var privateKey ed25519.PrivateKey
    privateKeyB64 := os.Getenv("ED25519_PRIVATE_KEY")
    if privateKeyB64 != "" {
        privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Invalid private key: %v\n", err)
            os.Exit(1)
        }
        if len(privateKeyBytes) != ed25519.PrivateKeySize {
            fmt.Fprintf(os.Stderr, "Invalid private key length\n")
            os.Exit(1)
        }
        privateKey = ed25519.PrivateKey(privateKeyBytes)
    }
    // TODO: Replace with actual private key if known
    // For example: privateKeyBytes := []byte{...}
    // privateKey = ed25519.PrivateKey(privateKeyBytes)

    args := os.Args[1:]
    if len(args) >= 2 {
        // Encode mode
        lic := CustomServer{
            Key:   args[0],
            Host:  args[1],
            API:   "",
            Relay: "",
        }
        if len(args) > 2 {
            lic.API = args[2]
        }
        if len(args) > 3 {
            lic.Relay = args[3]
        }

        result, err := genName(lic, privateKey)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("rustdesk-licensed-%s.exe\n", result)
    } else if len(args) == 1 {
        // Decode mode
        // Remove possible prefix/suffix
        input := strings.TrimPrefix(strings.TrimSuffix(args[0], ".exe"), "rustdesk-licensed-")
        input = strings.ReplaceAll(input, "-licensed-", "")

        lic, err := getCustomServerFromConfigString(input, publicKey)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("Decoded CustomServer:\n")
        fmt.Printf("  key: %s\n", lic.Key)
        fmt.Printf("  host: %s\n", lic.Host)
        fmt.Printf("  api: %s\n", lic.API)
        fmt.Printf("  relay: %s\n", lic.Relay)
    } else {
        // Interactive mode
        fmt.Println("No command-line arguments provided. Entering interactive mode.")
        lic, err := promptForInput()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }

        result, err := genName(lic, privateKey)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
        fmt.Printf("rustdesk-custom_serverd-%s.exe\n", result)
    }
}