package access

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/opentdf/platform/internal/auth"
	"github.com/opentdf/platform/internal/security"
	"github.com/opentdf/platform/protocol/go/authorization"

	kaspb "github.com/opentdf/platform/protocol/go/kas"
	"github.com/opentdf/platform/service/internal/auth"
	"github.com/opentdf/platform/service/internal/security"
	"github.com/opentdf/platform/service/kas/nanotdf"
	"github.com/opentdf/platform/service/kas/tdf3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const ivSize = 12
const tagSize = 12

type RequestBody struct {
	AuthToken       string         `json:"authToken"`
	KeyAccess       tdf3.KeyAccess `json:"keyAccess"`
	Policy          string         `json:"policy,omitempty"`
	Algorithm       string         `json:"algorithm,omitempty"`
	ClientPublicKey string         `json:"clientPublicKey"`
	PublicKey       interface{}    `json:"-"`
	SchemaVersion   string         `json:"schemaVersion,omitempty"`
}

type entityInfo struct {
	EntityID string `json:"sub"`
	ClientID string `json:"clientId"`
	Token    string `json:"-"`
}

const (
	ErrUser     = Error("request error")
	ErrInternal = Error("internal error")
)

func err400(s string) error {
	return errors.Join(ErrUser, status.Error(codes.InvalidArgument, s))
}

func err401(s string) error {
	return errors.Join(ErrUser, status.Error(codes.Unauthenticated, s))
}

func err403(s string) error {
	return errors.Join(ErrUser, status.Error(codes.PermissionDenied, s))
}

func err404(s string) error {
	return errors.Join(ErrUser, status.Error(codes.NotFound, s))
}

func err503(s string) error {
	return errors.Join(ErrInternal, status.Error(codes.Unavailable, s))
}

func generateHMACDigest(ctx context.Context, msg, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(msg)
	if err != nil {
		slog.WarnContext(ctx, "failed to compute hmac")
		return nil, errors.Join(ErrUser, status.Error(codes.InvalidArgument, "policy hmac"))
	}
	return mac.Sum(nil), nil
}

func verifySignedRequesToken(ctx context.Context, in *kaspb.RewrapRequest) (*RequestBody, jwt.Token, error) {
	// get dpop public key from context
	dpopJWK := auth.GetJWKFromContext(ctx)

	// if we don't have a dpop public key then we can't verify the request
	if dpopJWK == nil {
		slog.ErrorContext(ctx, "missing dpop public key")
		return nil, nil, err401("dpop public key missing")
	}

	// verify and validate the request token
	token, err := jwt.Parse([]byte(in.SignedRequestToken),
		jwt.WithKey(dpopJWK.Algorithm(), dpopJWK),
		jwt.WithValidate(true),
	)
	// we have failed to verify the signed request token
	if err != nil {
		slog.WarnContext(ctx, "unable to verify request token", "err", err)
		return nil, nil, err401("unable to verify request token")
	}

	rb, exists := token.Get("requestBody")
	if !exists {
		slog.WarnContext(ctx, "missing request body")
		return nil, nil, err400("missing request body")
	}

	var requestBody = new(RequestBody)

	err = json.Unmarshal([]byte(rb.(string)), &requestBody)
	if err != nil {
		slog.WarnContext(ctx, "invalid request body")
		return nil, nil, err400("invalid request body")
	}

	slog.DebugContext(ctx, "extract public key", "requestBody.ClientPublicKey", requestBody.ClientPublicKey)
	block, _ := pem.Decode([]byte(requestBody.ClientPublicKey))
	if block == nil {
		slog.WarnContext(ctx, "missing clientPublicKey")
		return nil, nil, err400("clientPublicKey failure")
	}

	// Try to parse the clientPublicKey
	clientPublicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		slog.WarnContext(ctx, "failure to parse clientPublicKey", "err", err)
		return nil, nil, err400("clientPublicKey parse failure")
	}
	// Check to make sure the clientPublicKey is a supported key type
	switch clientPublicKey.(type) {
	case *rsa.PublicKey:
		requestBody.PublicKey = clientPublicKey.(*rsa.PublicKey)
		return requestBody, token, nil
	case *ecdsa.PublicKey:
		requestBody.PublicKey = clientPublicKey.(*ecdsa.PublicKey)
		return requestBody, token, nil
	default:
		slog.WarnContext(ctx, fmt.Sprintf("clientPublicKey not a supported key, was [%T]", clientPublicKey))
		return nil, nil, err400("clientPublicKey unsupported type")
	}
}

