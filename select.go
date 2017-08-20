package sqlb

type Selectable struct {
    alias string
    projected *List
    subjects []Element
    filters []*Expression
    groupBy *GroupByClause
    orderBy *OrderByClause
    limit *LimitClause
}

func (s *Selectable) ArgCount() int {
    argc := s.projected.ArgCount()
    for _, subj := range s.subjects {
        argc += subj.ArgCount()
    }
    for _, filter := range s.filters {
        argc += filter.ArgCount()
    }
    if s.groupBy != nil {
        argc += s.groupBy.ArgCount()
    }
    if s.orderBy != nil {
        argc += s.orderBy.ArgCount()
    }
    if s.limit != nil {
        argc += s.limit.ArgCount()
    }
    return argc
}

func (s *Selectable) Alias(alias string) {
    s.alias = alias
}

func (s *Selectable) As(alias string) *Selectable {
    s.Alias(alias)
    return s
}

func (s *Selectable) Size() int {
    size := len(Symbols[SYM_SELECT]) + len(Symbols[SYM_FROM])
    size += s.projected.Size()
    for _, subj := range s.subjects {
        size += subj.Size()
    }
    if s.alias != "" {
        size += len(Symbols[SYM_AS]) + len(s.alias)
    }
    nfilters := len(s.filters)
    if nfilters > 0 {
        size += len(Symbols[SYM_WHERE])
        size += len(Symbols[SYM_AND]) * (nfilters - 1)
        for _, filter := range s.filters {
            size += filter.Size()
        }
    }
    if s.groupBy != nil {
        size += s.groupBy.Size()
    }
    if s.orderBy != nil {
        size += s.orderBy.Size()
    }
    if s.limit != nil {
        size += s.limit.Size()
    }
    return size
}

func (s *Selectable) Scan(b []byte, args []interface{}) (int, int) {
    var bw, ac int
    bw += copy(b[bw:], Symbols[SYM_SELECT])
    pbw, pac := s.projected.Scan(b[bw:], args)
    bw += pbw
    ac += pac
    bw += copy(b[bw:], Symbols[SYM_FROM])
    for _, subj := range s.subjects {
        sbw, sac := subj.Scan(b[bw:], args)
        bw += sbw
        ac += sac
    }
    if s.alias != "" {
        bw += copy(b[bw:], Symbols[SYM_AS])
        bw += copy(b[bw:], s.alias)
    }
    if len(s.filters) > 0 {
        bw += copy(b[bw:], Symbols[SYM_WHERE])
        for x, filter := range s.filters {
            if x > 0 {
                bw += copy(b[bw:], Symbols[SYM_AND])
            }
            fbw, fac := filter.Scan(b[bw:], args[ac:])
            bw += fbw
            ac += fac
        }
    }
    if s.groupBy != nil {
        gbbw, gbac := s.groupBy.Scan(b[bw:], args[ac:])
        bw += gbbw
        ac += gbac
    }
    if s.orderBy != nil {
        obbw, obac := s.orderBy.Scan(b[bw:], args[ac:])
        bw += obbw
        ac += obac
    }
    if s.limit != nil {
        lbw, lac := s.limit.Scan(b[bw:], args[ac:])
        bw += lbw
        ac += lac
    }
    return bw, ac
}

func (s *Selectable) String() string {
    size := s.Size()
    argc := s.ArgCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.Scan(b, args)
    return string(b)
}

func (s *Selectable) StringArgs() (string, []interface{}) {
    size := s.Size()
    argc := s.ArgCount()
    args := make([]interface{}, argc)
    b := make([]byte, size)
    s.Scan(b, args)
    return string(b), args
}

func (s *Selectable) Where(e *Expression) *Selectable {
    s.filters = append(s.filters, e)
    return s
}

// Given one or more columns, either set or add to the GROUP BY clause for
// the Selectable
func (s *Selectable) GroupBy(cols ...Columnar) *Selectable {
    if len(cols) == 0 {
        return s
    }
    gb := s.groupBy
    if gb == nil {
        gb = &GroupByClause{
            cols: &List{
                elements: make([]Element, len(cols)),
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
// the Selectable
func (s *Selectable) OrderBy(sortCols ...*SortColumn) *Selectable {
    if len(sortCols) == 0 {
        return s
    }
    ob := s.orderBy
    if ob == nil {
        ob = &OrderByClause{
            cols: &List{
                elements: make([]Element, len(sortCols)),
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

func (s *Selectable) LimitWithOffset(limit int, offset int) *Selectable {
    lc := &LimitClause{limit: limit}
    lc.offset = &offset
    s.limit = lc
    return s
}

func (s *Selectable) Limit(limit int) *Selectable {
    lc := &LimitClause{limit: limit}
    s.limit = lc
    return s
}

func Select(items ...Element) *Selectable {
    // TODO(jaypipes): Make the memory allocation more efficient below by
    // looping through the elements and determining the number of element struct
    // pointers to allocate instead of just making an empty array of Element
    // pointers.
    res := &Selectable{
        projected: &List{},
    }

    subjSet := make(map[Element]bool, 0)

    // For each scannable item we've received in the call, check what concrete
    // type they are and, depending on which type they are, either add them to
    // the returned Selectable's projected List or query the underlying
    // table metadata to generate a list of all columns in that table.
    for _, item := range items {
        switch item.(type) {
            case *Column:
                v := item.(*Column)
                res.projected.elements = append(res.projected.elements, v)
                subjSet[v.def.table] = true
            case *List:
                v := item.(*List)
                for _, el := range v.elements {
                    res.projected.elements = append(res.projected.elements, el)
                    if isColumn(el) {
                        c := el.(*Column)
                        subjSet[c.def.table] = true
                    }
                }
            case *Table:
                v := item.(*Table)
                for _, c := range v.Columns() {
                    res.projected.elements = append(res.projected.elements, c)
                }
                subjSet[v] = true
            case *TableDef:
                v := item.(*TableDef)
                for _, cd := range v.ColumnDefs() {
                    c := &Column{def: cd}
                    res.projected.elements = append(res.projected.elements, c)
                }
                t := &Table{def: v}
                subjSet[t] = true
            case *ColumnDef:
                v := item.(*ColumnDef)
                c := &Column{def: v}
                res.projected.elements = append(res.projected.elements, c)
                subjSet[v.table] = true
        }
    }
    subjects := make([]Element, len(subjSet))
    x := 0
    for scannable, _ := range subjSet {
        subjects[x] = scannable
        x++
    }
    res.subjects = subjects
    return res
}