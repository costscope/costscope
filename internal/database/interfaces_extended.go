//go:build qb_extended

package database

// ExtendedQueryBuilder augments the core QueryBuilder when built with the
// `qb_extended` tag. Implementations should embed or wrap the core builder.
// NOTE: This preserves backward compatibility while allowing advanced features
// to iterate off the fast path.
//
// To enable: go build -tags qb_extended ./...
// If the tag is absent these methods are not part of the build, reducing size
// and dead surface.
//
// Telemetry: Each Build/BuildCount/BuildExplain invocation increments
// costscope_querybuilder_build_total{type="extended"} so runtime usage of the
// extended surface can be observed and potentially pruned later.

type ExtendedQueryBuilder interface {
	QueryBuilder

	// Advanced operations
	Join(table, condition string) QueryBuilder
	LeftJoin(table, condition string) QueryBuilder
	Having(condition string, args ...interface{}) QueryBuilder
	WithCTE(name, query string) QueryBuilder

	// Secondary build forms
	BuildCount() (string, []interface{}, error)
	BuildExplain() (string, []interface{}, error)
}

// (Optional) helper to assert at compile time when implementing an extended builder.
// var _ ExtendedQueryBuilder = (*YourBuilder)(nil)
