package services

import (
	"crypto/rand"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"go.mongodb.org/mongo-driver/mongo"
	"goapi/config"
	"goapi/errors/errorCodes"
	"goapi/errors/errorDesc"
	"goapi/repositories"
	"golang.org/x/crypto/argon2"
	"strings"
	"time"
)

type AuthService interface {
	Login(emailAddress string, password string) (credentialMatch bool, signedToken string, statusCode int, err error, errorCode string)

	JwtGenerate(userID string) jwt.Token
	JwtVerifyCanBeRefreshed(token *jwt.Token) bool

	ExtractJWTString(ctx *fiber.Ctx) (string, error)
	ExtractJWT(ctx *fiber.Ctx) (*jwt.Token, error)

	JWTRefresh(ctx *fiber.Ctx) (jwt string, statusCode int, err error, errorCode string)
}

// NewAuthService returns the default auth service.
func NewAuthService(repo repositories.UserRepository) AuthService {
	return &authService{
		repo: repo,
	}
}

type authService struct {
	repo repositories.UserRepository
}

// Login method
// From the emailAddress given by client, it will try to find the salt and password
// from the user in the database
// Provided password will then be hashed and salted
// If it matches with the user password, a new JWT is sent
//
// NOTE: for optimal security the client application must always tells the user
// "email or password incorrect", even if the user does not exists in database!
func (a authService) Login(emailAddress string, providedPassword string) (bool, string, int, error, string) {
	// Looks for the user salt and password in database
	userID, userPassword, salt, err := a.repo.SelectForLogin(emailAddress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, "", fiber.StatusUnauthorized, fmt.Errorf(errorDesc.CredentialDoesNotMatch), errorCodes.CredentialDoesNotMatch
		}
		return false, "", fiber.StatusInternalServerError, err, errorCodes.InternalServerError
	}

	// Check if pass are same, if not return error
	if !verifyPasswordMatch(providedPassword, salt, userPassword) {
		return false, "", fiber.StatusUnauthorized, fmt.Errorf(errorDesc.CredentialDoesNotMatch), errorCodes.CredentialDoesNotMatch
	}

	// Generates a new token for user
	newToken := a.JwtGenerate(userID)
	newTokenString, err := newToken.SignedString([]byte(config.HmacSampleSecret))
	if err != nil {
		return false, "", fiber.StatusInternalServerError, err, errorCodes.InternalServerError
	}
	return true, newTokenString, fiber.StatusOK, nil, ""
}

// Generates a new JWT
// A JWT contains the id of the user and the time it will expire which is calculated according
// to the duration set in the jwtDuration.yml file
func (a authService) JwtGenerate(userID string) jwt.Token {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		//"generationTime": time.Now().Format(time.RFC3339),
		"sub": userID,
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute * time.Duration(config.JWTExpirationTimeInMinutes)).Unix(),
	})
	/*
		// Code to sign the token and convert it to string
		tokenString, err := token.SignedString([]byte(config.HmacSampleSecret))
		if err != nil {
			return "", err
		}*/
	return *token
}

// Extract the JWT from the authorization header and returns it as a string
// If no JWT were provided, returns an error
func (a authService) ExtractJWTString(ctx *fiber.Ctx) (string, error) {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf(errorDesc.NoTokenWereProvided)
	}

	authHeaderParts := strings.Split(authHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		return "", fmt.Errorf(errorDesc.AuthorizationHeaderMustBeBearerToken)
	}

	return authHeaderParts[1], nil
}

// Extract the JWT from the authorization header and returns a jwt.Token object
// If no JWT or a bad JWT were provided, returns an error
// The token will be returned even if it is expired
func (a authService) ExtractJWT(ctx *fiber.Ctx) (*jwt.Token, error) {
	tokenString, err := a.ExtractJWTString(ctx)

	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(errorDesc.TokenSignatureIsNotValid)
		}
		return []byte(config.HmacSampleSecret), nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors != jwt.ValidationErrorExpired {
				return nil, err // if the error is other than token is expired
			}
		}
	}
	return token, nil // returns the token even if it is expired
}

// Refresh the JWT
// Provided JWT is extracted from context and analyzed
// JWT will not be refreshed if it is still valid or goes beyond the refresh deadline
func (a authService) JWTRefresh(ctx *fiber.Ctx) (tokenString string, statusCode int, err error, errorCode string) {
	token, err := a.ExtractJWT(ctx)
	if err != nil {
		return "", fiber.StatusBadRequest, err, errorCodes.BadRequest
	}
	if token != nil && token.Valid { // Token is valid
		return "", fiber.StatusBadRequest, fmt.Errorf(errorDesc.JWTIsStillValid), errorCodes.JWTIsStillValid
	}

	canBeRefreshed := a.JwtVerifyCanBeRefreshed(token)
	if !canBeRefreshed {
		return "", fiber.StatusUnauthorized, fmt.Errorf(errorDesc.JWTExpiredCannotBeRefreshed), errorCodes.JWTExpiredCannotBeRefreshed
	}

	// From here token can be refreshed
	userID := token.Claims.(jwt.MapClaims)["sub"].(string)
	_, found := a.repo.SelectBy(userID)
	if !found { // user not found, probably disabled
		return "", fiber.StatusUnauthorized, err, errorCodes.ResourceNotFound
	}

	newToken := a.JwtGenerate(userID)
	newTokenString, err := newToken.SignedString([]byte(config.HmacSampleSecret))
	return newTokenString, fiber.StatusOK, nil, ""
}

// Verify that a JWT can be refreshed, according to the duration set in the jwtDuration.yml file
func (a authService) JwtVerifyCanBeRefreshed(token *jwt.Token) bool {
	claims := token.Claims.(jwt.MapClaims)
	exp := int64(claims["exp"].(float64))
	expirationTime := time.Unix(time.Now().Unix(), 0).Sub(time.Unix(exp, 0))
	// println("expired from", expirationTime.String())
	expirationDeadline := config.JWTRefreshDeadlineInHours * time.Hour
	if expirationTime < expirationDeadline {
		return true
	}
	return false
}

// passwordSent: the unchanged password receive from request
// salt: salt of the user, got from database
// userPassword: the good and hashed password of the user, got from database
// This function will hashAndSalt and salt the provided password and check it it match with the user's password
func verifyPasswordMatch(passwordSent string, salt string, userPassword string) bool {
	hashedPassword := hashAndSalt([]byte(passwordSent), []byte(salt))
	if hashedPassword == userPassword {
		return true
	}
	return false
}

// Examine the domain of the email address
// This aims to prevent users from using throwable email addresses
func validEmailAddress(email string) bool {
	components := strings.Split(email, "@")
	domain := components[1]
	fmt.Printf("Domain: %s\n", domain)
	if domain == "" {
		return false
	}
	return true
}

// Generate a salt
// Using the crypto/rand dependency, verified CSPRNG
func generateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, err
}

// Parameters for argon2id
type argon2idParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// Hash  and salt the password
// Using Argon2id, one of the best current hash methods
func hashAndSalt(password []byte, salt []byte) string {
	p := &argon2idParams{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  32,
		keyLength:   32,
	}
	hash := argon2.IDKey(password, salt, p.iterations, p.memory, p.parallelism, p.keyLength)
	return string(hash)
}
