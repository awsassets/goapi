package controllers

import (
	"github.com/gofiber/fiber"
	"goapi/config"
	"goapi/errors/errorCodes"
	"goapi/services"
)

type AuthController struct {
	AuthService services.AuthService
	UserService services.UserService
}

// For Register see POST in UserController

// Login method
// Parse the email and password provided and pass them to the AuthService.Login method
// If credentials match, a new JWT is sent, otherwise an error is sent
// POST: http://localhost:8080/auth/login
func (c *AuthController) Login(ctx *fiber.Ctx) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := ctx.BodyParser(&credentials)

	credentialMatch, newTokenSigned, statusCode, err, errorCode := c.AuthService.Login(credentials.Email, credentials.Password)
	if !credentialMatch {
		_ = ctx.Status(statusCode).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCode,
		})
		return
	}

	data := make(map[string]string)
	data["token"] = newTokenSigned
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// JWT refresh method
// Send the context to AuthService.JWTRefresh method that will parse the JWT and give a new one if possible
// GET: http://localhost:8080/auth/refreshjwt
func (c *AuthController) RefreshJWT(ctx *fiber.Ctx) {
	newToken, statusCode, err, errorCode := c.AuthService.JWTRefresh(ctx)

	if err != nil {
		_ = ctx.Status(statusCode).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCode,
		})
		return
	}
	data := make(map[string]string)
	data["token"] = newToken
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// POST: http://localhost:8080/auth/resetpassword
func (c *AuthController) ResetPassword(ctx *fiber.Ctx) {

	// check get by email
	// if no email return false

	// check if pass are same
	// if not return false

	userID := "TODO"
	// generate new token
	newToken := c.AuthService.JwtGenerate(userID)
	newTokenString, err := newToken.SignedString([]byte(config.HmacSampleSecret))
	if err != nil {
		_ = ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.InternalServerError,
		})
		return
	}
	data := make(map[string]string)
	data["token"] = newTokenString
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

/*
// GetAll userID from token
user := ctx.Locals("user").(*jwt.Token)
claims := user.Claims.(jwt.MapClaims)
userID := claims["sub"].(string)
println("USER ID FROM TOKEN:", userID)
*/
