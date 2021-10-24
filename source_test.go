//go:generate go run github.com/golang/mock/mockgen -package migrationsql -destination fs_mock_test.go io/fs DirEntry,ReadDirFS,File

package migrationsql

import (
	"errors"
	"io"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_parseSQLFile(t *testing.T) {
	t.Run("should match a file with type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dirEntry := NewMockDirEntry(ctrl)

		dirEntry.EXPECT().IsDir().Return(false)
		wantID := "123"
		wantDescription := "description with new data"
		wantType := "down"
		givenName := "123_description_with_new_data.down.sql"
		dirEntry.EXPECT().Name().Return(givenName)

		gotID, gotDescription, gotType := parseSQLFile(dirEntry)
		assert.Equal(t, gotID, wantID)
		assert.Equal(t, gotDescription, wantDescription)
		assert.Equal(t, gotType, wantType)
	})

	t.Run("should match a file without type", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dirEntry := NewMockDirEntry(ctrl)

		dirEntry.EXPECT().IsDir().Return(false)
		wantID := "123"
		wantDescription := "description with new data"
		wantType := ""
		givenName := "123_description_with_new_data.sql"
		dirEntry.EXPECT().Name().Return(givenName)

		gotID, gotDescription, gotType := parseSQLFile(dirEntry)
		assert.Equal(t, gotID, wantID)
		assert.Equal(t, gotDescription, wantDescription)
		assert.Equal(t, gotType, wantType)
	})

	t.Run("should not match when the entry is a directory", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dirEntry := NewMockDirEntry(ctrl)

		dirEntry.EXPECT().IsDir().Return(true)

		gotID, _, _ := parseSQLFile(dirEntry)
		assert.Empty(t, gotID)
	})

	t.Run("should not match file name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		dirEntry := NewMockDirEntry(ctrl)

		dirEntry.EXPECT().IsDir().Return(false)
		dirEntry.EXPECT().Name().Return("crazy_file_name.txt")

		gotID, _, _ := parseSQLFile(dirEntry)
		assert.Empty(t, gotID)
	})
}

func Test_loadMigrationFile(t *testing.T) {
	t.Run("should load the migration content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		wantFile := "random file"
		fs := NewMockReadDirFS(ctrl)
		f := NewMockFile(ctrl)
		wantContent := "migration content"

		fs.EXPECT().Open(wantFile).Return(f, nil)
		f.EXPECT().Read(gomock.Any()).DoAndReturn(func(d []byte) (int, error) {
			copy(d, wantContent)
			return len(wantContent), io.EOF
		})
		f.EXPECT().Close().Return(nil)

		gotContent, err := loadMigrationFile(fs, wantFile)
		assert.NoError(t, err)
		assert.Equal(t, wantContent, gotContent)
	})

	t.Run("should return empty when migration name is empty", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fs := NewMockReadDirFS(ctrl)

		content, err := loadMigrationFile(fs, "")
		assert.NoError(t, err)
		assert.Empty(t, content)
	})

	t.Run("should fail when opening migration fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fs := NewMockReadDirFS(ctrl)
		wantErr := errors.New("random error")

		fs.EXPECT().Open(gomock.Any()).Return(nil, wantErr)

		_, err := loadMigrationFile(fs, "random file")
		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("should fail reading the migration content", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		fs := NewMockReadDirFS(ctrl)
		f := NewMockFile(ctrl)
		wantErr := errors.New("random error")

		fs.EXPECT().Open(gomock.Any()).Return(f, nil)
		f.EXPECT().Read(gomock.Any()).Return(0, wantErr)
		f.EXPECT().Close().Return(nil)

		_, err := loadMigrationFile(fs, "random file")
		assert.ErrorIs(t, err, wantErr)
	})
}

func Test_newSourceSQLFromFiles(t *testing.T) {
	// t.Run("should fail when getting the ")
}