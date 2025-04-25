package jwt_test

import (
	"context"
	"testing"

	// Packages
	"github.com/mutablelogic/go-server/pkg/auth/jwt"
	"github.com/stretchr/testify/assert"
)

func Test_JWT_001(t *testing.T) {
	assert := assert.New(t)
	jwt, err := jwt.New(context.Background())
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(jwt)
}

func Test_JWT_002(t *testing.T) {
	assert := assert.New(t)
	jwt, err := jwt.New(context.Background(), jwt.WithGoogle())
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(jwt)
	token, err := jwt.Decode(Token())
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(token)
	t.Log(token)
}

func Test_JWT_003(t *testing.T) {
	assert := assert.New(t)

	region := "eu-central-1"
	userPoolId := "eu-central-1_JYdTrAi2T"

	jwt, err := jwt.New(context.Background(), jwt.WithGoogle(), jwt.WithCognito(region, userPoolId))
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(jwt)
	token, err := jwt.Decode(Token())
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(token)
	t.Log(token)
}

func Token() string {
	return `eyJhbGciOiJSUzI1NiIsImtpZCI6IjIzZjdhMzU4Mzc5NmY5NzEyOWU1NDE4ZjliMjEzNmZjYzBhOTY0NjIiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMTYyODA2NzIxNTAzOTc5NTUwOTMiLCJoZCI6Im11dGFibGVsb2dpYy5jb20iLCJlbWFpbCI6ImRqdEBtdXRhYmxlbG9naWMuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5iZiI6MTc0NTYwMzMwMSwibmFtZSI6IkRhdmlkIFRob3JwZSIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS9BQ2c4b2NKcDdVcVh5OEluWGZOM251b3BLU3JFalFOaHVjcHV4VzBncDlTeDg5NW03aDk0WkJFSz1zOTYtYyIsImdpdmVuX25hbWUiOiJEYXZpZCIsImZhbWlseV9uYW1lIjoiVGhvcnBlIiwiaWF0IjoxNzQ1NjAzNjAxLCJleHAiOjE3NDU2MDcyMDEsImp0aSI6IjRjNGMwNjA2Njk1M2FkZDg5ZjY1MDhjNjYyYmFlNjM1ZWE3N2NjZDgifQ.Pfv5PtP2jR9XVQLH51KXWc0jI9P5UW92JouuXtv8XiKjrr9UOkEVn3v0G4JVKhM5ykRhmi6tz9jwUMK9YXcl45t4Ccbz5nX6VfRSABN9ZnnWIeV8LIUyX-rY3ZEZPOC0tB1wiOym8XRXmVpeXzcjbQO6sXw4mbSvq80qdHxPuiybMQ4bJRmqCVky_Z1NULenEUuhh0YdjxI3W-RFwOeNaewJf2JqRikr4dc-mKX4v4aU7tU1THZyHsshUOQ_LPBJmr_tIrhh6Qe22e6PL-KzKNePixQCWxJRR9FxxQwX9jG7Z-8sqoFZkPjofbHnv_NHSUim23R5zJWJVGSG2A8Irw`
}
