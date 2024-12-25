package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Minutes struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	SessionID primitive.ObjectID `bson:"session_id" json:"session_id"`
	Content   string             `bson:"content" json:"content"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`
	UpdatedBy primitive.ObjectID `bson:"updated_by" json:"updated_by"`
}
