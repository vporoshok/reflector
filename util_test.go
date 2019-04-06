package reflector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTagsFromStruct(t *testing.T) {
	v := struct {
		A string `foo:"bar"`
		B int    `foo:"baz"`
		C bool   `foo:"-"`
		D float64
	}{
		A: "test",
		B: 12,
		C: true,
		D: 42,
	}

	names := ExtractTagsFromStruct("foo", v)
	assert.Len(t, names, 2)
	assert.Contains(t, names, "bar")
	assert.Contains(t, names, "baz")
}

func TestStructToMapByTags(t *testing.T) {
	v := struct {
		A string   `foo:"bar"`
		B int      `foo:"baz"`
		C bool     `foo:"-"`
		D *float64 `foo:"d"`
		E *float64
	}{
		A: "test",
		B: 12,
		C: true,
	}

	withoutNils := StructToMapByTags("foo", &v, true)
	assert.Len(t, withoutNils, 2)
	assert.Equal(t, "test", withoutNils["bar"])
	assert.Equal(t, 12, withoutNils["baz"])

	withNils := StructToMapByTags("foo", &v, false)
	assert.Len(t, withNils, 3)
	assert.Equal(t, "test", withNils["bar"])
	assert.Equal(t, 12, withNils["baz"])
	assert.Nil(t, withNils["d"])
}
