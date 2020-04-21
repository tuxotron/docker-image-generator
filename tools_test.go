package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileNameWithoutExtension(t *testing.T) {

	filename := filenameWithoutExtension("test.ini")
	assert.Equal(t, "test", filename)

	filename = filenameWithoutExtension("test.ext.ini")
	assert.Equal(t, "test.ext", filename)

}
