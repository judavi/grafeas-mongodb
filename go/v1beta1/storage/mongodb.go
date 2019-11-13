package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/fernet/fernet-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	grafeasConfig "github.com/grafeas/grafeas/go/config"
	"github.com/grafeas/grafeas/go/name"
	"github.com/grafeas/grafeas/go/v1beta1/storage"
	pb "github.com/grafeas/grafeas/proto/v1beta1/grafeas_go_proto"
	prpb "github.com/grafeas/grafeas/proto/v1beta1/project_go_proto"
	"github.com/judavi/grafeas-mongodb/go/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	fieldmaskpb "google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Occurrence struct {
	project_name    string
	occurrence_name string
	note_id         string
	data            string
}

type MongoDb struct {
	*mongo.Database
	PaginationKey string
}

func MongodbStorageTypeProvider(storageType string, storageConfig *grafeasConfig.StorageConfiguration) (*storage.Storage, error) {
	if storageType != "mongodb" {
		return nil, errors.New(fmt.Sprintf("Unknown storage type %s, must be 'mongodb'", storageType))
	}

	var storeConfig config.MongoDbConfig

	err := grafeasConfig.ConvertGenericConfigToSpecificType(storageConfig, &storeConfig)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to create MongoDbConfig, %s", err))
	}

	s := NewMongoDbStore(&storeConfig)
	storage := &storage.Storage{
		Ps: s,
		Gs: s,
	}

	return storage, nil
}

func NewMongoDbStore(config *config.MongoDbConfig) *MongoDb {

	// Set client options
	clientOptions := options.Client().ApplyURI(config.Uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("Database server is not alive")
	}

	// Get a handle for your collection
	database := client.Database(config.Database)

	return &MongoDb{
		Database:      database,
		PaginationKey: "",
	}

}

// CreateProject adds the specified project to the store
func (pg *MongoDb) CreateProject(ctx context.Context, pID string, p *prpb.Project) (*prpb.Project, error) {
	pName := name.FormatProject(pID)
	project := bson.D{primitive.E{Key: "name", Value: pName}}
	_, err := pg.Collection("projects").InsertOne(context.TODO(), project)
	if err != nil {
		log.Println("Failed to insert Project in database", err)
		return nil, status.Error(codes.Internal, "Failed to insert Project in database")
	}

	return p, nil
}

// DeleteProject deletes the project with the given pID from the store
func (pg *MongoDb) DeleteProject(ctx context.Context, pID string) error {
	pName := name.FormatProject(pID)
	_, err := pg.Collection("Project").DeleteOne(context.TODO(), bson.M{"name": pName})
	if err != nil {
		return status.Error(codes.Internal, "Failed to delete Project from database")
	}

	return nil
}

// GetProject returns the project with the given pID from the store
func (pg *MongoDb) GetProject(ctx context.Context, pID string) (*prpb.Project, error) {
	pName := name.FormatProject(pID)
	count, err := pg.Collection("Project").CountDocuments(context.TODO(), bson.M{"name": pName})
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to query Project from database")
	}
	if count > 0 {
		return nil, status.Errorf(codes.NotFound, "Project with name %q does not Exist", pName)
	}
	return &prpb.Project{Name: pName}, nil
}

// ListProjects returns up to pageSize number of projects beginning at pageToken (or from
// start if pageToken is the empty string).
func (pg *MongoDb) ListProjects(ctx context.Context, filter string, pageSize int, pageToken string) ([]*prpb.Project, string, error) {
	//id := decryptInt64(pageToken, pg.PaginationKey, 0)
	//TODO
	return nil, "", nil
}

// CreateNote adds the specified note
func (pg *MongoDb) CreateNote(ctx context.Context, pID, nID, uID string, n *pb.Note) (*pb.Note, error) {
	return nil, nil
}

// BatchCreateNotes batch creates the specified notes in memstore.
func (pg *MongoDb) BatchCreateNotes(ctx context.Context, pID, uID string, notes map[string]*pb.Note) ([]*pb.Note, []error) {
	return nil, nil
}

// DeleteNote deletes the note with the given pID and nID
func (pg *MongoDb) DeleteNote(ctx context.Context, pID, nID string) error {
	return nil
}

// UpdateNote updates the existing note with the given pID and nID
func (pg *MongoDb) UpdateNote(ctx context.Context, pID, nID string, n *pb.Note, mask *fieldmaskpb.FieldMask) (*pb.Note, error) {
	return nil, nil
}

// GetNote returns the note with project (pID) and note ID (nID)
func (pg *MongoDb) GetNote(ctx context.Context, pID, nID string) (*pb.Note, error) {
	return nil, nil
}

