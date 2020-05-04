package services

import (
	"errors"
	"github.com/gofiber/fiber"
	"goapi/errors/errorCodes"
	"goapi/errors/errorDesc"
	"goapi/models"
	"goapi/repositories"
)

type UserService interface {
	Insert(models.User) (statusCode int, insertedUserID string, err error, errorCode string)
	GetAll(limit int) (users []models.User, err error)
	GetByID(id string) (user models.User, err error)
	UpdateByID(id string, userUpdates models.User) (statusCode int, hasBeenUpdated bool, err error, errorCode string)
	DeleteByID(id string) (hasBeenDeleted bool, err error)
}

// NewUserService returns the default user service.
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

type userService struct {
	repo repositories.UserRepository
}

// Insert an user
// This will check if the email domain provided is ok, the email address is not already taken
// Password is hashed and salted with the security methods in the AuthService
func (s *userService) Insert(user models.User) (statusCode int, insertedUserID string, err error, errorCode string) {
	// Checks if email domain is good
	if !validEmailAddress(user.Email) {
		return fiber.StatusNotAcceptable, "failed", errors.New(errorDesc.EmailAddressDomainForbidden), errorCodes.EmailAddressDomainForbidden
	}

	// Checks if the email address given by the user already exists our database
	emailAddressAlreadyTaken, err := s.repo.EmailAddressExists(user.Email)
	if err != nil {
		return fiber.StatusInternalServerError, "failed", err, errorCodes.InternalServerError
	}
	if emailAddressAlreadyTaken {
		return fiber.StatusConflict, "failed", errors.New(errorDesc.EmailAddressAlreadyExists), errorCodes.EmailAddressAlreadyExists
	}
	salt, _ := generateSalt(32) // salt is []byte

	// Creates a salt and hashAndSalt the password given by user
	user.Salt = string(salt)
	user.Password = hashAndSalt([]byte(user.Password), salt)
	user.Verified = false
	user.Enabled = true

	insertedUserID, err = s.repo.Insert(user)
	if err != nil {
		return fiber.StatusInternalServerError, errorDesc.Unknown, err, errorCodes.Unknown
	}
	return fiber.StatusOK, insertedUserID, nil, ""
}

// Returns all users
// Use the limit parameter to limit the number of returned users
func (s *userService) GetAll(limit int) (users []models.User, err error) {
	return s.repo.SelectMany(limit)
}

// Returns an user by its id
func (s *userService) GetByID(id string) (user models.User, err error) {
	user, found := s.repo.SelectBy(id)
	if !found {
		return user, errors.New(errorDesc.ResourceNotFound)
	}
	return user, nil
}

// Update an user by its id
// If the email is requested to be updated, it will first check if the domain is valid
// and if it does not already exists
func (s *userService) UpdateByID(id string, user models.User) (statusCode int, hasBeenUpdated bool, err error, errorCode string) {

	// Checks if the email address given by the user already exists in the database
	if user.Email != "" {
		if !validEmailAddress(user.Email) {
			return fiber.StatusNotAcceptable, false, errors.New(errorDesc.EmailAddressDomainForbidden), errorCodes.EmailAddressDomainForbidden
		}
		emailAddressAlreadyTaken, err := s.repo.EmailAddressExists(user.Email)
		if err != nil {
			return fiber.StatusInternalServerError, false, err, errorCodes.InternalServerError
		}
		if emailAddressAlreadyTaken {
			return fiber.StatusConflict, false, errors.New(errorDesc.EmailAddressAlreadyExists), errorCodes.EmailAddressAlreadyExists
		}
	}
	// Check for password update and hash the new password if needed
	if user.Password != "" {
		salt, _ := generateSalt(32) // salt is []byte
		user.Salt = string(salt)
		user.Password = hashAndSalt([]byte(user.Password), salt)
	}

	hasBeenUpdated, err = s.repo.Update(id, user)
	if err != nil {
		return fiber.StatusNotFound, hasBeenUpdated, err, errorCodes.ResourceNotFound
	}
	return fiber.StatusOK, hasBeenUpdated, nil, ""
}

// Tells the UserRepository to delete an user by its id
func (s *userService) DeleteByID(id string) (hasBeenDeleted bool, err error) {
	hasBeenDeleted, err = s.repo.DeleteBy(id)
	if err != nil || !hasBeenDeleted {
		return false, err
	}
	return true, nil
}
