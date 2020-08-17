package esdump

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseElasticURL(t *testing.T) {
	assert := assert.New(t)
	assert.NotNil(assert)
	require := require.New(t)
	require.NotNil(require)

	if url, name, isFile, err := parseElasticURL("http://localhost:9200"); assert.NoError(err) {
		require.Equal("http://localhost:9200", url)
		require.Equal("", name)
		require.False(isFile)
	}
	if url, name, isFile, err := parseElasticURL("http://localhost:9200/index-"); assert.NoError(err) {
		require.Equal("http://localhost:9200", url)
		require.Equal("index-", name)
		require.False(isFile)
	}

	if url, name, isFile, err := parseElasticURL("test.txt"); assert.NoError(err) {
		require.Equal("test.txt", url)
		require.Equal("test.txt", name)
		require.True(isFile)
	}

	if url, name, isFile, err := parseElasticURL("/var/dump/es/dump1.txt"); assert.NoError(err) {
		require.Equal("/var/dump/es/dump1.txt", url)
		require.Equal("/var/dump/es/dump1.txt", name)
		require.True(isFile)
	}
	if url, name, isFile, err := parseElasticURL("file:///var/dump/es/dump1.txt"); assert.NoError(err) {
		require.Equal("file:///var/dump/es/dump1.txt", url)
		require.Equal("/var/dump/es/dump1.txt", name)
		require.True(isFile)
	}

}
