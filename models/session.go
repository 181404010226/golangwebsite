// your-project/models/session.go
package models

import (
    "go.mongodb.org/mongo-driver/bson/primitive"
    "time"
)

type User struct {
    ID        primitive.ObjectID `json:"id" bson:"_id"`
    Username  string            `json:"username" bson:"username"`
    AvatarURL string            `json:"avatar_url" bson:"avatar_url"`
}

type Session struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id"`  // 改为 _id 而不是 id
    Name        string             `json:"name"`
    CreatedAt   time.Time          `json:"created_at"`
    Participants []Participant     `json:"participants"`
    Summaries   []Summary          `json:"summaries"`
}

type Participant struct {
    ID         primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
    Username   string             `json:"username"`
    AvatarURL  string             `json:"avatar_url"`
    Summarized bool               `json:"summarized"`
}

type Summary struct {
    ParticipantID primitive.ObjectID `json:"participant_id"`
    Content        string             `json:"content"`
    Comments       []Comment          `json:"comments"`
    CreatedAt      time.Time          `json:"created_at"`
}

type Comment struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
    Username  string             `json:"username" bson:"username"`
    AvatarURL string             `json:"avatar_url" bson:"avatar_url"`
    Content   string             `json:"content" bson:"content"`
    Stars     int                `json:"stars" bson:"stars"`
    CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}