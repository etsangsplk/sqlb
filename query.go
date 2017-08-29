package sqlb

type Query struct {
    b []byte
    args []interface{}
    sel *selectClause
}

func (q *Query) String() string {
    size := q.sel.size()
    argc := q.sel.argCount()
    if len(q.args) != argc  {
        q.args = make([]interface{}, argc)
    }
    if len(q.b) != size {
        q.b = make([]byte, size)
    }
    q.sel.scan(q.b, q.args)
    return string(q.b)
}

func (q *Query) StringArgs() (string, []interface{}) {
    size := q.sel.size()
    argc := q.sel.argCount()
    if len(q.args) != argc  {
        q.args = make([]interface{}, argc)
    }
    if len(q.b) != size {
        q.b = make([]byte, size)
    }
    q.sel.scan(q.b, q.args)
    return string(q.b), q.args
}

func (q *Query) Where(e *Expression) *Query {
    q.sel.addWhere(e)
    return q
}

func (q *Query) GroupBy(cols ...Columnar) *Query {
    q.sel.addGroupBy(cols...)
    return q
}

func (q *Query) OrderBy(scols ...*sortColumn) *Query {
    q.sel.addOrderBy(scols...)
    return q
}

func (q *Query) Limit(limit int) *Query {
    q.sel.setLimit(limit)
    return q
}

func (q *Query) LimitWithOffset(limit int, offset int) *Query {
    q.sel.setLimitWithOffset(limit, offset)
    return q
}

func (q *Query) As(alias string) *Query {
    q.sel.setAlias(alias)
    return q
}

func Select(items ...element) *Query {
    sel := &selectClause{
        projected: &List{},
    }

    selectionMap := make(map[uint64]selection, 0)
    projectionMap := make(map[uint64]projection, 0)

    // For each scannable item we've received in the call, check what concrete
    // type they are and, depending on which type they are, either add them to
    // the returned selectClause's projected List or query the underlying
    // table metadata to generate a list of all columns in that table.
    for _, item := range items {
        switch item.(type) {
            case *joinClause:
                j := item.(*joinClause)
                if ! containsJoin(sel, j) {
                    sel.joins = append(sel.joins, j)
                    if _, ok := selectionMap[j.left.selectionId()]; ! ok {
                        selectionMap[j.left.selectionId()] = j.left
                        for _, proj := range j.left.projections() {
                            projId := proj.projectionId()
                            _, projExists := projectionMap[projId]
                            if ! projExists {
                                addToProjections(sel, proj)
                                projectionMap[projId] = proj
                            }
                        }
                    }
                    if _, ok := selectionMap[j.right.selectionId()]; ! ok {
                        for _, proj := range j.right.projections() {
                            projId := proj.projectionId()
                            _, projExists := projectionMap[projId]
                            if ! projExists {
                                addToProjections(sel, proj)
                                projectionMap[projId] = proj
                            }
                        }
                    }
                }
            case *Column:
                v := item.(*Column)
                sel.projected.elements = append(sel.projected.elements, v)
                selectionMap[v.tbl.selectionId()] = v.tbl
            case *List:
                v := item.(*List)
                for _, el := range v.elements {
                    sel.projected.elements = append(sel.projected.elements, el)
                    if isColumn(el) {
                        c := el.(*Column)
                        selectionMap[c.tbl.selectionId()] = c.tbl
                    }
                }
            case *Table:
                v := item.(*Table)
                for _, cd := range v.tdef.projections() {
                    addToProjections(sel, cd)
                }
                selectionMap[v.selectionId()] = v
            case *TableDef:
                v := item.(*TableDef)
                for _, cd := range v.projections() {
                    addToProjections(sel, cd)
                }
                selectionMap[v.selectionId()] = v
            case *ColumnDef:
                v := item.(*ColumnDef)
                addToProjections(sel, v)
                selectionMap[v.tdef.selectionId()] = v.tdef
        }
    }
    selections := make([]selection, len(selectionMap))
    x := 0
    for _, sel := range selectionMap {
        selections[x] = sel
        x++
    }
    sel.selections = selections
    return &Query{sel: sel}
}