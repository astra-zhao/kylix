// orm_query.go — SQL query builder with fluent API.
package stdlib

import (
	"fmt"
	"strings"
)

// QueryBuilder builds parameterized SQL queries with a fluent API.
type QueryBuilder struct {
	table      string
	conditions []condition
	orderBy    []string
	limit      int
	offset     int
	joins      []string
	selectCols []string
	distinct   bool
	groupBy    []string
	having     []condition
}

// condition holds a single WHERE or HAVING clause.
type condition struct {
	column   string
	operator string
	value    interface{}
	logic    string // AND or OR
}

// NewQueryBuilder creates a query builder for the given table.
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:      table,
		conditions: make([]condition, 0),
		orderBy:    make([]string, 0),
		selectCols: []string{"*"},
	}
}

func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = columns
	return qb
}

func (qb *QueryBuilder) Distinct() *QueryBuilder {
	qb.distinct = true
	return qb
}

func (qb *QueryBuilder) Where(column, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: operator, value: value, logic: "AND"})
	return qb
}

func (qb *QueryBuilder) OrWhere(column, operator string, value interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: operator, value: value, logic: "OR"})
	return qb
}

// WhereIn adds a WHERE column IN (...) condition.
func (qb *QueryBuilder) WhereIn(column string, values []interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: "IN", value: values, logic: "AND"})
	return qb
}

// WhereBetween adds a WHERE column BETWEEN min AND max condition.
func (qb *QueryBuilder) WhereBetween(column string, min, max interface{}) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: "BETWEEN", value: []interface{}{min, max}, logic: "AND"})
	return qb
}

func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: "IS NULL", logic: "AND"})
	return qb
}

func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	qb.conditions = append(qb.conditions, condition{column: column, operator: "IS NOT NULL", logic: "AND"})
	return qb
}

func (qb *QueryBuilder) Join(table, cond string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("JOIN %s ON %s", table, cond))
	return qb
}

func (qb *QueryBuilder) LeftJoin(table, cond string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, cond))
	return qb
}

func (qb *QueryBuilder) RightJoin(table, cond string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("RIGHT JOIN %s ON %s", table, cond))
	return qb
}

func (qb *QueryBuilder) OrderBy(column, direction string) *QueryBuilder {
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

func (qb *QueryBuilder) Having(column, operator string, value interface{}) *QueryBuilder {
	qb.having = append(qb.having, condition{column: column, operator: operator, value: value, logic: "AND"})
	return qb
}

func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Page sets LIMIT and OFFSET from a 1-based page number.
func (qb *QueryBuilder) Page(page, pageSize int) *QueryBuilder {
	qb.limit = pageSize
	qb.offset = (page - 1) * pageSize
	return qb
}

// BuildSelect assembles a SELECT query and returns the SQL string and argument slice.
func (qb *QueryBuilder) BuildSelect() (string, []interface{}) {
	var query strings.Builder
	args := make([]interface{}, 0)

	query.WriteString("SELECT ")
	if qb.distinct {
		query.WriteString("DISTINCT ")
	}
	query.WriteString(strings.Join(qb.selectCols, ", "))
	query.WriteString(" FROM ")
	query.WriteString(qb.table)

	for _, join := range qb.joins {
		query.WriteString(" " + join)
	}

	if len(qb.conditions) > 0 {
		query.WriteString(" WHERE ")
		for i, cond := range qb.conditions {
			if i > 0 {
				query.WriteString(" " + cond.logic + " ")
			}
			query.WriteString(cond.column)
			switch cond.operator {
			case "IN":
				values := cond.value.([]interface{})
				placeholders := make([]string, len(values))
				for j, v := range values {
					placeholders[j] = "?"
					args = append(args, v)
				}
				query.WriteString(fmt.Sprintf(" IN (%s)", strings.Join(placeholders, ",")))
			case "BETWEEN":
				values := cond.value.([]interface{})
				query.WriteString(" BETWEEN ? AND ?")
				args = append(args, values[0], values[1])
			case "IS NULL", "IS NOT NULL":
				query.WriteString(" " + cond.operator)
			default:
				query.WriteString(fmt.Sprintf(" %s ?", cond.operator))
				args = append(args, cond.value)
			}
		}
	}

	if len(qb.groupBy) > 0 {
		query.WriteString(" GROUP BY " + strings.Join(qb.groupBy, ", "))
	}

	if len(qb.having) > 0 {
		query.WriteString(" HAVING ")
		for i, cond := range qb.having {
			if i > 0 {
				query.WriteString(" " + cond.logic + " ")
			}
			query.WriteString(cond.column)
			query.WriteString(fmt.Sprintf(" %s ?", cond.operator))
			args = append(args, cond.value)
		}
	}

	if len(qb.orderBy) > 0 {
		query.WriteString(" ORDER BY " + strings.Join(qb.orderBy, ", "))
	}
	if qb.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", qb.limit))
	}
	if qb.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", qb.offset))
	}

	return query.String(), args
}

// BuildCount wraps BuildSelect as a COUNT(*) query.
func (qb *QueryBuilder) BuildCount() (string, []interface{}) {
	origCols := qb.selectCols
	qb.selectCols = []string{"COUNT(*) as count"}
	query, args := qb.BuildSelect()
	qb.selectCols = origCols
	return query, args
}
