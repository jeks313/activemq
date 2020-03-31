package archive

import (
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Archives is a set of archive files, one per time period per key
type Archives struct {
	Path string // path to write to
	sync.Mutex
	maxBytes int
	archives map[string]*Archive
}

// New creates a new Archives et
func New(maxBytes int) *Archives {
	a := &Archives{}
	a.archives = make(map[string]*Archive)
	a.maxBytes = maxBytes
	return a
}

// CheckAndClose goes through the open archives and closes them if they are done with
func (a *Archives) CheckAndClose() {
	a.Lock()
	defer a.Unlock()
	for k, arch := range a.archives {
		if arch.NeedsRotation(0) {
			err := arch.Close()
			if err != nil {
				log.Error().Err(err).Str("key", k).Str("filename", arch.filename).Msg("failed to close archive")
				continue
			}
			delete(a.archives, k)
		}
	}
}

// Writes a document with a certain key
func (a *Archives) Write(topic, key string, doc []byte) error {
	k := fmt.Sprintf("%s/%s", topic, key)
	if a, ok := a.archives[k]; ok {
		return writeErr(a, doc)
	}
	arch := &Archive{topic: topic, key: key, maxBytes: a.maxBytes, path: a.Path}
	err := arch.Open()
	if err != nil {
		log.Error().Err(err).Str("filename", arch.filename).Msg("write: failed to open new archive file")
		return err
	}
	a.Lock()
	defer a.Unlock()
	a.archives[k] = arch
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
	sync.Mutex
	topic          string
	dateTimeFormat string
	template       string
	key            string
	filename       string
	out            *os.File
	path           string
	logger         zerolog.Logger
	writes         int64
	sizeBytes      int
	maxBytes       int
	index          int
}

// Open opens an archive file based on the current time and key
func (a *Archive) Open() error {
	a.Lock()
	defer a.Unlock()
	a.filename = a.formatFilename()
	a.logger = log.With().Str("filename", a.filename).Str("key", a.key).Logger()
	a.logger.Info().Msg("opening new archive")
	a.writes = 0
	a.sizeBytes = 0
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
	a.logger.Debug().Msg("opened new file")
	return nil
}

// Close closes and syncs a file
func (a *Archive) Close() error {
	a.Lock()
	defer a.Unlock()
	a.logger.Info().Int64("writes", a.writes).Msg("closing")
	err := a.out.Sync()
	if err != nil {
		log.Error().Err(err).Msg("failed to sync file on close")
		return err
	}
	a.writes = 0
	a.sizeBytes = 0
	return a.out.Close()
}

const template = "topic=<TOPIC>_dt=<DATETIME>_accountUID=<KEY>_part=<INDEX>.log"
const dateTimeFormat = "2006-01-02T15:00Z"

// Write writes a provided document to the archive, checks if filename needs rotation
func (a *Archive) Write(doc []byte) (int, error) {
	if a.NeedsRotation(len(doc) + 1) {
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
	a.Lock()
	defer a.Unlock()
	n, err := a.out.Write(doc)
	a.sizeBytes = a.sizeBytes + n
	return n, err
}

func (a *Archive) formatFilename() string {
	if a.template == "" {
		a.template = template
	}
	if a.dateTimeFormat == "" {
		a.dateTimeFormat = dateTimeFormat
	}
	// TODO: this is terrible, should make this a better formatting than string replacements
	datetime := time.Now().Format(a.dateTimeFormat)
	filename := strings.Replace(a.template, "<DATETIME>", datetime, -1)
	filename = strings.Replace(filename, "<TOPIC>", a.topic, -1)
	filename = strings.Replace(filename, "<KEY>", a.key, -1)
	filename = strings.Replace(filename, "<INDEX>", fmt.Sprintf("%02d", a.index), -1)
	return path.Join(a.path, filename)
}

// NeedsRotation checks the filename to see if we need to roll this file over
func (a *Archive) NeedsRotation(currentWriteSize int) bool {
	a.Lock()
	defer a.Unlock()
	filename := a.formatFilename() // picks up time rotation
	if filename != a.filename {
		a.index = 0
		return true
	}
	if a.sizeBytes+currentWriteSize > a.maxBytes { // size rotation so increment index
		a.index = a.index + 1
		return true
	}
	return false
}
