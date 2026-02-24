package devidp

import (
	"context"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gopkg.in/yaml.v3"
)

const (
	defaultTokenTTL   = time.Hour
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

type Config struct {
	Issuer   string        `yaml:"issuer"`
	Listen   string        `yaml:"listen"`
	Audience string        `yaml:"audience"`
	TokenTTL time.Duration `yaml:"token_ttl"`
	Key      KeyConfig     `yaml:"key"`
	Clients  []Client      `yaml:"clients"`
}

type KeyConfig struct {
	KID        string `yaml:"kid"`
	PrivateKey string `yaml:"private_key"`
}

type Client struct {
	ID     string   `yaml:"id"`
	Secret string   `yaml:"secret"`
	Roles  []string `yaml:"roles"`
	Scopes []string `yaml:"scopes"`
}

func LoadConfig(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) applyDefaults() {
	if c.TokenTTL == 0 {
		c.TokenTTL = defaultTokenTTL
	}
}

func (c Config) validate() error {
	if c.Issuer == "" {
		return errors.New("issuer is required")
	}
	if c.Listen == "" {
		return errors.New("listen is required")
	}
	if c.Audience == "" {
		return errors.New("audience is required")
	}
	if c.Key.PrivateKey == "" {
		return errors.New("key.private_key is required")
	}
	if len(c.Clients) == 0 {
		return errors.New("clients is required")
	}
	return nil
}

func Run(ctx context.Context, cfg Config) error {
	server, err := NewServer(cfg)
	if err != nil {
		return err
	}
	return server.Run(ctx)
}

type Server struct {
	cfg     Config
	signer  jwk.Key
	jwks    jwk.Set
	clients map[string]Client
}

func NewServer(cfg Config) (*Server, error) {
	privateKey, err := loadRSAPrivateKey(cfg.Key.PrivateKey)
	if err != nil {
		return nil, err
	}

	signer, err := jwk.FromRaw(privateKey)
	if err != nil {
		return nil, fmt.Errorf("build jwk: %w", err)
	}
	if cfg.Key.KID != "" {
		if err := signer.Set(jws.KeyIDKey, cfg.Key.KID); err != nil {
			return nil, fmt.Errorf("set kid: %w", err)
		}
	}
	if err := signer.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		return nil, fmt.Errorf("set alg: %w", err)
	}

	publicKeyJWK, err := jwk.FromRaw(privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("build public jwk: %w", err)
	}
	if cfg.Key.KID != "" {
		if err := publicKeyJWK.Set(jws.KeyIDKey, cfg.Key.KID); err != nil {
			return nil, fmt.Errorf("set public kid: %w", err)
		}
	}
	if err := publicKeyJWK.Set(jwk.AlgorithmKey, jwa.RS256); err != nil {
		return nil, fmt.Errorf("set public alg: %w", err)
	}

	jwks := jwk.NewSet()
	if err := jwks.AddKey(publicKeyJWK); err != nil {
		return nil, fmt.Errorf("add jwk: %w", err)
	}

	clients := make(map[string]Client, len(cfg.Clients))
	for _, client := range cfg.Clients {
		clients[client.ID] = client
	}

	return &Server{
		cfg:     cfg,
		signer:  signer,
		jwks:    jwks,
		clients: clients,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", s.handleOIDCConfig)
	mux.HandleFunc("/jwks", s.handleJWKS)
	mux.HandleFunc("/token", s.handleToken)
	mux.HandleFunc("/healthz", s.handleHealth)

	server := &http.Server{
		Addr:              s.cfg.Listen,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info(
		"dev idp listening",
		slog.String("listen", s.cfg.Listen),
		slog.String("issuer", s.cfg.Issuer),
	)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) handleOIDCConfig(w http.ResponseWriter, _ *http.Request) {
	issuer := strings.TrimRight(s.cfg.Issuer, "/")
	cfg := map[string]any{
		"issuer":                                issuer,
		"jwks_uri":                              issuer + "/jwks",
		"token_endpoint":                        issuer + "/token",
		"grant_types_supported":                 []string{"client_credentials"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic", "client_secret_post"},
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.jwks)
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "invalid_request"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_request"})
		return
	}

	if grantType := r.Form.Get("grant_type"); grantType != "client_credentials" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported_grant_type"})
		return
	}

	client, err := s.authenticateClient(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid_client"})
		return
	}

	scope := strings.TrimSpace(r.Form.Get("scope"))
	accessToken, expiresIn, err := s.issueToken(client, scope)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "server_error"})
		return
	}

	resp := map[string]any{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int64(expiresIn.Seconds()),
	}
	if scope != "" {
		resp["scope"] = scope
	}

	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) authenticateClient(r *http.Request) (Client, error) {
	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
	}
	if clientID == "" || clientSecret == "" {
		return Client{}, errors.New("missing client credentials")
	}

	client, ok := s.clients[clientID]
	if !ok {
		return Client{}, errors.New("unknown client")
	}

	if subtle.ConstantTimeCompare([]byte(clientSecret), []byte(client.Secret)) != 1 {
		return Client{}, errors.New("invalid secret")
	}

	return client, nil
}

func (s *Server) issueToken(client Client, scope string) (string, time.Duration, error) {
	now := time.Now()

	builder := jwt.NewBuilder().
		Issuer(s.cfg.Issuer).
		Subject(client.ID).
		Audience([]string{s.cfg.Audience}).
		IssuedAt(now).
		Expiration(now.Add(s.cfg.TokenTTL))

	tok, err := builder.Build()
	if err != nil {
		return "", 0, err
	}

	if err := tok.Set("azp", client.ID); err != nil {
		return "", 0, err
	}
	if err := tok.Set("client_id", client.ID); err != nil {
		return "", 0, err
	}
	if len(client.Roles) > 0 {
		if err := tok.Set("roles", client.Roles); err != nil {
			return "", 0, err
		}
	}
	if scope != "" {
		if err := tok.Set("scope", scope); err != nil {
			return "", 0, err
		}
	}

	headers := jws.NewHeaders()
	if s.signer.KeyID() != "" {
		if err := headers.Set(jws.KeyIDKey, s.signer.KeyID()); err != nil {
			return "", 0, err
		}
	}

	alg := s.signer.Algorithm()
	if alg == nil {
		alg = jwa.RS256
	}
	signed, err := jwt.Sign(tok, jwt.WithKey(alg, s.signer, jws.WithProtectedHeaders(headers)))
	if err != nil {
		return "", 0, err
	}
	return string(signed), s.cfg.TokenTTL, nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func loadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key: %w", err)
	}

	block, _ := pem.Decode(raw)
	if block == nil {
		return nil, errors.New("failed to decode private key PEM")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	key, ok := parsed.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("unsupported private key type %T", parsed)
	}

	return key, nil
}
