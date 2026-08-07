package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"splitnow/auth-service/internal/handlers/mocks"
	"splitnow/auth-service/internal/token"
	"splitnow/auth-service/internal/types"
)

// test setup
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

//===== Tests for POST "/users" =====//

// testing user registration REST endpoint (positive case)
func TestCreateUser(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		CreateUser("example@gmail.com", "Foo", mock.Anything).
		Return(types.User{
			ID:        1,
			Email:     "example@gmail.com",
			Username:  "Foo",
			CreatedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
		}, nil)

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	body := bytes.NewBufferString(`{"email":"example@gmail.com","username":"Foo","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var user types.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)

	// check if returned fields are correct
	assert.Equal(t, "example@gmail.com", user.Email)
	assert.Equal(t, "Foo", user.Username)
	assert.Equal(t, time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC), user.CreatedAt)

	// check that it does not return the password in any way
	assert.NotContains(t, w.Body.String(), "password")
}

// email is already registered
func TestCreateUserEmailTaken(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		CreateUser("example@gmail.com", "Foo", mock.Anything).
		Return(types.User{}, types.ErrEmailTaken)

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	body := bytes.NewBufferString(`{"email":"example@gmail.com","username":"Foo","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// assert that that response specified which field is already taken
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	assert.Equal(t, "email", respBody["field"])
}

// username is already taken
func TestCreateUserUsernameTaken(t *testing.T) {
	store := mocks.NewStore(t)
	store.EXPECT().
		CreateUser("example@gmail.com", "Foo", mock.Anything).
		Return(types.User{}, types.ErrUsernameTaken)

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	body := bytes.NewBufferString(`{"email":"example@gmail.com","username":"Foo","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	// assert that that response specified which field is already taken
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	assert.Equal(t, "username", respBody["field"])
}

func TestCreateUserStoreError(t *testing.T) {
	store := mocks.NewStore(t)
	// introduce error in db
	store.EXPECT().
		CreateUser("example@gmail.com", "Foo", mock.Anything).
		Return(types.User{}, errors.New("database unavailable"))

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	body := bytes.NewBufferString(`{"email":"example@gmail.com","username":"Foo","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/users", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// assert that it gives correct 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateUserInvalidBody(t *testing.T) {
	store := mocks.NewStore(t)

	router := gin.Default()
	router.POST("/users", CreateUser(store))

	cases := []struct {
		name string
		body string
	}{
		{"missing email", `{"username":"Foo","password":"1234"}`},
		{"missing username", `{"email":"example@gmail.com","password":"1234"}`},
		{"missing password", `{"email":"example@gmail.com","username":"Foo"}`},
		{"wrong type", `{"email":123,"username":"Foo","password":"1234"}`},
		{"malformed json", `{"email":`},
		{"empty body", ``},
		{"password too short", `{"email":"example@gmail.com","username":"Foo","password":"1234"}`},
		{"username too short", `{"email":"example@gmail.com","username":"ab","password":"12345678"}`},
		{"username too long", `{"email":"example@gmail.com","username":"aaaaaaaaaaaaaaaaaaaaa","password":"12345678"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, `{"error":"invalid body"}`, w.Body.String())
		})
	}
}

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	assert.NoError(t, err)
	assert.NotEqual(t, "secret", hash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret")))
}

func TestHashPasswordTooLong(t *testing.T) {
	_, err := HashPassword(strings.Repeat("a", 73))
	assert.ErrorIs(t, err, bcrypt.ErrPasswordTooLong)
}

//===== Tests for POST "/login" =====//

func newTestManager(t *testing.T) *token.Manager {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	dir := t.TempDir()
	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(privatePath,
		pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0600))

	publicDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(publicPath,
		pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600))

	mgr, err := token.New(privatePath, publicPath)
	assert.NoError(t, err)

	return mgr
}

func TestLoginValidCredentials(t *testing.T) {
	store := mocks.NewStore(t)
	passwordHash, err := HashPassword("12345678")
	assert.NoError(t, err)

	store.EXPECT().
		GetUserByEmail("johndoe@gmail.com").
		Return(types.User{
			ID:        1,
			Username:  "John Doe",
			Email:     "johndoe@gmail.com",
			Password:  passwordHash,
			CreatedAt: time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC),
		}, nil)

	tm := newTestManager(t)
	router := gin.Default()
	router.POST("/login", Login(store, tm))

	body := bytes.NewBufferString(`{"email":"johndoe@gmail.com","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// assert that its valid json containing a token field with some token
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	tokenString, ok := respBody["token"]
	assert.True(t, ok)

	claims, err := tm.Verify(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "John Doe", claims.Username)
}

func TestLoginInvalidPassword(t *testing.T) {
	store := mocks.NewStore(t)
	passwordHash, err := HashPassword("12345678")
	assert.NoError(t, err)

	store.EXPECT().
		GetUserByEmail("johndoe@gmail.com").
		Return(types.User{
			ID:        1,
			Username:  "John Doe",
			Email:     "johndoe@gmail.com",
			Password:  passwordHash,
			CreatedAt: time.Date(2026, 6, 23, 12, 0, 0, 0, time.UTC),
		}, nil)

	tm := newTestManager(t)
	router := gin.Default()
	router.POST("/login", Login(store, tm))

	// send wrong password
	body := bytes.NewBufferString(`{"email":"johndoe@gmail.com","password":"99999999"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// assert that the valid json does not contain a token at all
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	_, ok := respBody["token"]
	assert.False(t, ok)

	// assert that the message does not specify what is incorrect
	assert.JSONEq(t, `{"error":"invalid credentials"}`, w.Body.String())

}

// test that an incorrect email / user not found is handled properly
func TestLoginUserNotFound(t *testing.T) {
	store := mocks.NewStore(t)

	store.EXPECT().
		GetUserByEmail("john@gmail.com").
		Return(types.User{}, types.ErrUserNotFound)

	tm := newTestManager(t)
	router := gin.Default()
	router.POST("/login", Login(store, tm))

	// send wrong email
	body := bytes.NewBufferString(`{"email":"john@gmail.com","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// NOTE: it should always be the same as no password, never 404
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// assert that the valid json does not contain a token at all
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	_, ok := respBody["token"]
	assert.False(t, ok)

	// assert that the message does not specify what is incorrect
	assert.JSONEq(t, `{"error":"invalid credentials"}`, w.Body.String())
}

// tests for missing fields
func TestLoginInvalidBody(t *testing.T) {
	tm := newTestManager(t)
	router := gin.Default()
	router.POST("/login", Login(mocks.NewStore(t), tm))

	cases := []struct {
		name string
		body string
	}{
		{"missing email", `{"password":"12345678"}`},
		{"missing password", `{"email":"johndoe@gmail.com"}`},
		{"missing both", `{}`},
		{"empty email", `{"email":"","password":"12345678"}`},
		{"wrong type", `{"email":123,"password":"12345678"}`},
		{"malformed json", `{"email":`},
		{"empty body", ``},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(tc.body))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.JSONEq(t, `{"error":"email and password are required"}`, w.Body.String())
		})
	}
}

func TestLoginStoreError(t *testing.T) {
	store := mocks.NewStore(t)

	store.EXPECT().
		GetUserByEmail("johndoe@gmail.com").
		Return(types.User{}, errors.New("database unavailable"))

	tm := newTestManager(t)
	router := gin.Default()
	router.POST("/login", Login(store, tm))

	body := bytes.NewBufferString(`{"email":"johndoe@gmail.com","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// assert that the valid json does not contain a token at all
	var respBody map[string]string
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &respBody))
	_, ok := respBody["token"]
	assert.False(t, ok)
}

type failingTokenIssuer struct{}

func (failingTokenIssuer) Issue(userID int, username string) (string, error) {
	return "", errors.New("signing failed")
}

func TestLoginTokenIssueError(t *testing.T) {
	store := mocks.NewStore(t)
	passwordHash, err := HashPassword("12345678")
	assert.NoError(t, err)

	store.EXPECT().
		GetUserByEmail("johndoe@gmail.com").
		Return(types.User{
			ID: 1, Username: "John Doe", Email: "johndoe@gmail.com", Password: passwordHash,
		}, nil)

	router := gin.Default()
	router.POST("/login", Login(store, failingTokenIssuer{}))

	body := bytes.NewBufferString(`{"email":"johndoe@gmail.com","password":"12345678"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.JSONEq(t, `{"error":"failed to create token"}`, w.Body.String())
}
