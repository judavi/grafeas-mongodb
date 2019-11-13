package config

// DynamoDbConfig is the configuration for an AWS DynamoDB store.
type MongoDbConfig struct {
	Uri      string `mapstructure:"uri"`
	Database string `mapstructure:"database"`
}
