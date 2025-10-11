// Command jwt_quickstart prints a signed JWT with a roles claim for testing the Casbin PoC.
// It reads the HMAC secret from COSTSCOPE_JWT_SECRET and supports flags to set roles and issuer.
//
// Usage:
//
//	COSTSCOPE_JWT_SECRET=... go run ./examples/security/jwt_quickstart.go -roles admin,viewer -issuer costscope -ttl 60
//
// Then:
//
//	curl -i -H "Authorization: Bearer $(go run ./examples/security/jwt_quickstart.go -roles admin)" -X POST http://localhost:8080/api/v1/focus/convert
package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func mustSecret() []byte {
	s := os.Getenv("COSTSCOPE_JWT_SECRET")
	if len(s) == 0 {
		// Generate a throwaway secret for local experiments to avoid confusion
		// but print a warning so users know to align with the API server.
		buf := make([]byte, 48)
		_, _ = rand.Read(buf)
		fmt.Fprintln(os.Stderr, "WARN: COSTSCOPE_JWT_SECRET not set; generated a temporary secret for this token. The API must use the same secret to accept it.")
		return []byte(base64.StdEncoding.EncodeToString(buf))
	}
	return []byte(s)
}

func main() {
	rolesFlag := flag.String("roles", "admin", "Comma-separated role names (e.g. admin,viewer). 'role:' prefix is added automatically where missing.")
	issuerFlag := flag.String("issuer", "", "Optional JWT issuer (set to match the API's configured issuer, if any).")
	subjectFlag := flag.String("sub", "user@example", "Optional subject (user identifier).")
	ttlMinutes := flag.Int("ttl", 60, "Token TTL in minutes.")
	flag.Parse()

	roles := strings.Split(*rolesFlag, ",")
	normRoles := make([]string, 0, len(roles))
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		if strings.HasPrefix(r, "role:") {
			normRoles = append(normRoles, r)
		} else {
			normRoles = append(normRoles, "role:"+r)
		}
	}

	now := time.Now()
	expires := now.Add(time.Duration(*ttlMinutes) * time.Minute)

	claims := jwt.MapClaims{
		// roles as a simple array; the server extracts and normalizes these
		"roles": normRoles,
		"sub":   *subjectFlag,
		"iat":   now.Unix(),
		"exp":   expires.Unix(),
	}
	if *issuerFlag != "" {
		claims["iss"] = *issuerFlag
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(mustSecret())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to sign token:", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
