package middlewares

import (
	"encoding/base64"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/elithrar/simple-scrypt"
	"github.com/vivowares/eywa/Godeps/_workspace/src/github.com/zenazn/goji/web"
	. "github.com/vivowares/eywa/configs"
	. "github.com/vivowares/eywa/models"
	. "github.com/vivowares/eywa/utils"
	"net/http"
)

func AdminAuthenticator(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" || r.Method == "OPTIONS" {
			h.ServeHTTP(w, r)
		} else {
			if len(r.Header.Get("Authentication")) != 0 {
				tokenStr := r.Header.Get("Authentication")
				auth, err := DecryptAuthToken(tokenStr)
				if err != nil {
					Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
				} else {
					if auth.Username == Config().Security.Dashboard.Username {
						if asBytes, err := base64.URLEncoding.DecodeString(auth.TokenString); err != nil {
							Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
						} else {
							if scrypt.CompareHashAndPassword(asBytes, []byte(Config().Security.Dashboard.Password)) != nil {
								Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
							} else {
								c.Env["auth_token"] = auth
								h.ServeHTTP(w, r)
							}
						}
					} else {
						Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid username or password"})
					}
				}
			} else {
				Render.JSON(w, http.StatusUnauthorized, map[string]string{"error": "empty Authentication header"})
			}
		}
	}

	return http.HandlerFunc(fn)
}