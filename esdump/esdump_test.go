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

	if url, name, err := parseElasticURL("http://localhost:9200"); assert.NoError(err) {
		require.Equal("http://localhost:9200", url)
		require.Equal("", name)
	}
	if url, name, err := parseElasticURL("http://localhost:9200/index-"); assert.NoError(err) {
		require.Equal("http://localhost:9200", url)
		require.Equal("index-", name)
	}
}
