package viper

import (
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spf13/viper/internal/testutil"
)

func Test_searchInPath(t *testing.T) {
	t.Run("find file without extension", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config"))
		require.NoError(t, err)

		_, err = file.WriteString(`key: value`)
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigType("yaml")

		err = v.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "value", v.Get("key"))
	})

	t.Run("cannot find file with extension", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigType("yaml")

		err = v.ReadInConfig()
		require.EqualError(t, err, ConfigFileNotFoundError{v.configName, fmt.Sprintf("%s", v.configPaths)}.Error())
	})

	t.Run("find file with supported extension", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		_, err = file.WriteString(`key: value`)
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigType("yaml")

		err = v.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "value", v.Get("key"))
	})

	t.Run("find file with specific extension from multiple supported files", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		jsonFile, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.json"))
		require.NoError(t, err)

		jsonFile.Close()

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		_, err = file.WriteString(`key: value`)
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigType("yaml")

		err = v.ReadInConfig()
		require.NoError(t, err)

		assert.Equal(t, "value", v.Get("key"))
	})

	t.Run("find file with unsupported extension", func(t *testing.T) {
		fs := afero.NewMemMapFs()

		err := fs.Mkdir(testutil.AbsFilePath(t, "/etc/viper"), 0o777)
		require.NoError(t, err)

		jsonFile, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.json"))
		require.NoError(t, err)

		jsonFile.Close()

		file, err := fs.Create(testutil.AbsFilePath(t, "/etc/viper/config.yaml"))
		require.NoError(t, err)

		_, err = file.WriteString(`key: value`)
		require.NoError(t, err)

		file.Close()

		v := New()

		v.SetFs(fs)
		v.AddConfigPath("/etc/viper")
		v.SetConfigType("xyz")

		err = v.ReadInConfig()
		require.EqualError(t, err, UnsupportedConfigError("xyz").Error())
	})
}
