package controllers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"goapi/config"
	"goapi/errors/errorCodes"
	"goapi/errors/errorDesc"
	"goapi/models"
	"goapi/services"
)

type HouseController struct {
	Service services.HouseService
}

// Post a house
// Note you must provide a name to insert a house, the userID will be extracted from JWT
// "rooms" is an array of models.room objects, you can insert as many as you want
// Note you must provide a name and surface > 0 for each room
// POST http://localhost:5000/houses
func (c *HouseController) Post(ctx *fiber.Ctx) {
	var house models.House
	err := ctx.BodyParser(&house)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.BadRequest,
		})
		return
	}
	// Check if each room fields is filled
	oneRoomFieldEmpty := false
	for _, room := range *house.Rooms {
		if room.Name == "" || room.Surface == 0 {
			oneRoomFieldEmpty = true
		}
	}
	// GetAll the user ID from its JWT
	user := ctx.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	house.UserID = claims["sub"].(string)
	// Check if the required fields are filled
	if house.Name == "" || oneRoomFieldEmpty {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     errorDesc.RequiredFieldEmpty,
			"errorCode": errorCodes.BadRequest,
		})
		return
	}

	// UserID will be checked in service in order to be sure user exists
	statusCode, insertedHouseID, err, errorCode := c.Service.Insert(house)
	if err != nil {
		_ = ctx.Status(statusCode).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCode,
		})
		return
	}
	data := make(map[string]string)
	data["insertedHouseID"] = insertedHouseID
	_ = ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// DEV STATUS ONLY
// Returns a list of the houses
// GET http://localhost:5000/houses
func (c *HouseController) GetAll(ctx *fiber.Ctx) {
	houses, err := c.Service.GetAll(config.LimitElementsReturnedFromDatabase) // limits the nb of returned houses
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
		"data":    houses,
	})
}

// Returns house by its id
// GET http://localhost:5000/houses/id
func (c *HouseController) GetByID(ctx *fiber.Ctx) {
	id := ctx.Params("id")
	house, err := c.Service.GetByID(id)
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
		"data":    house,
	})
}

// Returns house by its id
// GET http://localhost:5000/houses/ofUser/id
func (c *HouseController) GetByUserID(ctx *fiber.Ctx) {
	id := ctx.Params("id")
	houses, err := c.Service.GetByUserID(id)
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
		"data":    houses,
	})
}

// Updates a house (JSON accepted only)
// This method can be used to updates the userID, name or rooms fields
// You can update the ones that you want, but a room must contain all its fields
// PATCH http://localhost:5000/houses/id
func (c *HouseController) PatchBy(ctx *fiber.Ctx) {
	// Map the fields in a house object
	id := ctx.Params("id")
	var house models.House
	err := ctx.BodyParser(&house)
	if err != nil {
		_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":   false,
			"error":     err.Error(),
			"errorCode": errorCodes.BadRequest,
		})
		return
	}
	if house.Rooms != nil {
		// Check if each room fields is well filled
		for _, room := range *house.Rooms {
			if room.Name == "" || room.Surface == 0 {
				_ = ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success":   false,
					"error":     errorDesc.RequiredFieldEmpty,
					"errorCode": errorCodes.BadRequest,
				})
				return
			}
		}
	}


	// Send the update request to service and parse results
	statusCode, hasBeenUpdated, err, errorCode := c.Service.UpdateByID(id, house)
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

// Deletes a house
// DELETE http://localhost:5000/houses/id
func (c *HouseController) DeleteBy(ctx *fiber.Ctx) {
	id := ctx.Params("id")
	hasBeenDeleted, err := c.Service.DeleteByID(id)
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
