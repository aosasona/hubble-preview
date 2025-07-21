package queue

import (
	"github.com/golang-queue/queue"
	"github.com/rs/zerolog/log"
)

type logger struct{}

// Error implements queue.Logger.
func (l *logger) Error(args ...any) {
	log.Error().Str("source", "queue").Msgf("%v", args)
}

// Errorf implements queue.Logger.
func (l *logger) Errorf(format string, args ...any) {
	log.Error().Str("source", "queue").Msgf(format, args...)
}

// Fatal implements queue.Logger.
func (l *logger) Fatal(args ...any) {
	log.Fatal().Str("source", "queue").Msgf("%v", args)
}

// Fatalf implements queue.Logger.
func (l *logger) Fatalf(format string, args ...any) {
	log.Fatal().Str("source", "queue").Msgf(format, args...)
}

// Info implements queue.Logger.
func (l *logger) Info(args ...any) {
	log.Info().Str("source", "queue").Msgf("%v", args)
}

// Infof implements queue.Logger.
func (l *logger) Infof(format string, args ...any) {
	log.Info().Str("source", "queue").Msgf(format, args...)
}

var _ queue.Logger = (*logger)(nil)
