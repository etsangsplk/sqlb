package sqlb

type SelectClause struct {
    alias string
    projected *List
    selections []selection
    joins []*JoinClause
    filters []*Expression
    groupBy *GroupByClause
    orderBy *OrderByClause
    limit *LimitClause
}

func (s *SelectClause) argCount() int {
    argc := s.projected.argCount()
    for _, sel := range s.selections {
        argc += sel.argCount()
    }
    for _, join := range s.joins {
        argc += join.argCount()
    }
    for _, filter := range s.filters {
        argc += filter.argCount()
    }
    if s.groupBy != nil {
        argc += s.groupBy.argCount()
    }
    if s.orderBy != nil {
        argc += s.orderBy.argCount()
    }
    if s.limit != nil {
        argc += s.limit.argCount()
    }
    return argc
}

func (s *SelectClause) Alias(alias string) {
    s.alias = alias
}

func (s *SelectClause) As(alias string) *SelectClause {
    s.Alias(alias)
    return s
}

func (s *SelectClause) size() int {
    size := len(Symbols[SYM_SELECT]) + len(Symbols[SYM_FROM])
    size += s.projected.size()
    for _, sel := range s.selections {
        size += sel.size()
    }
    if s.alias != "" {
        size += len(Symbols[SYM_AS]) + len(s.alias)
    }
    for _, join := range s.joins {
        size += join.size()
    }
    nfilters := len(s.filters)
    if nfilters > 0 {
        size += len(Symbols[SYM_WHERE])
        size += len(Symbols[SYM_AND]) * (nfilters - 1)
        for _, filter := range s.filters {
            size += filter.size()
        }
    }
    if s.groupBy != nil {
        size += s.groupBy.size()
    }
    if s.orderBy != nil {
        size += s.orderBy.size()
    }
    if s.limit != nil {
        size += s.limit.size()
    }
    return size
}

func (s *SelectClause) scan(b []byte, args []interface{}) (int, int) {
    var bw, ac int
    bw += copy(b[bw:], Symbols[SYM_SELECT])
    pbw, pac := s.projected.scan(b[bw:], args)
    bw += pbw
    ac += pac
    bw += copy(b[bw:], Symbols[SYM_FROM])
    for _, sel := range s.selections {
        sbw, sac := sel.scan(b[bw:], args)
        bw += sbw
        ac += sac
    }
    if s.alias != "" {
        bw += copy(b[bw:], Symbols[SYM_AS])
        bw += copy(b[bw:], s.alias)
    }
    for _, join := range s.joins {
        jbw, jac := join.scan(b[bw:], args)
        bw += jbw
        ac += jac
    }
    if len(s.filters) > 0 {
        bw += copy(b[bw:], Symbols[SYM_WHERE])
        for x, filter := range s.filters {
            if x > 0 {
                bw += copy(b[bw:], Symbols[SYM_AND])
            }
            fbw, fac := filter.scan(b[bw:], args[ac:])
            bw += fbw
            ac += fac
        }
    }
    if s.groupBy != nil {
        gbbw, gbac := s.groupBy.scan(b[bw:], args[ac:])
        bw += gbbw
        ac += gbac
    }
    if s.orderBy != nil {
        obbw, obac := s.orderBy.scan(b[bw:], args[ac:])
        bw += obbw
        ac += obac
    }
    if s.limit != nil {
        lbw, lac := s.limit.scan(b[bw:], args[ac:])
        bw += lbw
        ac += lac
    }
    return bw, ac
}

func (s *SelectClause) String() string {
    size := s.size()
    argc := s.argCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.scan(b, args)
    return string(b)
}

func (s *SelectClause) StringArgs() (string, []interface{}) {
    size := s.size()
    argc := s.argCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.scan(b, args)
    return string(b), args
}

func (s *SelectClause) Where(e *Expression) *SelectClause {
    s.filters = append(s.filters, e)
    return s
}

// Given one or more columns, either set or add to the GROUP BY clause for
// the SelectClause
func (s *SelectClause) GroupBy(cols ...Columnar) *SelectClause {
    if len(cols) == 0 {
        return s
    }
    gb := s.groupBy
    if gb == nil {
        gb = &GroupByClause{
            cols: &List{
                elements: make([]element, len(cols)),
            },
        }
        for x, c := range cols {
            gb.cols.elements[x] = c.Column()
        }
    } else {
        for _, c := range cols {
            gb.cols.elements = append(gb.cols.elements, c.Column())
        }
    }
    s.groupBy = gb
    return s
}

