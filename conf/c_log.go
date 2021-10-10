package conf

import (
	"github.com/caarlos0/env"
	"github.com/sirupsen/logrus"
)

type LOG struct {
	LogLevel   string `env:"LOG_LEVEL" envDefault:"debug"`
	LogLines   bool   `env:"LOG_LINES" envDefault:"false"`
	LogJson    bool   `env:"LOG_JSON" envDefault:"false"`
	TimeFormat string `env:"LOG_TIME_FORMAT" envDefault:"2006-01-02T15:04:05Z"`
}

func (c *config) LOG() *logrus.Entry {
	if c.log != nil {
		return c.log
	}

	settings := &LOG{}

	if err := env.Parse(settings); err != nil {
		c.LOG().WithError(err).Fatal("parsing LOG configuration")
	}

	level, _ := logrus.ParseLevel(settings.LogLevel)

	logger := logrus.New()
	if settings.LogJson {
		formatter := &logrus.JSONFormatter{}
		formatter.TimestampFormat = settings.TimeFormat
		logger.SetFormatter(formatter)
	} else {
		formatter := &logrus.TextFormatter{}
		formatter.TimestampFormat = settings.TimeFormat
		formatter.FullTimestamp = true
		formatter.ForceColors = true
		logger.SetFormatter(formatter)
	}
	logger.SetLevel(level)
	logger.SetReportCaller(settings.LogLines)

	c.log = logrus.NewEntry(logger)

	return c.log
}
