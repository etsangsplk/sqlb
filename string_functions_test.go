package sqlb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrimFunctions(t *testing.T) {
	assert := assert.New(t)

	m := testFixtureMeta()
	users := m.Table("users")
	colUserName := users.C("name")

	tests := []struct {
		name  string
		el    element
		qs    map[Dialect]string
		qargs []interface{}
	}{
		{
			name: "TRIM(column) or BTRIM(column)",
			el:   Trim(colUserName),
			qs: map[Dialect]string{
				DIALECT_MYSQL:      "TRIM(users.name)",
				DIALECT_POSTGRESQL: "BTRIM(users.name)",
			},
		},
		{
			name: "LTRIM(column) or TRIM(LEADING FROM column)",
			el:   LTrim(colUserName),
			qs: map[Dialect]string{
				DIALECT_MYSQL:      "LTRIM(users.name)",
				DIALECT_POSTGRESQL: "TRIM(LEADING FROM users.name)",
			},
		},
		{
			name: "RTRIM(column) or TRIM(TRAILING FROM column)",
			el:   RTrim(colUserName),
			qs: map[Dialect]string{
				DIALECT_MYSQL:      "RTRIM(users.name)",
				DIALECT_POSTGRESQL: "TRIM(TRAILING FROM users.name)",
			},
		},
	}
	for _, test := range tests {
		expArgc := len(test.qargs)
		argc := test.el.argCount()
		assert.Equal(expArgc, argc)

		// Test each SQL dialect output
		for dialect, qs := range test.qs {
			test.el.setDialect(dialect)
			expLen := len(qs)
			size := test.el.size()
			size += interpolationLength(dialect, argc)
			assert.Equal(expLen, size)

			b := make([]byte, size)
			curArg := 0
			written := test.el.scan(b, test.qargs, &curArg)

			assert.Equal(written, size)
			assert.Equal(qs, string(b))
		}
	}
}