// Given one or more sort columns, either set or add to the ORDER BY clause for
// the SelectClause
func (s *SelectClause) OrderBy(sortCols ...*SortColumn) *SelectClause {
    if len(sortCols) == 0 {
        return s
    }
    ob := s.orderBy
    if ob == nil {
        ob = &OrderByClause{
            cols: &List{
                elements: make([]element, len(sortCols)),
            },
        }
        for x, sc := range sortCols {
            ob.cols.elements[x] = sc
        }
    } else {
        for _, sc := range sortCols {
            ob.cols.elements = append(ob.cols.elements, sc)
        }
    }
    s.orderBy = ob
    return s
}

func (s *SelectClause) LimitWithOffset(limit int, offset int) *SelectClause {
    lc := &LimitClause{limit: limit}
    lc.offset = &offset
    s.limit = lc
    return s
}

func (s *SelectClause) Limit(limit int) *SelectClause {
    lc := &LimitClause{limit: limit}
    s.limit = lc
    return s
}

func containsJoin(s *SelectClause, j *JoinClause) bool {
    for _, sj := range s.joins {
        if j == sj {
            return true
        }
    }
    return false
}

func addToProjections(s *SelectClause, p projection) {
    s.projected.elements = append(s.projected.elements, p)
}

func Select(items ...element) *SelectClause {
    // TODO(jaypipes): Make the memory allocation more efficient below by
    // looping through the elements and determining the number of element struct
    // pointers to allocate instead of just making an empty array of element
    // pointers.
    res := &SelectClause{
        projected: &List{},
    }

    selectionMap := make(map[uint64]selection, 0)
    projectionMap := make(map[uint64]projection, 0)

    // For each scannable item we've received in the call, check what concrete
    // type they are and, depending on which type they are, either add them to
    // the returned SelectClause's projected List or query the underlying
    // table metadata to generate a list of all columns in that table.
    for _, item := range items {
        switch item.(type) {
            case *JoinClause:
                j := item.(*JoinClause)
                if ! containsJoin(res, j) {
                    res.joins = append(res.joins, j)
                    if _, ok := selectionMap[j.left.selectionId()]; ! ok {
                        selectionMap[j.left.selectionId()] = j.left
                        for _, proj := range j.left.projections() {
                            projId := proj.projectionId()
                            _, projExists := projectionMap[projId]
                            if ! projExists {
                                addToProjections(res, proj)
                                projectionMap[projId] = proj
                            }
                        }
                    }
                    if _, ok := selectionMap[j.right.selectionId()]; ! ok {
                        for _, proj := range j.right.projections() {
                            projId := proj.projectionId()
                            _, projExists := projectionMap[projId]
                            if ! projExists {
                                addToProjections(res, proj)
                                projectionMap[projId] = proj
                            }
                        }
                    }
                }
            case *Column:
                v := item.(*Column)
                res.projected.elements = append(res.projected.elements, v)
                selectionMap[v.tbl.selectionId()] = v.tbl
            case *List:
                v := item.(*List)
                for _, el := range v.elements {
                    res.projected.elements = append(res.projected.elements, el)
                    if isColumn(el) {
                        c := el.(*Column)
                        selectionMap[c.tbl.selectionId()] = c.tbl
                    }
                }
            case *Table:
                v := item.(*Table)
                for _, cd := range v.tdef.projections() {
                    addToProjections(res, cd)
                }
                selectionMap[v.selectionId()] = v
            case *TableDef:
                v := item.(*TableDef)
                for _, cd := range v.projections() {
                    addToProjections(res, cd)
                }
                selectionMap[v.selectionId()] = v
            case *ColumnDef:
                v := item.(*ColumnDef)
                addToProjections(res, v)
                selectionMap[v.tdef.selectionId()] = v.tdef
        }
    }
    selections := make([]selection, len(selectionMap))
    x := 0
    for _, sel := range selectionMap {
        selections[x] = sel
        x++
    }
    res.selections = selections
    return res
}
