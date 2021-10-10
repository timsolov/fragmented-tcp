package conf

import (
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config interface {
	LOG() *logrus.Entry
}

type config struct {
	sync.Mutex

	log *logrus.Entry
}

func LoadDotEnv(stepsUp int) error {
	path := strings.Repeat("../", stepsUp) + ".env"
	if err := godotenv.Load(path); err != nil {
		return err
	}
	return nil
}

func New() Config {

	LoadDotEnv(0)

	c := &config{}

	return c
}
