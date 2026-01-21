package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"freestealer/database"
	"freestealer/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB creates a PostgreSQL test database
func setupTestDB(t *testing.T) {
	// Use PostgreSQL for testing
	host := os.Getenv("TEST_DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("TEST_DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("TEST_DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("TEST_DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	dbname := os.Getenv("TEST_DB_NAME")
	if dbname == "" {
		dbname = "freestealer_test"
	}

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable"

	var err error
	database.DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping test - PostgreSQL not available: %v", err)
		return
	}

	// Clean and migrate
	database.DB.Exec("DROP SCHEMA IF EXISTS public CASCADE")
	database.DB.Exec("CREATE SCHEMA public")

	// Auto-migrate the schema
	err = database.DB.AutoMigrate(&models.User{}, &models.Tier{}, &models.Vote{}, &models.Comment{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
}

// setupTestAuth initializes auth for testing
func setupTestAuth() {
	os.Setenv("SESSION_SECRET", "test-secret")
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GITHUB_CALLBACK_URL", "http://localhost:5050/auth/github/callback")

	store = sessions.NewCookieStore([]byte("test-secret"))
	SetJWTSecret("test-jwt-secret")
}

func TestInitAuth(t *testing.T) {
	// Set required environment variables
	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")

	// Test with default callback URL
	InitAuth()
	assert.NotNil(t, store, "Store should be initialized")

	// Test with custom callback URL
	os.Setenv("GITHUB_CALLBACK_URL", "http://example.com/callback")
	InitAuth()
	assert.NotNil(t, store, "Store should be initialized with custom callback")

	// Clean up
	os.Unsetenv("GITHUB_CLIENT_ID")
	os.Unsetenv("GITHUB_CLIENT_SECRET")
	os.Unsetenv("GITHUB_CALLBACK_URL")
}

func TestBeginAuthHandler(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/auth/github", nil)
	w := httptest.NewRecorder()

	BeginAuthHandler(w, req)

	// Should add provider query param
	assert.Contains(t, req.URL.RawQuery, "provider=github")
}

func TestLogoutHandler(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/auth/logout", nil)
	w := httptest.NewRecorder()

	// Create a session with user data
	session, _ := store.Get(req, "auth-session")
	session.Values["user_id"] = uint(1)
	session.Values["github_id"] = "123456"
	session.Save(req, w)

	// Create new request with the session cookie
	req = httptest.NewRequest("GET", "/auth/logout", nil)
	if len(w.Result().Cookies()) > 0 {
		req.Header.Set("Cookie", w.Result().Cookies()[0].String())
	}
	w = httptest.NewRecorder()

	LogoutHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Logged out successfully", response["message"])
}

func TestGetCurrentUser_NotAuthenticated(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()

	GetCurrentUser(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUser_Authenticated(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "testuser",
		Email:       "test@example.com",
		GitHubID:    "123456",
		GitHubLogin: "testuser",
		AvatarURL:   "http://example.com/avatar.jpg",
	}
	database.DB.Create(&user)

	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()

	// Create a session with user data
	session, _ := store.Get(req, "auth-session")
	session.Values["user_id"] = user.ID
	session.Save(req, w)

	// Create new request with the session cookie
	req = httptest.NewRequest("GET", "/auth/me", nil)
	if len(w.Result().Cookies()) > 0 {
		req.Header.Set("Cookie", w.Result().Cookies()[0].String())
	}
	w = httptest.NewRecorder()

	GetCurrentUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()

	// Create a session with non-existent user ID
	session, _ := store.Get(req, "auth-session")
	session.Values["user_id"] = uint(999)
	session.Save(req, w)

	// Create new request with the session cookie
	req = httptest.NewRequest("GET", "/auth/me", nil)
	if len(w.Result().Cookies()) > 0 {
		req.Header.Set("Cookie", w.Result().Cookies()[0].String())
	}
	w = httptest.NewRecorder()

	GetCurrentUser(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRequireAuth_NotAuthenticated(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireAuth_Authenticated(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	// Create a session with user data
	session, _ := store.Get(req, "auth-session")
	session.Values["user_id"] = uint(1)
	session.Save(req, w)

	// Create new request with the session cookie
	req = httptest.NewRequest("GET", "/protected", nil)
	if len(w.Result().Cookies()) > 0 {
		req.Header.Set("Cookie", w.Result().Cookies()[0].String())
	}
	w = httptest.NewRecorder()

	handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Protected content", w.Body.String())
	assert.Equal(t, "1", req.Header.Get("X-User-ID"))
}

func TestRequireAuth_InvalidUserID(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	// Create a session with invalid user ID (0)
	session, _ := store.Get(req, "auth-session")
	session.Values["user_id"] = uint(0)
	session.Save(req, w)

	// Create new request with the session cookie
	req = httptest.NewRequest("GET", "/protected", nil)
	if len(w.Result().Cookies()) > 0 {
		req.Header.Set("Cookie", w.Result().Cookies()[0].String())
	}
	w = httptest.NewRecorder()

	handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// JWT Tests

func TestGenerateTokens(t *testing.T) {
	setupTestAuth()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.ID = 1

	tokens, err := GenerateTokens(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, "Bearer", tokens.TokenType)
	assert.Equal(t, int64(86400), tokens.ExpiresIn) // 24 hours
}

func TestValidateToken_Valid(t *testing.T) {
	setupTestAuth()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.ID = 1

	tokens, _ := GenerateTokens(user)

	claims, err := ValidateToken(tokens.AccessToken)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "test@example.com", claims.Email)
}

func TestValidateToken_Invalid(t *testing.T) {
	setupTestAuth()

	_, err := ValidateToken("invalid-token")

	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	setupTestAuth()

	// Create an expired token
	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		Email:    "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-jwt-secret"))

	_, err := ValidateToken(tokenString)

	assert.Error(t, err)
}

func TestExtractTokenFromHeader_Valid(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	token, err := ExtractTokenFromHeader(req)

	assert.NoError(t, err)
	assert.Equal(t, "valid-token", token)
}

func TestExtractTokenFromHeader_Missing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	_, err := ExtractTokenFromHeader(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authorization header required")
}

func TestExtractTokenFromHeader_InvalidFormat(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	_, err := ExtractTokenFromHeader(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid authorization header format")
}

func TestExtractTokenFromHeader_WrongScheme(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Basic sometoken")

	_, err := ExtractTokenFromHeader(req)

	assert.Error(t, err)
}

func TestRequireJWTAuth_Valid(t *testing.T) {
	setupTestAuth()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.ID = 1

	tokens, _ := GenerateTokens(user)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()

	handler := RequireJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Protected content", w.Body.String())
	assert.Equal(t, "1", req.Header.Get("X-User-ID"))
}

func TestRequireJWTAuth_NoToken(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	handler := RequireJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequireJWTAuth_InvalidToken(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler := RequireJWTAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetCurrentUser_WithJWT(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "jwtuser",
		Email:       "jwt@example.com",
		GitHubID:    "789",
		GitHubLogin: "jwtuser",
	}
	database.DB.Create(&user)

	// Generate JWT token
	tokens, _ := GenerateTokens(&user)

	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()

	GetCurrentUser(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.User
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "jwtuser", response.Username)
	assert.Equal(t, "jwt@example.com", response.Email)
}

func TestRequireAuth_WithJWT(t *testing.T) {
	setupTestAuth()

	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	user.ID = 1

	tokens, _ := GenerateTokens(user)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w := httptest.NewRecorder()

	handler := RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected content"))
	})

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", req.Header.Get("X-User-ID"))
}

func TestRefreshTokenHandler_Valid(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "refreshuser",
		Email:       "refresh@example.com",
		GitHubID:    "456",
		GitHubLogin: "refreshuser",
	}
	database.DB.Create(&user)

	// Generate initial tokens
	tokens, _ := GenerateTokens(&user)

	// Request new tokens with refresh token
	reqBody, _ := json.Marshal(RefreshTokenRequest{RefreshToken: tokens.RefreshToken})
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RefreshTokenHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TokenResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
}

func TestRefreshTokenHandler_InvalidToken(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	reqBody, _ := json.Marshal(RefreshTokenRequest{RefreshToken: "invalid-token"})
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RefreshTokenHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRefreshTokenHandler_InvalidBody(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RefreshTokenHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRefreshTokenHandler_UserNotFound(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a token for a non-existent user
	user := &models.User{
		Username: "ghost",
		Email:    "ghost@example.com",
	}
	user.ID = 999

	tokens, _ := GenerateTokens(user)

	reqBody, _ := json.Marshal(RefreshTokenRequest{RefreshToken: tokens.RefreshToken})
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RefreshTokenHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Login Tests

func TestLoginHandler_WithEmail(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "loginuser",
		Email:       "login@example.com",
		GitHubID:    "login123",
		GitHubLogin: "loginuser",
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(LoginRequest{Email: "login@example.com"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
	assert.NotNil(t, response["tokens"])
	assert.NotNil(t, response["user"])
}

func TestLoginHandler_WithUsername(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "usernamelogin",
		Email:       "username@example.com",
		GitHubID:    "username123",
		GitHubLogin: "usernamelogin",
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(LoginRequest{Username: "usernamelogin"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
}

func TestLoginHandler_WithGitHubID(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create a test user
	user := models.User{
		Username:    "githublogin",
		Email:       "github@example.com",
		GitHubID:    "gh789",
		GitHubLogin: "githublogin",
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(LoginRequest{GitHubID: "gh789"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
}

func TestLoginHandler_UserNotFound(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	reqBody, _ := json.Marshal(LoginRequest{Email: "nonexistent@example.com"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_NoCredentials(t *testing.T) {
	setupTestAuth()

	reqBody, _ := json.Marshal(LoginRequest{})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestLoginHandler_InvalidBody(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Password Hashing Tests

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
}

func TestCheckPasswordHash_Valid(t *testing.T) {
	password := "testpassword123"
	hash, _ := HashPassword(password)

	result := CheckPasswordHash(password, hash)

	assert.True(t, result)
}

func TestCheckPasswordHash_Invalid(t *testing.T) {
	password := "testpassword123"
	hash, _ := HashPassword(password)

	result := CheckPasswordHash("wrongpassword", hash)

	assert.False(t, result)
}

// Password Login Tests

func TestLoginHandler_WithPassword(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Hash password
	hashedPassword, _ := HashPassword("secret123")

	// Create a test user with password
	user := models.User{
		Username: "passworduser",
		Email:    "password@example.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(LoginRequest{Email: "password@example.com", Password: "secret123"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Login successful", response["message"])
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Hash password
	hashedPassword, _ := HashPassword("secret123")

	// Create a test user with password
	user := models.User{
		Username: "wrongpassuser",
		Email:    "wrongpass@example.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(LoginRequest{Email: "wrongpass@example.com", Password: "wrongpassword"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLoginHandler_MissingPassword(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Hash password
	hashedPassword, _ := HashPassword("secret123")

	// Create a test user with password
	user := models.User{
		Username: "needspassuser",
		Email:    "needspass@example.com",
		Password: hashedPassword,
	}
	database.DB.Create(&user)

	// Try to login without password
	reqBody, _ := json.Marshal(LoginRequest{Email: "needspass@example.com"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Registration Tests

func TestRegisterHandler_Success(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	reqBody, _ := json.Marshal(RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Registration successful", response["message"])
	assert.NotNil(t, response["tokens"])
	assert.NotNil(t, response["user"])
}

func TestRegisterHandler_DuplicateEmail(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create existing user
	user := models.User{
		Username: "existinguser",
		Email:    "existing@example.com",
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(RegisterRequest{
		Username: "newuser",
		Email:    "existing@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegisterHandler_DuplicateUsername(t *testing.T) {
	setupTestDB(t)
	setupTestAuth()

	// Create existing user
	user := models.User{
		Username: "takenname",
		Email:    "taken@example.com",
	}
	database.DB.Create(&user)

	reqBody, _ := json.Marshal(RegisterRequest{
		Username: "takenname",
		Email:    "new@example.com",
		Password: "password123",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	setupTestAuth()

	reqBody, _ := json.Marshal(RegisterRequest{
		Username: "user",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_ShortPassword(t *testing.T) {
	setupTestAuth()

	reqBody, _ := json.Marshal(RegisterRequest{
		Username: "user",
		Email:    "user@example.com",
		Password: "123",
	})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRegisterHandler_InvalidBody(t *testing.T) {
	setupTestAuth()

	req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
