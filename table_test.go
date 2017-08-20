package sqlb

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestTable(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    t1 := &Table{
        def: td,
    }

    exp := "users"
    expLen := len(exp)
    s := t1.Size()
    assert.Equal(expLen, s)

    b := make([]byte, s)
    written, _ := t1.Scan(b, nil)

    assert.Equal(written, s)
    assert.Equal(exp, string(b))
}

func TestTableAlias(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    t1 := &Table{
        def: td,
        alias: "u",
    }

    exp := "users AS u"
    expLen := len(exp)
    s := t1.Size()
    assert.Equal(expLen, s)

    b := make([]byte, s)
    written, _ := t1.Scan(b, nil)

    assert.Equal(written, s)
    assert.Equal(exp, string(b))
}

func TestTableColumnDefs(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cdefs := []*ColumnDef{
         &ColumnDef{
            name: "id",
            table: td,
        },
        &ColumnDef{
            name: "email",
            table: td,
        },
    }
    td.columns = cdefs

    defs := td.ColumnDefs()

    assert.Equal(2, len(defs))
    for _, def := range defs {
        assert.Equal(td, def.table)
    }

    // Check stable order of insertion from above...
    assert.Equal(defs[0].name, "id")
    assert.Equal(defs[1].name, "email")
}

func TestTableColumn(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    cdefs := []*ColumnDef{
        &ColumnDef{
            name: "id",
            table: td,
        },
        &ColumnDef{
            name: "email",
            table: td,
        },
    }
    td.columns = cdefs

    c := td.Column("email")

    assert.Equal(td, c.table)
    assert.Equal("email", c.name)

    // Check an unknown column name returns nil
    unknown := td.Column("unknown")
    assert.Nil(unknown)
}

func TestTableAs(t *testing.T) {
    assert := assert.New(t)

    td := &TableDef{
        name: "users",
        schema: "test",
    }

    t1 := td.As("u")
    assert.Equal("u", t1.alias)
    assert.Equal(td, t1.def)
}