func verifyAndParsePolicy(ctx context.Context, requestBody *RequestBody, k []byte) (*Policy, error) {
	actualHMAC, err := generateHMACDigest(context.Background(), []byte(requestBody.Policy), k)
	if err != nil {
		slog.WarnContext(ctx, "unable to generate policy hmac", "err", err)
		return nil, err400("bad request")
	}
	expectedHMAC := make([]byte, base64.StdEncoding.DecodedLen(len(requestBody.KeyAccess.PolicyBinding)))
	n, err := base64.StdEncoding.Decode(expectedHMAC, []byte(requestBody.KeyAccess.PolicyBinding))
	if err == nil {
		n, err = hex.Decode(expectedHMAC, expectedHMAC[:n])
	}
	expectedHMAC = expectedHMAC[:n]
	if err != nil {
		slog.WarnContext(ctx, "invalid policy binding", "err", err)
		return nil, err400("bad request")
	}
	if !hmac.Equal(actualHMAC, expectedHMAC) {
		slog.WarnContext(ctx, "policy hmac mismatch", "actual", actualHMAC, "expected", expectedHMAC, "policyBinding", requestBody.KeyAccess.PolicyBinding)
		return nil, err400("bad request")
	}
	sDecPolicy, err := base64.StdEncoding.DecodeString(requestBody.Policy)
	if err != nil {
		slog.WarnContext(ctx, "unable to decode policy", "err", err)
		return nil, err400("bad request")
	}
	decoder := json.NewDecoder(strings.NewReader(string(sDecPolicy)))
	var policy Policy
	err = decoder.Decode(&policy)
	if err != nil {
		slog.WarnContext(ctx, "unable to decode policy", "err", err)
		return nil, err400("bad request")
	}
	return &policy, nil
}

func getEntityInfo(ctx context.Context) (*entityInfo, error) {
	var info = new(entityInfo)

	// check if metadata exists. if it doesn't not sure how we got to this point
	md, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		slog.WarnContext(ctx, "missing metadata")
		return nil, errors.New("missing metadata")
	}

	// if token is missing something went wrong in the authn interceptor
	t, exists := md["token"]
	if !exists {
		slog.WarnContext(ctx, "missing authorization header")
		return nil, errors.New("missing authorization header")
	}

	token, err := jwt.ParseInsecure([]byte(t[0]))
	if err != nil {
		slog.WarnContext(ctx, "unable to get token")
		return nil, errors.New("unable to get token")
	}

	sub, found := token.Get("sub")
	if found {
		info.EntityID = sub.(string)
	} else {
		slog.WarnContext(ctx, "missing sub")
	}

	// We have to check for the different ways the clientID can be stored in the token
	clientID, found := token.Get("clientId")
	if found {
		info.ClientID = clientID.(string)
	}

	clientID, found = token.Get("cid")
	if found {
		info.ClientID = clientID.(string)
	}

	clientID, found = token.Get("client_id")
	if found {
		info.ClientID = clientID.(string)
	}

	info.Token = string(t[0])

	return info, nil
}

func (p *Provider) Rewrap(ctx context.Context, in *kaspb.RewrapRequest) (*kaspb.RewrapResponse, error) {
	slog.DebugContext(ctx, "REWRAP")

	body, token, err := verifySignedRequesToken(ctx, in)
	if err != nil {
		return nil, err
	}

	entityInfo, err := getEntityInfo(ctx)
	if err != nil {
		return nil, err
	}

	if !strings.HasPrefix(body.KeyAccess.URL, p.URI.String()) {
		slog.InfoContext(ctx, "mismatched key access url", "keyAccessURL", body.KeyAccess.URL, "kasURL", p.URI.String())
	}

	if body.Algorithm == "" {
		body.Algorithm = "rsa:2048"
	}

	if body.Algorithm == "ec:secp256r1" {
		return nanoTDFRewrap(body, &p.Session, p.Session.EC.PrivateKey)
	}
	return p.tdf3Rewrap(ctx, body, token, entityInfo)
}

