package models

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// globally available mongodb variable for access within each model
var DB *mongo.Database
