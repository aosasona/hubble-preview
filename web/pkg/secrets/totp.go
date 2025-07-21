package secrets

import (
	"errors"

	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"golang.org/x/sync/errgroup"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(totp)
type Scope string

const ChunkSize = 8

var (
	ErrorUnknownScope   = errors.New("unknown scope")
	ErrorNotImplemented = errors.New("not implemented")
)

type Manager struct {
	config   *config.Config
	totpRepo repository.TOTPRepository

	debug bool
}

func NewManager(config *config.Config, repository repository.TOTPRepository) *Manager {
	return &Manager{config: config, totpRepo: repository}
}

func (m *Manager) EnableDebug()  { m.debug = true }
func (m *Manager) DisableDebug() { m.debug = false }

// HandleSecretRotation is a method to rotate secrets for each scope
func (m *Manager) HandleSecretRotation(scope Scope) error {
	log.Debug().Msgf("rotating secrets for scope: %s", scope)

	switch scope {
	case ScopeTotp:
		return m.handleTOTPSecretRotation()
	default:
		return ErrorUnknownScope
	}
}

/*
handleTOTPSecretRotation is a method to rotate TOTP secrets

This will retrieve all TOTP secrets using the older version of the secret and rotate them to the latest version.
*/
type chunkFailure struct {
	id  int32
	err error
}

func (m *Manager) handleTOTPSecretRotation() error {
	var (
		failedChunks []chunkFailure
		failedChan   = make(chan chunkFailure, ChunkSize)
	)

	// Start a goroutine to collect failed chunk IDs and their errors
	// We don't want to stop simply because one item in a chunk failed
	go func() {
		for {
			select {
			case id, ok := <-failedChan:
				if !ok {
					return
				}

				failedChunks = append(failedChunks, id)
			default:
				return
			}
		}
	}()

	// 1. Retrieve all TOTP secrets using the older version of the secret
	outdatedSecrets, err := m.totpRepo.FindOutdatedSecrets(m.config.LatestTotpKeyVersion())
	if err != nil {
		return err
	}

	if len(outdatedSecrets) == 0 {
		log.Info().Msg("no outdated TOTP secrets found")
		return nil
	}

	// 2. Chunk the outdated secrets
	chunks := [][]models.TotpSecret{outdatedSecrets}
	if len(outdatedSecrets) > ChunkSize {
		chunks = lib.Chunk(outdatedSecrets, ChunkSize)
	}

	// 3. Re-encrypt the TOTP secrets using the latest version of the secret
	eg := new(errgroup.Group)
	for i := range chunks {
		eg.Go(func() error {
			chunk := chunks[i]
			err := m.processTotpSecretsChunk(chunk, failedChan)
			return err
		})
	}

	// Close the failed channel
	close(failedChan)

	// Wait for all goroutines to finish
	if err := eg.Wait(); err != nil {
		log.Error().Msgf("failed to rotate TOTP secrets: %v", err)
	}

	// Log the failed chunks
	if len(failedChunks) > 0 {
		for _, failed := range failedChunks {
			log.Error().
				Err(failed.err).
				Int32("id", failed.id).
				Msg("failed to rotate TOTP secret")
		}
	}

	log.Info().Msg("TOTP secret rotation complete")

	return nil
}

func (m *Manager) processTotpSecretsChunk(
	chunk []models.TotpSecret,
	failedChan chan chunkFailure,
) error {
	chunksToSave := make([]*models.TotpSecret, 0, len(chunk))
	for i := range chunk {
		secret := &chunk[i]

		// Ensure we have a version less than the latest version
		if secret.Version == m.config.LatestTotpKeyVersion() {
			log.Info().
				Msgf("skipping secret %d, already using the latest version", secret.ID)

			continue
		}

		// Ensure we still have the key for the version
		key := m.config.GetTotpKey(secret.Version)
		if key == nil {
			failedChan <- chunkFailure{id: secret.ID, err: errors.New("key not found for version")}
			continue
		}

		// Decrypt the old hash with the matching version key
		plainTextSecret, err := lib.DecryptAES(lib.DecryptAesParams{
			Key:        *key,
			CipherText: secret.Hash,
		})
		if err != nil {
			failedChan <- chunkFailure{id: secret.ID, err: err}
			continue
		}

		// Re-encrypt the old hash with the latest version key
		newHash, err := lib.EncryptAES(lib.EncryptAesParams{
			Key:       m.config.GetLatestTotpKey(),
			PlainText: plainTextSecret,
		})
		if err != nil {
			failedChan <- chunkFailure{id: secret.ID, err: err}
			continue
		}

		// Update the TOTP secret with the new hash and version
		secret.Hash = newHash
		secret.Version = m.config.LatestTotpKeyVersion()

		log.Info().
			Int32("id", secret.ID).
			Msg("[pre-commit] rotated TOTP secret")

		chunksToSave = append(chunksToSave, secret)
	}

	// Free up the memory
	chunk = nil

	// 4. Update the TOTP secrets hashes and versions in the database
	if err := m.totpRepo.BatchUpdateSecrets(chunksToSave); err != nil {
		log.Error().Msgf("failed to update TOTP secrets: %v", err)
		return err
	}

	return nil
}
