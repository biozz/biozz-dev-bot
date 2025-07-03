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
	mongoClient *mongo.Client
	mongoUserID string
}

func New(mongoURI string, mongoUserID string) (*LibreChat, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}
	return &LibreChat{
		mongoClient: client,
		mongoUserID: mongoUserID,
	}, nil
}

func (c *LibreChat) Cleanup() error {
	return c.mongoClient.Disconnect(context.TODO())
}

/*
Here is an example data from LibreChat database

	{
	  "_id": {
	    "$oid": "67d5597ddfb166b55d3af9a5"
	  },
	  "conversationId": "5f3b328c-dfd8-4bf1-8033-9e03117e94db",
	  "user": "67d54d915f5f1dfc7994e482",
	  "__v": 0,
	  "_meiliIndex": true,
	  "createdAt": {
	    "$date": "2025-02-01T19:44:31.478Z"
	  },
	  "endpoint": "openAI",
	  "files": [],
	  "isArchived": false,
	  "messages": [],
	  "model": "gpt-4o-mini",
	  "tags": [],
	  "title": "Конвертация таблицы в формат",
	  "updatedAt": {
	    "$date": "2025-02-01T19:44:31.478Z"
	  }
	}
*/
func (c *LibreChat) MongoCreateConversation(user string, endpoint string, model string) (string, error) {
	conversationID := uuid.New().String()
	now := time.Now()
	
	conversation := bson.M{
		"conversationId": conversationID,
		"user":          user,
		"__v":           0,
		"_meiliIndex":   true,
		"createdAt":     now,
		"endpoint":      endpoint,
		"files":         []string{},
		"isArchived":    false,
		"messages":      []string{},
		"model":         model,
		"tags":          []string{},
		"title":         "New Chat",
		"updatedAt":     now,
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

/*
 * {
   "_id": {
     "$oid": "67d5597ddfb166b55d3af9d6"
   },
   "messageId": "e58ea3eb-85f8-4f21-8b5a-b18768287192",
   "__v": 0,
   "_meiliIndex": true,
   "conversationId": "5f3b328c-dfd8-4bf1-8033-9e03117e94db",
   "endpoint": "openAI",
   "error": false,
   "isCreatedByUser": true,
   "model": "gpt-4o-mini",
   "parentMessageId": "00000000-0000-0000-0000-000000000000",
   "sender": "GPT-4",
   "text": "...",
   "unfinished": false,
   "user": "67d54d915f5f1dfc7994e482"
 }
*/

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
	
	sender := "GPT-4"
	if isCreatedByUser {
		sender = "User"
	}
	
	message := bson.M{
		"messageId":        messageID,
		"__v":              0,
		"_meiliIndex":      true,
		"conversationId":   convo,
		"endpoint":         conversation.Endpoint,
		"error":            false,
		"isCreatedByUser":  isCreatedByUser,
		"model":            conversation.Model,
		"parentMessageId":  parentMessageID,
		"sender":           sender,
		"text":             text,
		"unfinished":       false,
		"user":             conversation.User,
	}

	collection := c.mongoClient.Database("LibreChat").Collection("messages")
	_, err = collection.InsertOne(context.TODO(), message)
	if err != nil {
		return "", err
	}

	return messageID, nil
}
