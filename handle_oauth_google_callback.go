package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/avearmin/gorage-sale/internal/auth"
	"github.com/avearmin/gorage-sale/internal/database"
	"github.com/avearmin/gorage-sale/internal/oauth2"
	"github.com/google/uuid"
)

func (cfg apiConfig) handleOAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg.getOAuthGoogleCallback(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "")
	}
}

func (cfg apiConfig) getOAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	oauthState := r.FormValue("state")

	if valid := cfg.StateStore.ValidateState(oauthState); !valid {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	cfg.StateStore.DeleteState(oauthState)

	code := r.FormValue("code")
	data, err := oauth2.GetUserDataFromGoogle(cfg.ClientID, cfg.ClientSecret, cfg.OAuthRedirectURL, code, r.Context())
	if err != nil {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	id, err := cfg.DB.GetUserIDByEmail(r.Context(), data.Email)
	if err != nil {
		if err == sql.ErrNoRows { // If the user does not exist, then we create one
			user, err := cfg.createUser(r.Context(), data.Name, data.Email)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Error creating user in DB")
				return

			}
			id = user.ID
		} else {
			respondWithError(w, http.StatusInternalServerError, "Error accessing DB")
			return
		}
	}

	accessToken, err := auth.CreateAccessToken(id, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating access token")
		return
	}
	refreshToken, err := auth.CreateRefreshToken(id, cfg.JwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token")
		return
	}

	var payload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	payload.AccessToken = accessToken
	payload.RefreshToken = refreshToken

	respondWithJSON(w, http.StatusOK, payload)
}

func (cfg apiConfig) createUser(ctx context.Context, name, email string) (database.User, error) {
	user, err := cfg.DB.CreateUser(ctx, database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Email:     email,
	})
	return user, err
}
