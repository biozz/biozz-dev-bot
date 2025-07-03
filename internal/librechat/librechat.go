package librechat

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type LibreChat struct {
	mongoClient *mongo.Client
	mongoUserID string
}

func New(mongoURI string, mongoUserID string) (*LibreChat, error) {
	client, _ := mongo.Connect(options.Client().ApplyURI(mongoURI))
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
	return "", nil
}

type Conversation struct {
	User     string `bson:"user,omitempty"`
	ID       string `bson:"conversationId,omitempty"`
	Endpoint string `bson:"endpoint,omitempty"`
	Model    string `bson:"model,omitempty"`
}

func (c *LibreChat) MongoGetConversation(convo string) (*Conversation, error) {
	return nil, nil
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
	return []Message{}, nil
}

func (c *LibreChat) MongoCreateMessage(convo string, text string, parentMessageID string, isCreatedByUser bool) (string, error) {
	return "", nil
}
