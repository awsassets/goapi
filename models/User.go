package models

// Password and salt fields will never be sent
type User struct {
	ID        string `json:"id" bson:"_id,omitempty"`
	FirstName string `json:"firstName" bson:"firstName,omitempty"`
	LastName  string `json:"lastName" bson:"lastName,omitempty"`
	Email     string `json:"email" bson:"email,omitempty"`
	Salt      string `json:"-" bson:"salt,omitempty"`
	Password  string `json:",omitempty" bson:"password,omitempty"`
	Language  string `json:"language" bson:"language,omitempty"`
	Verified  bool   `json:"verified" bson:"verified,omitempty"`
	Enabled   bool   `json:"enabled" bson:"enabled,omitempty"`
}