func (p *Provider) tdf3Rewrap(ctx context.Context, body *RequestBody, token jwt.Token, entity *entityInfo) (*kaspb.RewrapResponse, error) {
	symmetricKey, err := p.Session.DecryptOAEP(
		&p.Session.RSA.PrivateKey, body.KeyAccess.WrappedKey, crypto.SHA1, nil)
	if err != nil {
		slog.WarnContext(ctx, "failure to decrypt dek", "err", err)
		return nil, err400("bad request")
	}

	slog.DebugContext(ctx, "verifying policy binding", "requestBody.policy", body.Policy)
	policy, err := verifyAndParsePolicy(ctx, body, symmetricKey)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "extracting policy", "requestBody.policy", body.Policy)
	// changed to ClientID from Subject
	ent := authorization.Entity{
		EntityType: &authorization.Entity_Jwt{
			Jwt: token,
		},
	}
	if entity.ClientID != "" {
		ent = authorization.Entity{
			EntityType: &authorization.Entity_ClientId{
				ClientId: entity.ClientID,
			},
		}
	}

	access, err := canAccess(ctx, ent, *policy, p.SDK)

	if err != nil {
		slog.WarnContext(ctx, "Could not perform access decision!", "err", err)
		return nil, err403("forbidden")
	}

	if !access {
		slog.WarnContext(ctx, "Access Denied; no reason given")
		return nil, err403("forbidden")
	}

	rewrappedKey, err := tdf3.EncryptWithPublicKey(symmetricKey, body.PublicKey.(*rsa.PublicKey))
	if err != nil {
		slog.WarnContext(ctx, "rewrap: encryptWithPublicKey failed", "err", err, "clientPublicKey", &body.ClientPublicKey)
		return nil, err400("bad key for rewrap")
	}

	return &kaspb.RewrapResponse{
		EntityWrappedKey: rewrappedKey,
		SessionPublicKey: "",
		SchemaVersion:    schemaVersion,
	}, nil
}

func nanoTDFRewrap(body *RequestBody, session *security.HSMSession, key security.PrivateKeyEC) (*kaspb.RewrapResponse, error) {
	headerReader := bytes.NewReader(body.KeyAccess.Header)

	header, err := nanotdf.ReadNanoTDFHeader(headerReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse NanoTDF header: %w", err)
	}

	symmetricKey, err := session.GenerateNanoTDFSymmetricKey(header.EphemeralPublicKey.Key, key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	pub, ok := body.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to extract public key: %w", err)
	}

	// Convert public key to 65-bytes format
	pubKeyBytes := make([]byte, 1+len(pub.X.Bytes())+len(pub.Y.Bytes()))
	pubKeyBytes[0] = 0x4 // ID for uncompressed format
	if copy(pubKeyBytes[1:33], pub.X.Bytes()) != 32 || copy(pubKeyBytes[33:], pub.Y.Bytes()) != 32 {
		return nil, fmt.Errorf("failed to serialize keypair: %v", pub)
	}

	privateKeyHandle, publicKeyHandle, err := session.GenerateEphemeralKasKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}
	sessionKey, err := session.GenerateNanoTDFSessionKey(privateKeyHandle, pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	cipherText, err := wrapKeyAES(sessionKey, symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	// see explanation why Public Key starts at position 2
	//https://github.com/wqx0532/hyperledger-fabric-gm-1/blob/master/bccsp/pkcs11/pkcs11.go#L480
	pubGoKey, err := ecdh.P256().NewPublicKey(publicKeyHandle[2:])
	if err != nil {
		return nil, fmt.Errorf("failed to make public key") // Handle error, e.g., invalid public key format
	}

	pbk, err := x509.MarshalPKIXPublicKey(pubGoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert public Key to PKIX")
	}

	pemBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pbk,
	}
	pemString := string(pem.EncodeToMemory(pemBlock))

	return &kaspb.RewrapResponse{
		EntityWrappedKey: cipherText,
		SessionPublicKey: pemString,
		SchemaVersion:    schemaVersion,
	}, nil
}

func wrapKeyAES(sessionKey, dek []byte) ([]byte, error) {
	block, err := aes.NewCipher(sessionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	aesGcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create NewGCMWithTagSize: %w", err)
	}

	iv := make([]byte, ivSize)
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	cipherText := aesGcm.Seal(iv, iv, dek, nil)
	return cipherText, nil
}
