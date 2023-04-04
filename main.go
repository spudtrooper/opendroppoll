package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"

	"github.com/spudtrooper/goutil/check"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	verbose = flag.Bool("v", false, "Verbose logging")
)

func createDB(ctx context.Context) (*mongo.Database, error) {
	const port = 27017
	const dbName = "opendroppoll"

	uri := fmt.Sprintf("mongodb://localhost:%d", port)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)
	return db, nil
}

type notDiscoverable struct {
	EventName string `bson:"event_name"`
	ID        string `bson:"id"`
}

type foundIndex struct {
	EventName string `bson:"event_name"`
	ID        string `bson:"id"`
	Index     int    `bson:"index"`
	Name      string `bson:"name"`
}

func mustInsert(ctx context.Context, db *mongo.Database, event interface{}) {
	log.Printf("Inserting %+v", event)
	_, err := db.Collection("events").InsertOne(ctx, event)
	check.Err(err)
}

func realMain(ctx context.Context) {
	db, err := createDB(ctx)
	check.Err(err)

	cmd := exec.Command("opendrop", "-d", "find")

	stderr, _ := cmd.StderrPipe()
	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)

	// Receiver ID xxxxx54defdf is not discoverable
	var idIsNotDiscoverable = regexp.MustCompile(`Receiver ID ([0-9a-f]{12}) is not discoverable`)
	// Found  index 1  ID xxxxx54defdf  name Boooooo
	var foundIndexRE = regexp.MustCompile(`Found\s+index\s+(\d+)\s+ID\s+([0-9a-f]{12})\s+name\s+(.*)`)

	for scanner.Scan() {
		t := scanner.Text()
		if *verbose {
			log.Println(t)
		}
		if m := idIsNotDiscoverable.FindStringSubmatch(t); len(m) == 2 {
			id := m[1]
			evt := notDiscoverable{
				EventName: "not_discoverable",
				ID:        id,
			}
			mustInsert(ctx, db, evt)
			continue
		}
		if m := foundIndexRE.FindStringSubmatch(t); len(m) == 4 {
			indexStr, id, name := m[1], m[2], m[3]
			index, err := strconv.Atoi(indexStr)
			check.Err(err)
			evt := foundIndex{
				EventName: "found_index",
				ID:        id,
				Index:     index,
				Name:      name,
			}
			mustInsert(ctx, db, evt)
			continue
		}

	}
	cmd.Wait()
}

func main() {
	flag.Parse()
	realMain(context.Background())
}
