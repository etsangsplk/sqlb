package sqlb

type Columnar interface {
    Column() *Column
}

type Column struct {
    alias string
    def *ColumnDef
}

func (c *Column) Column() *Column {
    return c
}

func (c *Column) ArgCount() int {
    return 0
}

func (c *Column) Size() int {
    size := c.def.Size()
    if c.alias != "" {
        size += len(Symbols[SYM_AS]) + len(c.alias)
    }
    return size
}

func (c *Column) Scan(b []byte, args []interface{}) (int, int) {
    bw, _ := c.def.Scan(b, args)
    if c.alias != "" {
        bw += copy(b[bw:], Symbols[SYM_AS])
        bw += copy(b[bw:], c.alias)
    }
    return bw, 0
}

func (c *Column) Alias(alias string) {
    c.alias = alias
}

func (c *Column) As(alias string) *Column {
    c.Alias(alias)
    return c
}

func (c *Column) Desc() *SortColumn {
    return &SortColumn{el: c, desc: true}
}

func (c *Column) Asc() *SortColumn {
    return &SortColumn{el: c}
}

func isColumn(el Element) bool {
    switch el.(type) {
    case *Column:
        return true
    default:
        return false
    }
}