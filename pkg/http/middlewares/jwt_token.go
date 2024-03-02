package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/RafalSalwa/interview-app-srv/pkg/http/auth"
	"github.com/RafalSalwa/interview-app-srv/pkg/responses"

	"github.com/RafalSalwa/interview-app-srv/pkg/jwt"
	"github.com/gorilla/mux"
)

func ValidateJWTAccessToken(c auth.JWTConfig) mux.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authorizationHeader := r.Header.Get("Authorization")
			if authorizationHeader != "" {
				bearerToken := strings.Split(authorizationHeader, " ")

				if len(bearerToken) == 2 {
					sub, err := jwt.ValidateToken(bearerToken[1], c.Access.PublicKey)
					fmt.Printf("req: %#v\n", r.Header.Get("Authorization"))
					fmt.Println(sub, err)
					if err != nil {
						responses.RespondNotAuthorized(w, "Wrong access token")
						return
					}
					r.Header.Set("x-user-id", sub)
				}
			} else {
				responses.RespondNotAuthorized(w, "An authorization header is required")
				return
			}
			h.ServeHTTP(w, r)
		})
	}
}
