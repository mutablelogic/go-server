package main

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	// Packages
	jwt "github.com/golang-jwt/jwt/v5"
	client "github.com/mutablelogic/go-client"
)

func Token() string {
	return `eyJhbGciOiJSUzI1NiIsImtpZCI6IjIzZjdhMzU4Mzc5NmY5NzEyOWU1NDE4ZjliMjEzNmZjYzBhOTY0NjIiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMTYyODA2NzIxNTAzOTc5NTUwOTMiLCJoZCI6Im11dGFibGVsb2dpYy5jb20iLCJlbWFpbCI6ImRqdEBtdXRhYmxlbG9naWMuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5iZiI6MTc0NTYwMzMwMSwibmFtZSI6IkRhdmlkIFRob3JwZSIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS9BQ2c4b2NKcDdVcVh5OEluWGZOM251b3BLU3JFalFOaHVjcHV4VzBncDlTeDg5NW03aDk0WkJFSz1zOTYtYyIsImdpdmVuX25hbWUiOiJEYXZpZCIsImZhbWlseV9uYW1lIjoiVGhvcnBlIiwiaWF0IjoxNzQ1NjAzNjAxLCJleHAiOjE3NDU2MDcyMDEsImp0aSI6IjRjNGMwNjA2Njk1M2FkZDg5ZjY1MDhjNjYyYmFlNjM1ZWE3N2NjZDgifQ.Pfv5PtP2jR9XVQLH51KXWc0jI9P5UW92JouuXtv8XiKjrr9UOkEVn3v0G4JVKhM5ykRhmi6tz9jwUMK9YXcl45t4Ccbz5nX6VfRSABN9ZnnWIeV8LIUyX-rY3ZEZPOC0tB1wiOym8XRXmVpeXzcjbQO6sXw4mbSvq80qdHxPuiybMQ4bJRmqCVky_Z1NULenEUuhh0YdjxI3W-RFwOeNaewJf2JqRikr4dc-mKX4v4aU7tU1THZyHsshUOQ_LPBJmr_tIrhh6Qe22e6PL-KzKNePixQCWxJRR9FxxQwX9jG7Z-8sqoFZkPjofbHnv_NHSUim23R5zJWJVGSG2A8Irw`
}

func DecodeJWT(value string, cert *x509.Certificate) (*jwt.Token, error) {
	return jwt.Parse(value, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return cert.PublicKey, nil
	}, jwt.WithValidMethods([]string{
		jwt.SigningMethodRS256.Alg(),
		jwt.SigningMethodRS384.Alg(),
		jwt.SigningMethodRS512.Alg(),
	}))
}

func GetCerts(ctx context.Context, url string) (map[string]*x509.Certificate, error) {
	// Fetch the certs from the given URL
	fetcher, err := client.New(client.OptEndpoint(url))
	if err != nil {
		return nil, err
	}
	var response map[string]string
	if err := fetcher.DoWithContext(context.Background(), client.MethodGet, &response); err != nil {
		return nil, err
	}

	// Decode the PEM blocks
	result := make(map[string]*x509.Certificate, len(response))
	for k, v := range response {
		block, rest := pem.Decode([]byte(v))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block: %s", v)
		}
		if len(rest) > 0 {
			return nil, fmt.Errorf("extra data found after PEM block: %s", rest)
		}
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %s", err)
		}
		result[k] = cert
	}

	// Return the parsed certificates
	return result, nil
}

func main() {
	certs, err := GetCerts(context.TODO(), "https://www.googleapis.com/oauth2/v1/certs")
	if err != nil {
		panic(err)
	}

	// Use Key ID to get the certificate
	cert := certs["23f7a3583796f97129e5418f9b2136fcc0a96462"]
	fmt.Println(cert.PublicKeyAlgorithm)

	if token, err := DecodeJWT(Token(), cert); err != nil {
		panic(err)
	} else {
		data, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(data))
	}
}
