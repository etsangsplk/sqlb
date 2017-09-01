package sqlb

type SqlType int

const (
    SQL_TYPE_CHAR SqlType = iota
    SQL_TYPE_INT
    SQL_TYPE_FLOAT
    SQL_TYPE_DECIMAL
    SQL_TYPE_VARCHAR
    SQL_TYPE_TEXT
    SQL_TYPE_BINARY
)

var (
    sqlTypeToSymbol = map[SqlType]Symbol{
        SQL_TYPE_CHAR: SYM_TYPE_CHAR,
        SQL_TYPE_VARCHAR: SYM_TYPE_VARCHAR,
        SQL_TYPE_INT: SYM_TYPE_INT,
        SQL_TYPE_BINARY: SYM_TYPE_BINARY,
        SQL_TYPE_FLOAT: SYM_TYPE_FLOAT,
        SQL_TYPE_DECIMAL: SYM_TYPE_DECIMAL,
        SQL_TYPE_TEXT: SYM_TYPE_TEXT,
    }
)
