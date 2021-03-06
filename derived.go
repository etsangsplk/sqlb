//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//
package sqlb

// A derived table is a SELECT in the FROM clause. It is always aliased and the
// projections for a derived table take this alias as their selection alias.
//
// The projections of a derived table are not the same as the projections for
// the SELECT that is being wrapped. For example, given the following SQL
// statement:
//
// SELECT u.id, u.name FROM (
//   SELECT users.id, users.name FROM users
// ) AS u
//
// The inner SELECT's projections are columns from the users Table or TableDef.
// However, the derived table's projections are separate and include the alias
// of the derived table as the selection alias (u instead of users).
type derivedTable struct {
	alias string
	from  *selectStatement
}

// Return a collection of derivedColumn projections that have been constructed
// to refer to this derived table and not have any outer alias
func (dt *derivedTable) getAllDerivedColumns() []projection {
	nprojs := len(dt.from.projs)
	projs := make([]projection, nprojs)
	for x := 0; x < nprojs; x++ {
		p := dt.from.projs[x]
		switch p.(type) {
		case *Column:
			projs[x] = &derivedColumn{dt: dt, c: p.(*Column)}
		}
	}
	return projs
}

func (dt *derivedTable) projections() []projection {
	nprojs := len(dt.from.projs)
	projs := make([]projection, nprojs)
	for x := 0; x < nprojs; x++ {
		p := dt.from.projs[x]
		switch p.(type) {
		case *Column:
		}
	}
	return projs
}

func (dt *derivedTable) argCount() int {
	return dt.from.argCount()
}

func (dt *derivedTable) size(scanner *sqlScanner) int {
	size := dt.from.size(scanner)
	size += (len(Symbols[SYM_LPAREN]) + len(Symbols[SYM_RPAREN]) +
		len(Symbols[SYM_AS]) + len(dt.alias))
	return size
}

func (dt *derivedTable) scan(scanner *sqlScanner, b []byte, args []interface{}, curArg *int) int {
	bw := 0
	bw += copy(b[bw:], Symbols[SYM_LPAREN])
	bw += dt.from.scan(scanner, b[bw:], args, curArg)
	bw += copy(b[bw:], Symbols[SYM_RPAREN])
	bw += copy(b[bw:], Symbols[SYM_AS])
	bw += copy(b[bw:], dt.alias)
	return bw
}

// A derivedColumn is a type of projection that is produced from a derived
// table (SELECT in the FROM clause). What makes a derived column unique is
// that it uses the alias of the underlying column as its name in the outer
// projection.
//
// The inner projection is a column against an underlying table or table def.
// The outer projection will have the selection alias of the derived table and
// the name of the projection will be the alias or name of the underlying
// column. For example, given the following SQL:
//
// SELECT <outer> FROM (
//   SELECT users.id, users.name FROM users
// ) AS u
//
// <outer> should contain:
//
// &derivedColumn{dt: dt, c: &Column{name: "id", tbl: users}},
// &derivedColumn{dt: dt, c: &Column{name: "name", tbl: users}}
//
// when scanned into <outer>, that should produce:
//
// []byte("u.id, u.name")
//
// However, let's say that the inner projections have been
// aliased, like so:
//
// SELECT <outer> FROM (
//   SELECT
//     users.id AS user_id,
//     users.name AS user_name
//   FROM users
// ) AS u
//
// <outer> should instead contain:
//
// &derivedColumn{dt: dt, c: &Column{alias: "user_id". name: "id", tbl: users}},
// &derivedColumn{dt: dt, c: &Column{alias: "user_name", name: "name", tbl: users}}
//
// which, when scanned into <outer>, should produce:
//
// []byte("u.user_id, u.user_name")
//
// Finally, the derivedColumn can itself have an alias, which can result in the
// outermost projection looking like so:
//
// SELECT u.user_name AS uname FROM (
//   SELECT users.name AS user_name
//   FROM users
// ) AS u
type derivedColumn struct {
	alias string // This is the outermost alias
	c     *Column
	dt    *derivedTable
}

func (dc *derivedColumn) from() selection {
	return dc.dt
}

func (dc *derivedColumn) disableAliasScan() func() {
	origAlias := dc.alias
	dc.alias = ""
	return func() { dc.alias = origAlias }
}

func (dc *derivedColumn) argCount() int {
	return 0
}

func (dc *derivedColumn) size(scanner *sqlScanner) int {
	size := len(dc.dt.alias)
	size += len(Symbols[SYM_PERIOD])
	if dc.c.alias != "" {
		size += len(dc.c.alias)
	} else {
		size += len(dc.c.name)
	}
	if dc.alias != "" {
		size += len(Symbols[SYM_AS]) + len(dc.alias)
	}
	return size
}

func (dc *derivedColumn) scan(scanner *sqlScanner, b []byte, args []interface{}, curArg *int) int {
	bw := 0
	bw += copy(b[bw:], dc.dt.alias)
	bw += copy(b[bw:], Symbols[SYM_PERIOD])
	if dc.c.alias != "" {
		bw += copy(b[bw:], dc.c.alias)
	} else {
		bw += copy(b[bw:], dc.c.name)
	}
	if dc.alias != "" {
		bw += copy(b[bw:], Symbols[SYM_AS])
		bw += copy(b[bw:], dc.alias)
	}
	return bw
}
