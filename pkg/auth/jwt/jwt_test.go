package jwt_test

import (
	"context"
	"encoding/json"
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
	data, err := json.MarshalIndent(token, "", "  ")
	if !assert.NoError(err) {
		t.FailNow()
	}
	assert.NotNil(data)
	t.Log(string(data))
}

func Token() string {
	return `eyJhbGciOiJSUzI1NiIsImtpZCI6IjIzZjdhMzU4Mzc5NmY5NzEyOWU1NDE4ZjliMjEzNmZjYzBhOTY0NjIiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiIxMjE3NjA4MDg2ODgtaGJpaWJuaWgxdHJ0MnZyb2tocnRhMTdqZ2V1YWdwNGsuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMTYyODA2NzIxNTAzOTc5NTUwOTMiLCJoZCI6Im11dGFibGVsb2dpYy5jb20iLCJlbWFpbCI6ImRqdEBtdXRhYmxlbG9naWMuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsIm5iZiI6MTc0NTYxNjg5NywibmFtZSI6IkRhdmlkIFRob3JwZSIsInBpY3R1cmUiOiJodHRwczovL2xoMy5nb29nbGV1c2VyY29udGVudC5jb20vYS9BQ2c4b2NKcDdVcVh5OEluWGZOM251b3BLU3JFalFOaHVjcHV4VzBncDlTeDg5NW03aDk0WkJFSz1zOTYtYyIsImdpdmVuX25hbWUiOiJEYXZpZCIsImZhbWlseV9uYW1lIjoiVGhvcnBlIiwiaWF0IjoxNzQ1NjE3MTk3LCJleHAiOjE3NDU2MjA3OTcsImp0aSI6IjQ4ZWZhNDNmMGQyMjE0MmI4MTE4M2IyMmY3OTMzNGZlNTE5ZWYyNzgifQ.KVj1HSNvUfsNAFxX_4_eO-KM4u6hOFjp-Y6O9D43PkFf0DZQ_Xe9bk0xAob4vdfE0fHfrtsdgarZKhPg9tBmE9T93lUJtm7pcc0COLDopFCjVyduoc3cfNkVU2UYtdPAQup3ySSvGGRlh4o6AmiTtVG6cLLRwwHq-JLhdY7E74Wh_IjM7TYR4kc00XL6_bCAVyhvOEdm6tr5Ct4kuRu9Nym_nRk5jSwqtQNxySvOk2s-QJmTwvnI8iazG-vLZh0QQYJi0g-2uS4w8BT7uq7IjMkuwdcXT5S2ahL3jvsnb7SH8TMPLyir2wQpFOZld6mbrtUJfFLRReIHLjT3wuxq0g`
}
