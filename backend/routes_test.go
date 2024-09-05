package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

var (
	config     = getTestConfig()
	testServer = Server{
		Config: config,
	}
)
var testRouter = getTestRouter()

func getTestRouter() *chi.Mux {
	return testServer.getRouter()
}

func getTestConfig() Config {
	return Config{
		Title:      "title",
		Brand:      "brand",
		Support:    "support@example.com",
		SigningKey: "abc123",
		CookieAuth: CookieAuthConfig{
			CookieName: "PRAGA_TOKEN",
			Domain:     "localhost",
			Secure:     false,
		},
		Auth: AuthConfig{
			Mode: "email",
		},
		Email: EmailConfig{
			EmailProvider: "mailjet",
			ValidDomains:  []string{"example.com"},
			ValidEmails:   []string{},
			From:          "auth@example.com",
			FromName:      "Example Auth",
		},
		Mailjet: MailjetConfig{
			APIKeyPublic:  "",
			APIKeyPrivate: "",
		},
		Server: ServerConfig{
			ListenType: "http",
			Socket:     "",
			Host:       "0.0.0.0",
			Port:       8086,
		},
		JWT: JWTConfig{
			ValidSeconds: 86400,
		},
	}
}

func TestRouteEmailVerifyFail(t *testing.T) {
	// Test that invalid code does not verify
	buffer := bytes.NewBuffer([]byte{})
	err := json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: "user@example.com",
		Code:  "abcd1234",
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 400

	if recorder.Result().StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteEmailVerifyOk(t *testing.T) {
	// Test that a valid code verifies
	buffer := bytes.NewBuffer([]byte{})
	email := "user@example.com"
	err := json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: email,
		Code:  MakeVerifyCodeNow(testServer.Config.SigningKey, email),
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 204
	result := recorder.Result()

	if result.StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}

	cookies := result.Cookies()
	fmt.Printf("%s = %s\n", cookies[0].Name, cookies[0].Value)
}

func TestRouteEmailVerifyPrevious(t *testing.T) {
	// Test that previous time chunk's code passes verification
	buffer := bytes.NewBuffer([]byte{})
	email := "user@example.com"
	err := json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: email,
		Code:  MakeVerifyCodeTS(testServer.Config.SigningKey, email, time.Now().Add(-timeChunks)),
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 204

	if recorder.Result().StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteEmailVerifyExpired(t *testing.T) {
	// Test that timeChunks * 2 -old code fails verification
	buffer := bytes.NewBuffer([]byte{})
	email := "user@example.com"
	err := json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: email,
		Code:  MakeVerifyCodeTS(testServer.Config.SigningKey, email, time.Now().Add(-timeChunks*2)),
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 400

	if recorder.Result().StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteEmailVerifyUpcoming(t *testing.T) {
	// Test that trying to verify upcoming codes will fail
	buffer := bytes.NewBuffer([]byte{})
	email := "user@example.com"
	err := json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: email,
		Code:  MakeVerifyCodeTS(testServer.Config.SigningKey, email, time.Now().Add(timeChunks)),
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 400

	if recorder.Result().StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteEmailSendVerify(t *testing.T) {
	// Test that you can verify a newly requested code
	buffer := bytes.NewBuffer([]byte{})
	email := "user@example.com"
	err := json.NewEncoder(buffer).Encode(emailSendRequest{
		Email: email,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/api/email/send", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 204
	result := recorder.Result()

	if result.StatusCode != expectedStatus {
		t.Errorf("/api/email/send returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}

	// Test that you can verify a newly requested code
	buffer = bytes.NewBuffer([]byte{})
	err = json.NewEncoder(buffer).Encode(emailVerifyRequest{
		Email: email,
		Code:  testLastSentCode,
	})
	if err != nil {
		t.Fatal(err)
	}

	req, err = http.NewRequest("POST", "/api/email/verify", buffer)
	if err != nil {
		t.Fatal(err)
	}

	recorder = httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus = 204

	if recorder.Result().StatusCode != expectedStatus {
		t.Errorf("/api/email/verify returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteVerifyTokenNoCookie(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/verify-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 401
	result := recorder.Result()

	if result.StatusCode != expectedStatus {
		t.Errorf("/api/verify-token returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteVerifyTokenBadCookie(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/verify-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	http.SetCookie(recorder, &http.Cookie{Name: testServer.Config.CookieAuth.CookieName, Value: "invalid"})
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 401
	result := recorder.Result()

	if result.StatusCode != expectedStatus {
		t.Errorf("/api/verify-token returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}

func TestRouteVerifyTokenOk(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/verify-token", nil)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req.AddCookie(makeAuthCookie(&testServer, "user@example.com"))
	testRouter.ServeHTTP(recorder, req)
	expectedStatus := 204
	result := recorder.Result()

	if result.StatusCode != expectedStatus {
		t.Errorf("/api/verify-token returned status %d, expected %d", recorder.Result().StatusCode, expectedStatus)
	}
}
