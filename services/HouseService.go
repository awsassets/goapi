package services

import (
	"errors"
	"github.com/gofiber/fiber"
	"goapi/errors/errorCodes"
	"goapi/errors/errorDesc"
	"goapi/models"
	"goapi/repositories"
)

type HouseService interface {
	Insert(models.House) (statusCode int, insertedHouseID string, err error, errorCode string)

	GetAll(limit int) ([]models.House, error)
	GetByID(id string) (models.House, error)
	GetByUserID(id string) ([]models.House, error)

	UpdateByID(id string, updates models.House) (statusCode int, hasBeenUpdated bool, err error, errorCode string)

	DeleteByID(id string) (hasBeenDeleted bool, err error)
}

// NewHouseService returns the default house service.
func NewHouseService(houseRepo repositories.HouseRepository, userRepo repositories.UserRepository) HouseService {
	return &houseService{
		houseRepo: houseRepo,
		userRepo:  userRepo,
	}
}

type houseService struct {
	houseRepo repositories.HouseRepository
	userRepo  repositories.UserRepository
}

// Insert a house
// This will first check if the userID provided exists in the database
func (s *houseService) Insert(house models.House) (statusCode int, insertedHouseID string, err error, errorCode string) {
	_, found := s.userRepo.SelectBy(house.UserID)
	if !found {
		return fiber.StatusNotFound, "failed", errors.New(errorDesc.ResourceNotFound), errorCodes.ResourceNotFound
	}
	insertedHouseID, err = s.houseRepo.Insert(house)
	if err != nil {
		return fiber.StatusConflict, "failed", errors.New(errorDesc.ResourceNotFound), errorCodes.ResourceNotFound
	}
	return fiber.StatusCreated, insertedHouseID, nil, ""
}

// Returns all houses
// Use the limit parameter to limit the number of returned houses
func (s *houseService) GetAll(limit int) ([]models.House, error) {
	return s.houseRepo.SelectMany(limit)
}

// Returns a house by its id
func (s *houseService) GetByID(id string) (models.House, error) {
	house, found := s.houseRepo.SelectByID(id)
	if !found {
		return house, errors.New(errorDesc.ResourceNotFound)
	}
	return house, nil
}

// Returns houses of an user
func (s *houseService) GetByUserID(id string) ([]models.House, error) {
	houses, found := s.houseRepo.SelectByUserID(id)
	if !found {
		return houses, errors.New(errorDesc.ResourceNotFound)
	}
	return houses, nil
}

// Tells the HouseRepository to update a house by its id
func (s *houseService) UpdateByID(id string, updates models.House) (statusCode int, hasBeenUpdated bool, err error, errorCode string) {
	hasBeenUpdated, err = s.houseRepo.Update(id, updates)
	if err != nil {
		return fiber.StatusNotFound, hasBeenUpdated, err, errorCodes.ResourceNotFound
	}
	return fiber.StatusOK, hasBeenUpdated, nil, ""
}

// Tells the HouseRepository to delete a house by its id
func (s *houseService) DeleteByID(id string) (hasBeenDeleted bool, err error) {
	hasBeenDeleted, err = s.houseRepo.DeleteByID(id)
	if err != nil || !hasBeenDeleted {
		return false, err
	}
	return true, nil
}
