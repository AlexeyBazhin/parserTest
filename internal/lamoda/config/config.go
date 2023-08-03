package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	CredS3ID     string `envconfig:"CRED_S3_ID" required:"true"`
	CredS3Secret string `envconfig:"CRED_S3_SECRET" required:"true"`
	Endpoint     string `envconfig:"ENDPOINT" required:"true"`
	PrivatePort  string `envconfig:"PRIVATE_PORT" required:"true"`
}

func LoadConfig() *Config {
	cfg := Config{}

	// ex, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// exPath := filepath.Dir(ex)
	// fmt.Println(exPath)

	for _, fileName := range []string{".env.local", "../../.env"} { //.env.local for local secrets (higher priority than .env)
		err := godotenv.Load(fileName) // in cycle cause first error in varargs prevents loading next files
		if err != nil {
			log.Println("[LAMODA][CONFIG] ERROR: ", err)
		}
	}

	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalln(err)
	}

	return &cfg
}