// CreateOccurrence adds the specified occurrence
func (pg *MongoDb) CreateOccurrence(ctx context.Context, pID, uID string, o *pb.Occurrence) (*pb.Occurrence, error) {
	o = proto.Clone(o).(*pb.Occurrence)
	o.CreateTime = ptypes.TimestampNow()
	var id string
	if nr, err := uuid.NewRandom(); err != nil {
		return nil, status.Error(codes.Internal, "Failed to generate UUID")
	} else {
		id = nr.String()
	}
	o.Name = fmt.Sprintf("projects/%s/occurrences/%s", pID, id)

	m := jsonpb.Marshaler{}
	jsonObject, err := m.MarshalToString(o)
	if err != nil {
		log.Println("Unable to marshal occurrence into json", err)
		return nil, status.Error(codes.Internal, "Unable to marshal occurrence into json")
	}
	log.Println("HELLOOO")
	log.Println(jsonObject)

	nPID, nID, err := name.ParseNote(o.NoteName)
	if err != nil {
		log.Printf("Invalid note name: %v", o.NoteName)
		return nil, status.Error(codes.InvalidArgument, "Invalid note name")
	}
	//occurrence := Occurrence{pID, nPID, nID, jsonObject}

	var bdoc interface{}
	err = bson.UnmarshalExtJSON([]byte(jsonObject), true, &bdoc)
	if err != nil {
		panic(err)
	}
	_, err = pg.Collection("occurrences").InsertOne(context.Background(),
		bson.D{
			{"project_name", pID},
			{"occurrence_name", nPID},
			{"note_id", nID},
			{"data", &bdoc}})

	if err != nil {
		log.Println("Failed to insert Occurrence in database", err)
		return nil, status.Error(codes.Internal, "Failed to insert Occurrence in database")
	}

	return o, nil
}

// BatchCreateOccurrence batch creates the specified occurrences in PostreSQL.
func (pg *MongoDb) BatchCreateOccurrences(ctx context.Context, pID string, uID string, occs []*pb.Occurrence) ([]*pb.Occurrence, []error) {
	return nil, nil
}

// DeleteOccurrence deletes the occurrence with the given pID and oID
func (pg *MongoDb) DeleteOccurrence(ctx context.Context, pID, oID string) error {
	return nil
}

// UpdateOccurrence updates the existing occurrence with the given projectID and occurrenceID
func (pg *MongoDb) UpdateOccurrence(ctx context.Context, pID, oID string, o *pb.Occurrence, mask *fieldmaskpb.FieldMask) (*pb.Occurrence, error) {
	return nil, nil
}

// GetOccurrence returns the occurrence with pID and oID
func (pg *MongoDb) GetOccurrence(ctx context.Context, pID, oID string) (*pb.Occurrence, error) {
	return nil, nil
}

// ListOccurrences returns up to pageSize number of occurrences for this project beginning
// at pageToken, or from start if pageToken is the empty string.
func (pg *MongoDb) ListOccurrences(ctx context.Context, pID, filter, pageToken string, pageSize int32) ([]*pb.Occurrence, string, error) {
	return nil, "", nil
}

// GetOccurrenceNote gets the note for the specified occurrence from PostgreSQL.
func (pg *MongoDb) GetOccurrenceNote(ctx context.Context, pID, oID string) (*pb.Note, error) {
	return nil, nil
}

// ListNotes returns up to pageSize number of notes for this project (pID) beginning
// at pageToken (or from start if pageToken is the empty string).
func (pg *MongoDb) ListNotes(ctx context.Context, pID, filter, pageToken string, pageSize int32) ([]*pb.Note, string, error) {
	return nil, "", nil
}

// ListNoteOccurrences returns up to pageSize number of occcurrences on the particular note (nID)
// for this project (pID) projects beginning at pageToken (or from start if pageToken is the empty string).
func (pg *MongoDb) ListNoteOccurrences(ctx context.Context, pID, nID, filter, pageToken string, pageSize int32) ([]*pb.Occurrence, string, error) {
	return nil, "", nil
}

// GetVulnerabilityOccurrencesSummary gets a summary of vulnerability occurrences from storage.
func (pg *MongoDb) GetVulnerabilityOccurrencesSummary(ctx context.Context, projectID, filter string) (*pb.VulnerabilityOccurrencesSummary, error) {
	return &pb.VulnerabilityOccurrencesSummary{}, nil
}

// Encrypt int64 using provided key
func encryptInt64(v int64, key string) (string, error) {
	k, err := fernet.DecodeKey(key)
	if err != nil {
		return "", err
	}
	bytes, err := fernet.EncryptAndSign([]byte(strconv.FormatInt(v, 10)), k)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Decrypts encrypted int64 using provided key. Returns defaultValue if decryption fails.
func decryptInt64(encrypted string, key string, defaultValue int64) int64 {
	k, err := fernet.DecodeKey(key)
	if err != nil {
		return defaultValue
	}
	bytes := fernet.VerifyAndDecrypt([]byte(encrypted), time.Hour, []*fernet.Key{k})
	if bytes == nil {
		return defaultValue
	}
	decryptedValue, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		return defaultValue
	}
	return decryptedValue
}
