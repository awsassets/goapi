package controllers

import (
	"github.com/gofiber/fiber"
	"goapi/config"
	"goapi/errors/errorCodes"
	"goapi/errors/errorDesc"
	"goapi/models"
	"goapi/services"
)

type UserController struct {
	UserService services.UserService
	AuthService services.AuthService
}

// Post an user
// This method can be used a register
// Note that you must provide a firstName, lastName, email, password and language
// These are the required fields to insert an user in the database
// If everything is good it will return a JWT to interact with the other routes
// POST http://localhost:5000/users
func (c *UserController) Post(ctx *fiber.Ctx) {
	user := models.User{}
	err := ctx.BodyParser(&user)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.BadRequest,
		})
		return
	}

	// Returns an error if one of the required fields is empty
	if user.FirstName == "" || user.LastName == "" || user.Email == "" || user.Password == "" || user.Language == "" {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     errorDesc.RequiredFieldEmpty,
			"errorCode": errorCodes.BadRequest,
		})
		return
	}

	statusCode, insertedUserID, err, errorCode := c.UserService.Insert(user)
	if err != nil {
		_ = ctx.Status(statusCode).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCode,
		})
		return
	}

	// Generate the JWT and sign it
	token := c.AuthService.JwtGenerate(insertedUserID)
	tokenString, err := token.SignedString([]byte(config.HmacSampleSecret))
	if err != nil {
		_ = ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.InternalServerError,
		})
		return
	}
	data := make(map[string]string)
	data["token"] = tokenString
	data["userID"] = insertedUserID
	_ = ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// DEV STATUS ONLY
// Returns list of the users
// GET http://localhost:5000/users
func (c *UserController) GetAll(ctx *fiber.Ctx) {
	users, err := c.UserService.GetAll(config.LimitElementsReturnedFromDatabase)
	if err != nil {
		_ = ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.InternalServerError,
		})
		return
	}
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    users,
	})
}

// Returns an user by its id
// GET http://localhost:5000/users/id
func (c *UserController) GetByID(ctx *fiber.Ctx) {
	id := ctx.Params("id")
	user, err := c.UserService.GetByID(id)
	if err != nil {
		_ = ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.ResourceNotFound,
		})
		return
	}
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// Updates an user (JSON accepted only)
// This method can be used to updates the firstName, lastName, email or language fields
// You can update the ones that you want
// PATCH http://localhost:5000/users/id
func (c *UserController) PatchBy(ctx *fiber.Ctx) {
	// Map the fields in a user object
	id := ctx.Params("id")
	var user models.User
	err := ctx.BodyParser(&user)
	user.ID = ""
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.BadRequest,
		})
		return
	}

	// Send the update request to service and parse results
	statusCode, hasBeenUpdated, err, errorCode := c.UserService.UpdateByID(id, user)
	if !hasBeenUpdated {
		_ = ctx.Status(statusCode).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCode,
		})
		return
	}
	data := make(map[string]string)
	data["updatedID"] = id
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// DeleteUserBy deletes an user
// DELETE http://localhost:5000/users/id
func (c *UserController) DeleteBy(ctx *fiber.Ctx) {
	id := ctx.Params("id")
	hasBeenDeleted, err := c.UserService.DeleteByID(id)
	if err != nil {
		_ = ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.ResourceNotFound,
		})
		return
	}
	if !hasBeenDeleted { // Should never occur but just in case
		_ = ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":   false,
			"error":     errorDesc.Unknown,
			"errorCode": errorCodes.InternalServerError,
		})
		return
	}
	data := make(map[string]string)
	data["deletedID"] = id
	_ = ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}
