package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GoogleProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

func NewGoogleProvider(clientID, clientSecret, redirectURI string) *GoogleProvider {
	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *GoogleProvider) Name() string {
	return "google"
}

func (p *GoogleProvider) Authenticate(ctx context.Context, code string) (*OAuthUser, error) {
	accessToken, err := p.exchangeCode(ctx, code)
	if err != nil {
		return nil, err
	}

	return p.fetchUser(ctx, accessToken)
}

func (p *GoogleProvider) fetchUser(
	ctx context.Context,
	accessToken string,
) (*OAuthUser, error) {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("google: failed to fetch user info")
	}

	var payload struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	if payload.Email == "" {
		return nil, errors.New("google: email not returned")
	}

	return &OAuthUser{
		Email: payload.Email,
		Name:  payload.Name,
	}, nil
}

func (p *GoogleProvider) exchangeCode(
	ctx context.Context,
	code string,
) (string, error) {

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.clientID)
	form.Set("client_secret", p.clientSecret)
	form.Set("redirect_uri", p.redirectURI)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := p.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", errors.New("google: token exchange failed")
	}

	var payload struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}

	if payload.AccessToken == "" {
		return "", errors.New("google: empty access token")
	}

	return payload.AccessToken, nil
}
