package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goapi/models"
)

// HouseRepository handles the basic operations of a house entity/model.
type HouseRepository interface {
	Insert(house models.House) (string, error)

	SelectByID(id string) (house models.House, found bool)
	SelectByUserID(userId string) (houses []models.House, found bool)
	SelectMany(limit int) ([]models.House, error)

	Update(id string, houseUpdates models.House) (hasBeenUpdated bool, err error)

	DeleteByID(id string) (hasBeenDeleted bool, err error)
}

// NewHouseRepository returns a new house repository,
// Requires the collection corresponding to houses from the mongo database
func NewHouseRepository(collection *mongo.Collection) HouseRepository {
	return &houseRepository{collection: collection}
}

// houseRepository is a "HouseRepository"
// which manages the houses using the mongoDB collection
type houseRepository struct {
	collection *mongo.Collection
}

// Insert a house in database
func (f houseRepository) Insert(house models.House) (string, error) {
	insertOneResult, err := f.collection.InsertOne(context.TODO(), house)
	if err != nil {
		return "failed", err
	}
	insertedHouseID := insertOneResult.InsertedID.(primitive.ObjectID).Hex()
	return insertedHouseID, nil
}

// Select a house by its id from database
func (f houseRepository) SelectByID(id string) (house models.House, found bool) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	err := f.collection.FindOne(context.TODO(), filter).Decode(&house)
	if err != nil {
		return models.House{}, false // empty house object
	}
	return house, true
}

// Select houses by their userID from database
// As an user can have several houses, this function will return all of them
func (f houseRepository) SelectByUserID(userId string) ([]models.House, bool) {
	filter := bson.M{"userID": userId}
	findResult, err := f.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, false
	}
	var houses []models.House
	err = findResult.All(context.TODO(), &houses)
	if err != nil || len(houses) == 0 {
		return nil, false
	}
	return houses, true
}

// Select several houses from the database
// Use the limit parameter to limit the number of returned houses
func (f houseRepository) SelectMany(limit int) ([]models.House, error) {
	limit64 := int64(limit)
	option := options.FindOptions{Limit: &limit64}
	findResult, err := f.collection.Find(context.TODO(), bson.D{}, &option)
	if err != nil {
		return nil, err
	}
	var houses []models.House
	err = findResult.All(context.TODO(), &houses)
	if err != nil {
		return nil, err
	}
	return houses, nil
}

// Updates a houses in database
// Empty fields will not be updates (omitempty tag in model)
func (f houseRepository) Update(id string, house models.House) (hasBeenUpdated bool, err error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": house}
	updateResult := f.collection.FindOneAndUpdate(context.TODO(), filter, update)
	if updateResult.Err() != nil {
		return false, updateResult.Err() // house not found
	}
	return true, nil
}

// Deletes a house from database
func (f houseRepository) DeleteByID(id string) (bool, error) {
	objID, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": objID}
	_, err := f.collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return false, err
	}
	return true, nil
}
