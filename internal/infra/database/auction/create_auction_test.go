package auction

import (
	"context"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	mongoURL := os.Getenv("MONGODB_URL")
	if mongoURL == "" {
		mongoURL = "mongodb://admin:admin@localhost:27017/auctions?authSource=admin"
	}

	os.Setenv("AUCTION_DURATION", "3s")

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("MongoDB not available, skipping integration test: %v", err)
	}

	database := client.Database("auctions_test")
	defer database.Drop(ctx)

	repo := NewAuctionRepository(database)

	auctionEntity, intErr := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"This is a test auction product description",
		auction_entity.New,
	)
	if intErr != nil {
		t.Fatalf("Failed to create auction entity: %v", intErr)
	}

	if intErr := repo.CreateAuction(ctx, auctionEntity); intErr != nil {
		t.Fatalf("Failed to insert auction: %v", intErr)
	}

	// Verify auction is Active right after creation
	var auctionMongo AuctionEntityMongo
	filter := bson.M{"_id": auctionEntity.Id}
	if err := repo.Collection.FindOne(ctx, filter).Decode(&auctionMongo); err != nil {
		t.Fatalf("Failed to find auction: %v", err)
	}

	if auctionMongo.Status != auction_entity.Active {
		t.Fatalf("Expected auction status to be Active (%d), got %d",
			auction_entity.Active, auctionMongo.Status)
	}

	// Wait for the auction duration + buffer
	time.Sleep(4 * time.Second)

	// Verify auction is now Completed
	var closedAuction AuctionEntityMongo
	if err := repo.Collection.FindOne(ctx, filter).Decode(&closedAuction); err != nil {
		t.Fatalf("Failed to find auction after close: %v", err)
	}

	if closedAuction.Status != auction_entity.Completed {
		t.Fatalf("Expected auction status to be Completed (%d), got %d",
			auction_entity.Completed, closedAuction.Status)
	}

	t.Log("Auction was automatically closed after the configured duration")
}
