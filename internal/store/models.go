package store

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Article struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Title     string             `bson:"title"`
	URL       string             `bson:"url"`
	Source    string             `bson:"source"` // e.g., "hackernews", "devto"
	Summary   string             `bson:"summary,omitempty"`
	Category  string             `bson:"category,omitempty"`
	Score     float64            `bson:"score"`
	Hashtags  []string           `bson:"hashtags,omitempty"`
	Image     string             `bson:"image,omitempty"`
	Posted    bool               `bson:"posted"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

func SaveArticle(ctx context.Context, article *Article) error {
	collection := DB.Collection("articles")

	opts := options.Update().SetUpsert(true)
	filter := bson.M{"url": article.URL}
	update := bson.M{
		"$setOnInsert": bson.M{
			"createdAt": time.Now(),
		},
		"$set": bson.M{
			"title":     article.Title,
			"source":    article.Source,
			"score":     article.Score, // update score if we've scraped it again
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}
