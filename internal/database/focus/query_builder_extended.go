//go:build qb_extended

package focus

import (
	"fmt"

	"github.com/costscope/costscope/internal/database"
)

// Join adds a JOIN clause (extended build only)
func (qb *FOCUSQueryBuilder) Join(table, condition string) database.QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return qb
}

// LeftJoin adds a LEFT JOIN clause (extended build only)
func (qb *FOCUSQueryBuilder) LeftJoin(table, condition string) database.QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return qb
}

// Having adds a HAVING clause (extended build only)
func (qb *FOCUSQueryBuilder) Having(condition string, args ...interface{}) database.QueryBuilder {
	if len(args) > 0 {
		qb.having = append(qb.having, fmt.Sprintf(condition, args...))
	} else {
		qb.having = append(qb.having, condition)
	}
	return qb
}

// WithCTE adds a CTE (extended build only)
func (qb *FOCUSQueryBuilder) WithCTE(name, query string) database.QueryBuilder {
	qb.ctes = append(qb.ctes, fmt.Sprintf("%s AS (%s)", name, query))
	return qb
}
