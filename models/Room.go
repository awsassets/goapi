package models

type Room struct {
	Name           string `json:"name" bson:"name"`
	Description    string `json:"description" bson:"description"`
	Surface        int    `json:"surface" bson:"surface"`
}
