package store

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

const defaultFolderName = "srivastavcodes"

func CASPathTransformFunc(key string) PathKey {
	hash := sha256.Sum256([]byte(key))

	hashStr := hex.EncodeToString(hash[:])
	blocksize := 5
	sliceLen := len(hashStr) / blocksize

	paths := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		from, to := i*blocksize, (i*blocksize)+blocksize
		paths[i] = hashStr[from:to]
	}
	return PathKey{
		PathName: strings.Join(paths, "/"),
		FileName: hashStr,
	}
}

func DefaultPathTransformFn(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}

type PathTransformFn func(string) PathKey

type PathKey struct {
	PathName string
	FileName string
}

func (pk *PathKey) FullPath() string {
	return filepath.Join(pk.PathName, pk.FileName)
}

type StoreOpts struct {
	// Root is the root-folder-name of the folder/file structure.
	Root            string
	PathTransformFn PathTransformFn
}

type Store struct {
	log  zerolog.Logger
	opts StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	writer := zerolog.NewConsoleWriter()
	logger := zerolog.New(writer).With().Timestamp().Logger()

	if opts.PathTransformFn == nil {
		opts.PathTransformFn = DefaultPathTransformFn
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultFolderName
	}
	return &Store{
		opts: opts,
		log:  logger,
	}
}

// readStream looks for the file for (key) and reads the file contents
// into a buffer to read from. Returns a reader to read from and
// errs in case of PathError
func (s *Store) readStream(key string) (io.Reader, error) {
	pathkey := s.opts.PathTransformFn(key)

	pathToFile := filepath.Join(s.opts.Root, pathkey.FullPath())
	file, err := os.Open(pathToFile)

	buf := new(bytes.Buffer)

	n, err := io.Copy(buf, file)
	defer func() {
		s.log.Info().Msgf("read (%d) bytes from [%s]", n, file.Name())
	}()
	return buf, err
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathname := s.opts.PathTransformFn(key)

	rootPathName := filepath.Join(s.opts.Root, pathname.PathName)
	if err := os.MkdirAll(rootPathName, 0755); err != nil {
		return err
	}
	rootFilePath := filepath.Join(s.opts.Root, pathname.FullPath())

	file, err := os.OpenFile(rootFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("invalid path: %w\n", err)
	}
	defer file.Close()

	n, err := io.Copy(file, r)
	if err != nil {
		return fmt.Errorf("error while reading data: %w\n", err)
	}
	s.log.Info().Msgf("written (%d) bytes to disk: %s", n, rootFilePath)
	return nil
}

func (s *Store) Exists(key string) bool {
	pathkey := s.opts.PathTransformFn(key)

	pathToFile := filepath.Join(s.opts.Root, pathkey.FullPath())
	_, err := os.Stat(pathToFile)

	return !errors.Is(err, fs.ErrNotExist)
}

func (s *Store) DeleteHead(key string) error {
	pathkey := s.opts.PathTransformFn(key)

	pathToFile := filepath.Join(s.opts.Root, pathkey.FullPath())
	paths := strings.Split(pathToFile, "/")

	if len(paths) == 0 {
		return fmt.Errorf("path [%s] does not exists", paths)
	}
	err := os.RemoveAll(filepath.Join(paths[0], paths[1]))
	if err != nil {
		return fmt.Errorf("contents at %s could not be deleted", paths[1])
	}
	s.log.Info().Msgf("deleted [%s] from disk", pathToFile)
	return nil
}
