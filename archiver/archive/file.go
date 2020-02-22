package archive

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Archives is a set of archive files, one per time period per key
type Archives struct {
	keys map[string]*Archive
}

// New creates a new Archives et
func New() *Archives {
	a := &Archives{}
	a.keys = make(map[string]*Archive)
	return a
}

// Writes a document with a certain key
func (a *Archives) Write(topic, key string, doc []byte) error {
	k := fmt.Sprintf("%s/%s", topic, key)
	if a, ok := a.keys[k]; ok {
		return writeErr(a, doc)
	}
	arch := &Archive{topic: topic, key: key}
	err := arch.Open()
	if err != nil {
		log.Error().Err(err).Msgf("write: failed to open new archive file: %v", arch)
		return err
	}
	a.keys[k] = arch
	return writeErr(arch, doc)
}

func writeErr(arch *Archive, doc []byte) error {
	doc = append(doc, []byte("\n")...)
	n, err := arch.Write(doc)
	if err != nil {
		log.Error().Err(err).Msgf("write: failed to write document to archive: %v", arch)
		return err
	}
	log.Debug().Str("key", arch.key).Str("filename", arch.filename).Int("size", n).Msg("write: wrote document")
	return nil
}

// Archive is a single file representing a timeslice of data for a particular key
type Archive struct {
	topic          string
	dateTimeFormat string
	template       string
	key            string
	filename       string
	out            *os.File
	logger         zerolog.Logger
	writes         int64
}

// Open opens an archive file based on the current time and key
func (a *Archive) Open() error {
	a.filename = a.formatFilename()
	a.logger = log.With().Str("filename", a.filename).Str("key", a.key).Logger()
	a.logger.Info().Msg("opening new archive")
	a.writes = 0
	openFlag := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	f, err := os.OpenFile(a.filename, openFlag, 0644)
	if os.IsExist(err) {
		log.Info().Err(err).Msg("file already exists")
	}
	if !os.IsExist(err) && err != nil {
		log.Error().Err(err).Msg("unable to create or append file")
		return err
	}
	a.out = f
	return nil
}

// Close closes and syncs a file
func (a *Archive) Close() error {
	a.logger.Info().Int64("writes", a.writes).Msg("closing")
	err := a.out.Sync()
	if err != nil {
		log.Error().Err(err).Msg("failed to sync file on close")
		return err
	}
	return a.out.Close()
}

const template = "archive.<TOPIC>.<DATETIME>.<KEY>.log"
const dateTimeFormat = "2006-01-02_15"

// Write writes a provided document to the archive, checks if filename needs rotation
func (a *Archive) Write(doc []byte) (int, error) {
	if a.needsRotation() {
		err := a.Close()
		if err != nil {
			a.logger.Error().Err(err).Str("filename", a.filename).Msg("failed to close file")
			return 0, err
		}
		err = a.Open()
		if err != nil {
			a.logger.Error().Err(err).Str("filename", a.filename).Msg("failed to open file")
			return 0, err
		}
	}
	a.writes = a.writes + 1
	if a.writes%100 == 0 { // log every 100 writes so we have some activity logging TODO: change this to a tracer?
		a.logger.Info().Int64("writes", a.writes).Msg("document writes")
	}
	return a.out.Write(doc)
}

func (a *Archive) formatFilename() string {
	if a.template == "" {
		a.template = template
	}
	if a.dateTimeFormat == "" {
		a.dateTimeFormat = dateTimeFormat
	}
	datetime := time.Now().Format(a.dateTimeFormat)
	filename := strings.Replace(a.template, "<DATETIME>", datetime, -1)
	filename = strings.Replace(filename, "<TOPIC>", a.topic, -1)
	filename = strings.Replace(filename, "<KEY>", a.key, -1)
	return filename
}

// needsRotating checks the filename to see if we need to roll this file over
func (a *Archive) needsRotation() bool {
	filename := a.formatFilename()
	if filename != a.filename {
		return true
	}
	return false
}
