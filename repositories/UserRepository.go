package repositories

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goapi/errors/errorDesc"
	"goapi/models"
)

// UserRepository handles the basic operations of a user entity/model.
type UserRepository interface {
	Insert(user models.User) (string, error)

	SelectBy(id string) (user models.User, found bool)
	SelectForLogin(emailAddress string) (userID string, password string, salt string, err error)
	SelectMany(limit int) ([]models.User, error)

	Update(id string, userUpdates models.User) (hasBeenUpdated bool, err error)

	DeleteBy(id string) (bool, error)

	EmailAddressExists(emailAddress string) (bool, error)
}

// NewUserRepository returns a new user repository,
// Requires the collection corresponding to users from the mongo database
func NewUserRepository(collection *mongo.Collection) UserRepository {
	return &userCollectionRepository{collection: collection}
}

// userCollectionRepository is a "UserRepository"
// which manages the fridges using the mongoDB collection
type userCollectionRepository struct {
	collection *mongo.Collection
}

// Insert an user in database
func (u userCollectionRepository) Insert(user models.User) (string, error) {
	insertOneResult, err := u.collection.InsertOne(context.TODO(), user)
	if err != nil {
		return "", err
	}
	insertedUserID := insertOneResult.InsertedID.(primitive.ObjectID).Hex()
	return insertedUserID, nil
}

// Select and return an user by its ID
func (u userCollectionRepository) SelectBy(id string) (user models.User, found bool) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "enabled": true}
	err := u.collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return models.User{}, false // empty user object
	}
	user.Password = ""
	return user, true
}

// Select an user by its email address
// This method is used for the login method, it will return its password salt and userID
func (u userCollectionRepository) SelectForLogin(emailAddress string) (userID string, password string, salt string, err error) {
	user := struct {
		ID       string `bson:"_id"`
		Password string `bson:"password"`
		Salt     string `bson:"salt"`
	}{}
	filter := bson.M{"email": emailAddress, "enabled": true}
	err = u.collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return "", "", "", err // empty user object
	}
	return user.ID, user.Password, user.Salt, nil
}

// Select several users from the database
// Use the limit parameter to limit the number of returned users
func (u userCollectionRepository) SelectMany(limit int) ([]models.User, error) {
	limit64 := int64(limit)
	option := options.FindOptions{Limit: &limit64}
	findResult, err := u.collection.Find(context.TODO(), bson.D{}, &option)
	if err != nil {
		return nil, err
	}
	var users []models.User
	err = findResult.All(context.TODO(), &users)
	if err != nil {
		return nil, err
	}
	for key := range users {
		users[key].Password = ""
	}
	return users, nil
}

// Updates an user in database
// Empty fields will not be updates (omitempty tag in model)
func (u userCollectionRepository) Update(id string, user models.User) (hasBeenUpdated bool, err error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID, "enabled": true}
	update := bson.M{"$set": user}

	updateResult := u.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if updateResult.Err() != nil {
		return false, updateResult.Err() // user not found
	}
	return true, nil
}

// Deletes an user from database
func (u userCollectionRepository) DeleteBy(id string) (bool, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	deleteOneResult, err := u.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return false, err
	}
	if deleteOneResult.DeletedCount != 1 {
		if deleteOneResult.DeletedCount == 0 {
			return false, fmt.Errorf(errorDesc.ResourceNotFound)
		}
		return false, fmt.Errorf(errorDesc.Unknown) // Should never occur but just in case
	}
	return true, nil
}

// Check if an email address already exists within the users
// Will return true if so
// This method is used to prevent new users to register with an
// already existing email address or users to update their email
// address with already existing ones
func (u userCollectionRepository) EmailAddressExists(email string) (bool, error) {
	filter := bson.M{"email": email}
	limit64 := int64(5) // limit to 5 because the number is not relevant here
	option := options.FindOptions{Limit: &limit64}
	findResult, err := u.collection.Find(context.TODO(), filter, &option)
	if err != nil {
		return false, err
	}
	var users []models.User
	err = findResult.All(context.TODO(), &users)
	if err != nil {
		return false, err
	}
	if len(users) > 0 {
		return true, nil
	}
	return false, nil
}
