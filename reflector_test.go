package reflector

import (
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReflector(t *testing.T) {
	type Foo struct {
		ID   uuid.UUID `foo:"6b245e15-5c88-438b-a170-d8f97460083a"`
		List []int     `foo:"1,2,3,4"`
	}

	type Buz struct {
		B bool `foo:"true"`
	}

	type Bar struct {
		Foo
		B    Buz
		N    time.Duration `foo:"1m"`
		S    string
		Skip string `foo:"-"`
	}

	x := Bar{
		S:    "test",
		Skip: "skip",
	}

	r := New(&x)
	m := r.ExtractTags("foo", WithoutMinus())

	assert.Equal(t, map[string]string{
		"ID":   "6b245e15-5c88-438b-a170-d8f97460083a",
		"List": "1,2,3,4",
		"N":    "1m",
		"S":    "",
		"B.B":  "true",
	}, m)
	assert.Equal(t, &x, r.Value())

	require.NoError(t, r.Apply(m))
	assert.Equal(t, Bar{
		Foo: Foo{
			ID:   uuid.FromStringOrNil("6b245e15-5c88-438b-a170-d8f97460083a"),
			List: []int{1, 2, 3, 4},
		},
		B: Buz{
			B: true,
		},
		N:    time.Minute,
		S:    "",
		Skip: "skip",
	}, x)
}
