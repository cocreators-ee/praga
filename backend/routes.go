package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var testLastSentCode = ""

func authFailed(w http.ResponseWriter) {
	w.WriteHeader(401)
}

func newAuthCookie(srv *Server) *http.Cookie {
	cookie := &http.Cookie{
		Name:     srv.Config.CookieAuth.CookieName,
		HttpOnly: true,
		Secure:   srv.Config.CookieAuth.Secure,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	if srv.Config.CookieAuth.Domain != "localhost" {
		cookie.Domain = srv.Config.CookieAuth.Domain
	}

	return cookie
}

func clearAuthCookie(srv *Server, w http.ResponseWriter) {
	cookie := newAuthCookie(srv)
	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	http.SetCookie(w, cookie)
}

func makeAuthCookie(srv *Server, email string) *http.Cookie {
	token, err := MakeToken(srv, email)
	if err != nil {
		if DEBUG {
			log.Printf("Error making token: %s\n", err)
		}
		return nil
	}

	cookie := newAuthCookie(srv)
	cookie.Value = token
	cookie.Expires = time.Now().Add(time.Duration(srv.Config.JWT.ValidSeconds) * time.Second)
	return cookie
}

func setAuthCookie(srv *Server, email string, w http.ResponseWriter) {
	cookie := makeAuthCookie(srv, email)
	if cookie != nil {
		http.SetCookie(w, cookie)
	}
}

func validateToken(srv *Server, token string) bool {
	_, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Ensure signing method is correct
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(srv.Config.SigningKey), nil
	})

	if err != nil {
		if DEBUG {
			log.Printf("Token validation failed: %s\n", err)
		}
		return false
	}

	return true
}

func MakeToken(srv *Server, email string) (string, error) {
	expireDuration := time.Duration(srv.Config.JWT.ValidSeconds) * time.Second

	claims := &jwt.RegisteredClaims{
		Subject:   email,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
		ID:        uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(srv.Config.SigningKey))
	if err != nil {
		if DEBUG {
			log.Printf("Failed to sign token: %s\n", err)
		}
		return "", err
	}
	return tokenString, nil
}

func sendCode(srv *Server, email string, code string) {
	if DEBUG {
		log.Printf("New code for %s: %s", email, code)
	}

	testLastSentCode = code

	// Skip @example.com - e.g. for tests
	if strings.HasSuffix(email, "@example.com") {
		return
	}

	if srv.Config.Auth.EmailProvider == "mailjet" {
		srv.MailjetSender.sendEmailViaMailjet(email, srv.Config.Brand, code, srv.Config.Support)
	}
}

func validateRequest(payload interface{}) bool {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	if err := validate.Struct(payload); err != nil {
		var validationErrors validator.ValidationErrors
		if DEBUG {
			if errors.As(err, &validationErrors) {
				for _, verr := range validationErrors {
					log.Print(strings.Replace(verr.Error(), "Config.", "", 1))
				}
			} else {
				log.Print(err)
			}
		}
		return false
	}

	return true
}

func registerRoutes(srv *Server, r *chi.Mux) {
	// Get relevant configuration for frontend
	r.Get("/api/config", func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(ConfigResponse{
			Title:   srv.Config.Title,
			Brand:   srv.Config.Brand,
			Support: srv.Config.Support,
		})

		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		} else {
			w.WriteHeader(200)
		}
	})

	// Verify token from Nginx requests
	r.Get("/api/verify-token", func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie(srv.Config.CookieAuth.CookieName)
		if err != nil {
			// Failed to read cookie
			if DEBUG {
				if errors.Is(err, http.ErrNoCookie) {
					log.Printf("Verify request missing cookies %s\n", srv.Config.CookieAuth.CookieName)
				} else {
					log.Printf("Error getting cookie %s\n", err)
				}
			}

			authFailed(w)
			return
		}

		if !validateToken(srv, token.Value) {
			// Token validation failed - clear and report error
			clearAuthCookie(srv, w)
			authFailed(w)
			return
		}

		// Token validated successfully, report success
		w.WriteHeader(204)
	})

	// Send a new code
	r.Post("/api/email/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(400)
			return
		}

		var req EmailSendRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		if !validateRequest(req) {
			w.WriteHeader(400)
			return
		}

		// Check if the given email is valid
		validEmail := false
		for _, domain := range srv.Config.Email.ValidDomains {
			if strings.HasSuffix(req.Email, "@"+domain) {
				validEmail = true
				break
			}
		}

		if !validEmail {
			for _, email := range srv.Config.Email.ValidEmails {
				if req.Email == email {
					validEmail = true
					break
				}
			}
		}

		// Only send if the email is valid
		if validEmail {
			code := MakeVerifyCodeNow(srv.Config.SigningKey, req.Email)
			sendCode(srv, req.Email, code)
		} else {
			if DEBUG {
				log.Printf("%s is not allowed to log in", req.Email)
			}
		}

		// Always report success, we don't want to expose if the email is valid or not
		w.WriteHeader(204)
	})

	// Verify code
	r.Post("/api/email/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(400)
			return
		}

		var req EmailVerifyRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			w.WriteHeader(400)
			return
		}

		if !validateRequest(req) {
			w.WriteHeader(400)
			return
		}

		if CheckVerifyCode(req.Code, srv.Config.SigningKey, req.Email) {
			setAuthCookie(srv, req.Email, w)
			w.WriteHeader(204)
		} else {
			w.WriteHeader(400)
		}
	})
}
