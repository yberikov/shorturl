package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth/internal/domain/models"
	"auth/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Storage struct {
	client            *mongo.Client
	collection        *mongo.Collection
	counterCollection *mongo.Collection
}

type UserDocument struct {
	ID       int64  `bson:"_id"`
	Email    string `bson:"email"`
	Password string `bson:"password"`
}

type Counter struct {
	ID  string `bson:"_id"`
	Seq int64  `bson:"seq"`
}

func New(uri, database, collection string) (*Storage, error) {

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf(" create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := client.Database(database)
	coll := db.Collection(collection)
	counterColl := db.Collection("counters") // Initialize the counter collection

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("create index: %w", err)
	}

	return &Storage{client: client, collection: coll, counterCollection: counterColl}, nil
}

func (s *Storage) getNextSequence(ctx context.Context, name string) (int64, error) {
	filter := bson.M{"_id": name}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

	var counter Counter
	err := s.counterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("find and update sequence: %w", err)
	}

	return counter.Seq, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passHash []byte) (int64, error) {

	userID, err := s.getNextSequence(ctx, "userid")

	if err != nil {
		return 0, err
	}

	doc := UserDocument{
		ID:       userID,
		Email:    email,
		Password: string(passHash),
	}

	_, err = s.collection.InsertOne(ctx, doc)
	if err != nil {
		var mongoWriteException mongo.WriteException
		if errors.As(err, &mongoWriteException) {
			for _, writeError := range mongoWriteException.WriteErrors {
				if writeError.Code == 11000 {
					return 0, fmt.Errorf("%w", storage.ErrUserExists)
				}
			}
		}
		return 0, fmt.Errorf("insert document: %w", err)
	}

	return userID, nil
}

func (s *Storage) GetUser(ctx context.Context, email string) (models.User, error) {

	var doc UserDocument
	filter := bson.D{{Key: "email", Value: email}}

	err := s.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("find document: %w", err)
	}

	return models.User{
		ID:       doc.ID,
		Email:    doc.Email,
		PassHash: []byte(doc.Password),
	}, nil
}
