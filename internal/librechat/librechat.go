package librechat

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LibreChat struct {
	mongoClient  *mongo.Client
	mongoUserID  string
	mongoTag     string
	convoModel   string
	summaryModel string
}

const (
	EndpointOpenAI = "openAI"
)

type NewParams struct {
	MongoURI     string
	MongoUserID  string
	MongoTag     string
	ConvoModel   string
	SummaryModel string
}

func New(params NewParams) (*LibreChat, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(params.MongoURI))
	if err != nil {
		return nil, err
	}

	lc := &LibreChat{
		mongoClient:  client,
		mongoUserID:  params.MongoUserID,
		mongoTag:     params.MongoTag,
		convoModel:   params.ConvoModel,
		summaryModel: params.SummaryModel,
	}

	// Check if tag exists, create if not
	err = lc.ensureTag()
	if err != nil {
		return nil, err
	}

	return lc, nil
}

func (c *LibreChat) Cleanup() error {
	return c.mongoClient.Disconnect(context.TODO())
}

func (c *LibreChat) MongoCreateConversation(endpoint string) (string, error) {
	conversationID := uuid.New().String()
	now := time.Now()

	conversation := bson.M{
		"conversationId": conversationID,
		"user":           c.mongoUserID,
		"__v":            0,
		"_meiliIndex":    true,
		"createdAt":      now,
		"endpoint":       endpoint,
		"files":          []string{},
		"isArchived":     false,
		"messages":       []string{},
		"model":          c.convoModel,
		"tags":           []string{c.mongoTag},
		"title":          "New Chat",
		"updatedAt":      now,
	}

	collection := c.mongoClient.Database("LibreChat").Collection("conversations")
	_, err := collection.InsertOne(context.TODO(), conversation)
	if err != nil {
		return "", err
	}

	return conversationID, nil
}

type Conversation struct {
	User     string `bson:"user,omitempty"`
	ID       string `bson:"conversationId,omitempty"`
	Endpoint string `bson:"endpoint,omitempty"`
	Model    string `bson:"model,omitempty"`
}

func (c *LibreChat) MongoGetConversation(convo string) (*Conversation, error) {
	collection := c.mongoClient.Database("LibreChat").Collection("conversations")

	var conversation Conversation
	filter := bson.M{"conversationId": convo}
	err := collection.FindOne(context.TODO(), filter).Decode(&conversation)
	if err != nil {
		return nil, err
	}

	return &conversation, nil
}

const DefaultParentMessageID = "00000000-0000-0000-0000-000000000000"

type Message struct {
	ID              string `bson:"messageId,omitempty"`
	Text            string `bson:"text,omitempty"`
	IsCreatedByUser bool   `bson:"isCreatedByUser,omitempty"`
	ParentMessageID string `bson:"parentMessageId,omitempty"`
}

func (c *LibreChat) MongoGetConversationMessages(convo string) ([]Message, error) {
	collection := c.mongoClient.Database("LibreChat").Collection("messages")

	filter := bson.M{"conversationId": convo}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var messages []Message
	for cursor.Next(context.TODO()) {
		var message Message
		if err := cursor.Decode(&message); err != nil {
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (c *LibreChat) MongoCreateMessage(convo string, text string, parentMessageID string, isCreatedByUser bool) (string, error) {
	messageID := uuid.New().String()

	// Get conversation details to populate required fields
	conversation, err := c.MongoGetConversation(convo)
	if err != nil {
		return "", err
	}

	sender := c.convoModel
	if isCreatedByUser {
		sender = "User"
	}

	message := bson.M{
		"messageId":       messageID,
		"__v":             0,
		"_meiliIndex":     true,
		"conversationId":  convo,
		"endpoint":        conversation.Endpoint,
		"error":           false,
		"isCreatedByUser": isCreatedByUser,
		"model":           conversation.Model,
		"parentMessageId": parentMessageID,
		"sender":          sender,
		"text":            text,
		"unfinished":      false,
		"user":            conversation.User,
	}

	collection := c.mongoClient.Database("LibreChat").Collection("messages")
	_, err = collection.InsertOne(context.TODO(), message)
	if err != nil {
		return "", err
	}

	return messageID, nil
}

type Tag struct {
	User        string    `bson:"user,omitempty"`
	Tag         string    `bson:"tag,omitempty"`
	Count       int       `bson:"count,omitempty"`
	CreatedAt   time.Time `bson:"createdAt,omitempty"`
	Description string    `bson:"description,omitempty"`
	Position    int       `bson:"position,omitempty"`
	UpdatedAt   time.Time `bson:"updatedAt,omitempty"`
}

func (c *LibreChat) ensureTag() error {
	collection := c.mongoClient.Database("LibreChat").Collection("conversationtags")

	filter := bson.M{"user": c.mongoUserID, "tag": c.mongoTag}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return err
	}

	if count == 0 {
		_, err = c.MongoAddTag(c.mongoTag, "Auto-created tag for biozz dev bot")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *LibreChat) MongoGetTags() ([]string, error) {
	collection := c.mongoClient.Database("LibreChat").Collection("tags")

	filter := bson.M{"user": c.mongoUserID}
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var tags []string
	for cursor.Next(context.TODO()) {
		var tag Tag
		if err := cursor.Decode(&tag); err != nil {
			continue
		}
		tags = append(tags, tag.Tag)
	}

	return tags, nil
}

func (c *LibreChat) MongoAddTag(tagName string, description string) (string, error) {
	now := time.Now()

	// Get current max position
	collection := c.mongoClient.Database("LibreChat").Collection("conversationtags")
	filter := bson.M{"user": c.mongoUserID}
	opts := options.FindOne().SetSort(bson.M{"position": -1})

	var lastTag Tag
	err := collection.FindOne(context.TODO(), filter, opts).Decode(&lastTag)
	position := 1
	if err == nil {
		position = lastTag.Position + 1
	}

	tag := bson.M{
		"user":        c.mongoUserID,
		"tag":         tagName,
		"__v":         0,
		"count":       0,
		"createdAt":   now,
		"description": description,
		"position":    position,
		"updatedAt":   now,
	}

	result, err := collection.InsertOne(context.TODO(), tag)
	if err != nil {
		return "", err
	}

	return result.InsertedID.(bson.ObjectID).Hex(), nil
}

func (c *LibreChat) MongoUpdateConversationTitle(convoID string, title string) error {
	collection := c.mongoClient.Database("LibreChat").Collection("conversations")

	filter := bson.M{"conversationId": convoID}
	update := bson.M{
		"$set": bson.M{
			"title":     title,
			"updatedAt": time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.TODO(), filter, update)
	return err
}
