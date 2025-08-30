package store

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathTransformFunc(t *testing.T) {
	key := "bestpicture"
	pathname := CASPathTransformFunc(key)
	originalExpected := "0c5f470056f2abe659c2a9508bcf371da0af411f3564b9c404bbc7189e749f3c"
	pathnameExpected := "0c5f4/70056/f2abe/659c2/a9508/bcf37/1da0a/f411f/3564b/9c404/bbc71/89e74"
	require.Equal(t, originalExpected, pathname.FileName)
	require.Equal(t, pathnameExpected, pathname.PathName)
}

func TestStore(t *testing.T) {
	opts := StoreOpts{
		PathTransformFn: CASPathTransformFunc,
	}
	store := NewStore(opts)

	key := "my special picture"
	data := []byte("some jpg bytes")

	err := store.writeStream(key, bytes.NewReader(data))
	require.NoError(t, err)

	require.True(t, store.Exists(key))

	r, err := store.readStream(key)
	require.NoError(t, err)

	b, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Equal(t, b, data)

	err = store.DeleteHead(key)
	require.NoError(t, err)
}
