package configor

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ExampleConfig struct {
	APPName string `default:"app name"`

	DB struct {
		Name     string
		User     string `default:"root"`
		Password string `required:"true" env:"DBPassword"`
		Port     uint   `default:"3306"`
	}

	Contacts []struct {
		Name  string
		Email string `required:"true"`
	}
}

const exampleConfigAllYAML = `
appname: test

db:
  name: test
  user: test
  password: test
  port: 1234

contacts:
  - name: i test
    email: test@test.com
`

const exampleConfigAllJSON = `
{
    "appname": "test",
    "db": {
        "name": "test",
        "user": "test",
        "password": "test",
        "port": 1234
    },
    "contacts": [
        {
            "name": "i test",
            "email": "test@test.com"
        }
    ]
}
`

const exampleConfigAllTOML = `
appname = "test"

[db]
name = "test"
user = "test"
password = "test"
port = 1234

[[contacts]]
name = "i test"
email = "test@test.com"
`

const exampleConfigDefaultedYAML = `
appname: app name

db:
  name: test
  password: test

contacts:
  - email: test@test.com
`

const exampleConfigDefaultedJSON = `
{
    "appname": "app name",
    "db": {
        "name": "test",
		"password": "test"
    },
    "contacts": [
        {
            "email": "test@test.com"
        }
    ]
}
`

const exampleConfigDefaultedTOML = `
appname = "app name"

[db]
name = "test"
password = "test"

[[contacts]]
email = "test@test.com"
`

const exampleConfigMissingRequiredYAML = `
appname: app name

db:
  name: test

contacts:
  - name: i test
    email: test@test.com
`

const exampleConfigMissingRequiredJSON = `
{
    "appname": "test",
    "db": {
        "name": "test"
    },
    "contacts": [
        {
            "name": "i test",
			"email": "test@test.com"
        }
    ]
}
`

const exampleConfigMissingRequiredTOML = `
appname = "test"

[db]
name = "test"

[[contacts]]
name = "i test"
email = "test@test.com"
`

func withTempFile(extension, content string, f func(string)) {
	tmpfile := os.TempDir() + "/configor_test" + extension
	defer os.Remove(tmpfile)
	err := os.WriteFile(tmpfile, []byte(content), 0644)
	if err != nil {
		panic(err)
	}
	f(tmpfile)
}

func withBatch(mapping map[string]string, f func(string)) {
	for extension, content := range mapping {
		withTempFile(extension, content, f)
	}
}

func TestLoadConfig(t *testing.T) {
	withBatch(map[string]string{
		".yaml": exampleConfigAllYAML,
		".json": exampleConfigAllJSON,
		".toml": exampleConfigAllTOML,
	}, func(tmpfile string) {
		config := &ExampleConfig{}
		err := Load(config, tmpfile)
		assert.NoError(t, err)
		assert.Equal(t, "test", config.APPName)
		assert.Equal(t, "test", config.DB.Name)
		assert.Equal(t, "test", config.DB.User)
		assert.Equal(t, "test", config.DB.Password)
		assert.Equal(t, uint(1234), config.DB.Port)
		assert.Equal(t, 1, len(config.Contacts))
		assert.Equal(t, "i test", config.Contacts[0].Name)
		assert.Equal(t, "test@test.com", config.Contacts[0].Email)
	})
	withBatch(map[string]string{
		".yaml": exampleConfigDefaultedYAML,
		".json": exampleConfigDefaultedJSON,
		".toml": exampleConfigDefaultedTOML,
	}, func(tmpfile string) {
		config := &ExampleConfig{}
		err := Load(config, tmpfile)
		assert.NoError(t, err)
		assert.Equal(t, "app name", config.APPName)
		assert.Equal(t, "test", config.DB.Name)
		assert.Equal(t, "root", config.DB.User)
		assert.Equal(t, "test", config.DB.Password)
		assert.Equal(t, uint(3306), config.DB.Port)
		assert.Equal(t, 1, len(config.Contacts))
		assert.Equal(t, "", config.Contacts[0].Name)
		assert.Equal(t, "test@test.com", config.Contacts[0].Email)
	})
	withBatch(map[string]string{
		".yaml": exampleConfigMissingRequiredYAML,
		".json": exampleConfigMissingRequiredJSON,
		".toml": exampleConfigMissingRequiredTOML,
	}, func(tmpfile string) {
		config := &ExampleConfig{}
		err := Load(config, tmpfile)
		assert.Error(t, err)
	})
}

func TestEnvOverride(t *testing.T) {
	conf := New(&Config{
		EnvironmentPrefix:    "GOTIFY_CONFIGOR_TEST",
		ErrorOnUnmatchedKeys: true,
	})

	assert.NoError(t, os.Setenv("DBPassword", "TEST"))
	assert.NoError(t, os.Setenv("GOTIFY_CONFIGOR_TEST_DB_NAME", "name"))

	withBatch(map[string]string{
		".yaml": exampleConfigMissingRequiredYAML,
		".json": exampleConfigMissingRequiredJSON,
		".toml": exampleConfigMissingRequiredTOML,
	}, func(tmpfile string) {
		config := &ExampleConfig{}
		err := conf.Load(config, tmpfile)
		assert.NoError(t, err)
		assert.Equal(t, "name", config.DB.Name)
		assert.Equal(t, "TEST", config.DB.Password)
	})
}
