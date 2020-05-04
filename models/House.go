package models

type House struct {
	ID     string  `json:"id" bson:"_id,omitempty"`
	UserID string  `json:"userID" bson:"userID,omitempty"`
	Name   string  `json:"name" bson:"name,omitempty"`
	City   string  `json:"city" bson:"city,omitempty"`
	Rooms  *[]Room `json:"rooms" bson:"rooms,omitempty"`
}
