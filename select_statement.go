//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//
package sqlb

type selectStatement struct {
	projs      []projection
	selections []selection
	joins      []*joinClause
	where      *whereClause
	groupBy    *groupByClause
	having     *havingClause
	orderBy    *orderByClause
	limit      *limitClause
}

func (s *selectStatement) argCount() int {
	argc := 0
	for _, p := range s.projs {
		argc += p.argCount()
	}
	for _, sel := range s.selections {
		argc += sel.argCount()
	}
	for _, join := range s.joins {
		argc += join.argCount()
	}
	if s.where != nil {
		argc += s.where.argCount()
	}
	if s.groupBy != nil {
		argc += s.groupBy.argCount()
	}
	if s.having != nil {
		argc += s.having.argCount()
	}
	if s.orderBy != nil {
		argc += s.orderBy.argCount()
	}
	if s.limit != nil {
		argc += s.limit.argCount()
	}
	return argc
}

func (s *selectStatement) size(scanner *sqlScanner) int {
	size := len(Symbols[SYM_SELECT])
	nprojs := len(s.projs)
	for _, p := range s.projs {
		size += p.size(scanner)
	}
	size += (len(Symbols[SYM_COMMA_WS]) * (nprojs - 1)) // the commas...
	nsels := len(s.selections)
	if nsels > 0 {
		size += len(scanner.format.SeparateClauseWith)
		size += len(Symbols[SYM_FROM])
		for _, sel := range s.selections {
			size += sel.size(scanner)
		}
		size += (len(Symbols[SYM_COMMA_WS]) * (nsels - 1)) // the commas...
		for _, join := range s.joins {
			size += join.size(scanner)
		}
	}
	if s.where != nil {
		size += s.where.size(scanner)
	}
	if s.groupBy != nil {
		size += s.groupBy.size(scanner)
	}
	if s.having != nil {
		size += s.having.size(scanner)
	}
	if s.orderBy != nil {
		size += s.orderBy.size(scanner)
	}
	if s.limit != nil {
		size += s.limit.size(scanner)
	}
	return size
}

func (s *selectStatement) scan(scanner *sqlScanner, b []byte, args []interface{}, curArg *int) int {
	bw := 0
	bw += copy(b[bw:], Symbols[SYM_SELECT])
	nprojs := len(s.projs)
	for x, p := range s.projs {
		bw += p.scan(scanner, b[bw:], args, curArg)
		if x != (nprojs - 1) {
			bw += copy(b[bw:], Symbols[SYM_COMMA_WS])
		}
	}
	nsels := len(s.selections)
	if nsels > 0 {
		bw += copy(b[bw:], scanner.format.SeparateClauseWith)
		bw += copy(b[bw:], Symbols[SYM_FROM])
		for x, sel := range s.selections {
			bw += sel.scan(scanner, b[bw:], args, curArg)
			if x != (nsels - 1) {
				bw += copy(b[bw:], Symbols[SYM_COMMA_WS])
			}
		}
		for _, join := range s.joins {
			bw += join.scan(scanner, b[bw:], args, curArg)
		}
	}
	if s.where != nil {
		bw += s.where.scan(scanner, b[bw:], args, curArg)
	}
	if s.groupBy != nil {
		bw += s.groupBy.scan(scanner, b[bw:], args, curArg)
	}
	if s.having != nil {
		bw += s.having.scan(scanner, b[bw:], args, curArg)
	}
	if s.orderBy != nil {
		bw += s.orderBy.scan(scanner, b[bw:], args, curArg)
	}
	if s.limit != nil {
		bw += s.limit.scan(scanner, b[bw:], args, curArg)
	}
	return bw
}

func (s *selectStatement) addJoin(jc *joinClause) *selectStatement {
	s.joins = append(s.joins, jc)
	return s
}

func (s *selectStatement) addWhere(e *Expression) *selectStatement {
	if s.where == nil {
		s.where = &whereClause{filters: make([]*Expression, 0)}
	}
	s.where.filters = append(s.where.filters, e)
	return s
}

// Given one or more columns, either set or add to the GROUP BY clause for
// the selectStatement
func (s *selectStatement) addGroupBy(cols ...projection) *selectStatement {
	if len(cols) == 0 {
		return s
	}
	gb := s.groupBy
	if gb == nil {
		gb = &groupByClause{
			cols: make([]projection, len(cols)),
		}
		for x, c := range cols {
			gb.cols[x] = c
		}
	} else {
		for _, c := range cols {
			gb.cols = append(gb.cols, c)
		}
	}
	s.groupBy = gb
	return s
}

func (s *selectStatement) addHaving(e *Expression) *selectStatement {
	if s.having == nil {
		s.having = &havingClause{conditions: make([]*Expression, 0)}
	}
	s.having.conditions = append(s.having.conditions, e)
	return s
}

// Given one or more sort columns, either set or add to the ORDER BY clause for
// the selectStatement
func (s *selectStatement) addOrderBy(sortCols ...*sortColumn) *selectStatement {
	if len(sortCols) == 0 {
		return s
	}
	ob := s.orderBy
	if ob == nil {
		ob = &orderByClause{
			scols: make([]*sortColumn, len(sortCols)),
		}
		for x, sc := range sortCols {
			ob.scols[x] = sc
		}
	} else {
		for _, sc := range sortCols {
			ob.scols = append(ob.scols, sc)
		}
	}
	s.orderBy = ob
	return s
}

func (s *selectStatement) setLimitWithOffset(limit int, offset int) *selectStatement {
	lc := &limitClause{limit: limit}
	lc.offset = &offset
	s.limit = lc
	return s
}

func (s *selectStatement) setLimit(limit int) *selectStatement {
	lc := &limitClause{limit: limit}
	s.limit = lc
	return s
}

func containsJoin(s *selectStatement, j *joinClause) bool {
	for _, sj := range s.joins {
		if j == sj {
			return true
		}
	}
	return false
}

func addToProjections(s *selectStatement, p projection) {
	s.projs = append(s.projs, p)
}

func (s *selectStatement) removeSelection(toRemove selection) {
	idx := -1
	for x, sel := range s.selections {
		if sel == toRemove {
			idx = x
			break
		}
	}
	if idx == -1 {
		return
	}
	s.selections = append(s.selections[:idx], s.selections[idx+1:]...)
}
