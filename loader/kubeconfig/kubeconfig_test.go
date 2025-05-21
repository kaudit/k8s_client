package kubeconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain and test data setup
func TestMain(m *testing.M) {
	// Setup test files
	setupTestFiles()

	// Run tests
	code := m.Run()

	// Clean up test files
	cleanupTestFiles()

	os.Exit(code)
}

// Test data and helpers
var (
	testDir      string
	validConfig  = []byte("valid-kube-config-data")
	emptyConfig  = []byte("")
	validPath    string
	nonExistPath string
	emptyPath    string
	noReadPath   string
)

func setupTestFiles() {
	// Create a temporary directory for test files
	var err error
	testDir, err = os.MkdirTemp("", "k8s-auth-test-")
	if err != nil {
		panic("Failed to create temp directory: " + err.Error())
	}

	// Create a valid config file
	validPath = filepath.Join(testDir, "valid-config.yaml")
	if err := os.WriteFile(validPath, validConfig, 0644); err != nil {
		panic("Failed to create test file: " + err.Error())
	}

	// Create an empty config file
	emptyPath = filepath.Join(testDir, "empty-config.yaml")
	if err := os.WriteFile(emptyPath, emptyConfig, 0644); err != nil {
		panic("Failed to create empty test file: " + err.Error())
	}

	// Create a file with no read permissions
	noReadPath = filepath.Join(testDir, "no-read-permission.yaml")
	if err := os.WriteFile(noReadPath, []byte("test"), 0200); err != nil {
		panic("Failed to create no-read test file: " + err.Error())
	}

	// Define a path that doesn't exist
	nonExistPath = filepath.Join(testDir, "non-existent.yaml")
}

func cleanupTestFiles() {
	_ = os.RemoveAll(testDir)
}

// Tests for Load method
func TestLoad(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		// Create loader with valid path
		l := NewK8sConfigLoader(validPath)

		// Load the config
		data, err := l.Load()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, validConfig, data)
	})

	t.Run("loads empty config file", func(t *testing.T) {
		// Create loader with empty file path
		l := NewK8sConfigLoader(emptyPath)

		// Load the config
		data, err := l.Load()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, emptyConfig, data)
	})

	t.Run("fails on non-existent file", func(t *testing.T) {
		// Create loader with non-existent path
		l := NewK8sConfigLoader(nonExistPath)

		// Load the config
		data, err := l.Load()

		// Assert
		require.Error(t, err)
		assert.Nil(t, data)
		// The specific error message comes from the implementation
	})

	t.Run("fails on empty path", func(t *testing.T) {
		// Create loader with empty path
		l := NewK8sConfigLoader("")

		// Load the config
		data, err := l.Load()

		// Assert
		require.Error(t, err)
		assert.Nil(t, data)
		// The specific error message comes from the implementation
	})

	t.Run("fails when file has no read permissions", func(t *testing.T) {
		// Skip on Windows as file permissions work differently
		if os.Getenv("GOOS") == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		// Create loader pointing to a file that exists but can't be read
		l := NewK8sConfigLoader(noReadPath)

		// Load the config
		data, err := l.Load()

		// Assert
		require.Error(t, err)
		assert.Nil(t, data)
		// The specific error message comes from the implementation
	})
}
