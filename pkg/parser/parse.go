package parser

import (
	"errors"
	"fmt"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

const (
    SELECT string = "SELECT"
    UPDATE  	  = "UPDATE"
    DELETE		  = "DELETE"
)

// Reflect an SQL statement.
type Statement struct {
	Type   string            // Type of statement ex.(SELECT, UPDATE, DELETE)
	Params map[string]string // Statement parameters associated with a source/target table.
}

func stringSlice(list *pg_query.List) []string {
	items := []string{}
	for _, item := range list.Items {
		if n, ok := item.Node.(*pg_query.Node_String_); ok {
			items = append(items, n.String_.Sval)
		}
	}
	return items
}

func stringSliceFromNodes(node []*pg_query.Node) []string {
	var items []string
	for _, item := range node {
		if n, ok := item.Node.(*pg_query.Node_String_); ok {
			items = append(items, n.String_.Sval)
		}
	}
	return items
}

type relation struct {
	Catalog string
	Schema  string
	Name    string
}

func parseRelationFromNodes(list []*pg_query.Node) (*relation, error) {
	parts := stringSliceFromNodes(list)
	switch len(parts) {
	case 1:
		return &relation{
			Name: parts[0],
		}, nil
	case 2:
		return &relation{
			Schema: parts[0],
			Name:   parts[1],
		}, nil
	case 3:
		return &relation{
			Catalog: parts[0],
			Schema:  parts[1],
			Name:    parts[2],
		}, nil
	default:
		return nil, fmt.Errorf("invalid name: %s", joinNodes(list, "."))
	}
}

func parseRelationFromRangeVar(rv *pg_query.RangeVar) *relation {
	return &relation{
		Catalog: rv.Catalogname,
		Schema:  rv.Schemaname,
		Name:    rv.Relname,
	}
}

func parseRelation(in *nodes.Node) (*relation, error) {
	switch n := in.Node.(type) {
	case *nodes.Node_List:
		return parseRelationFromNodes(n.List.Items)
	case *nodes.Node_RangeVar:
		return parseRelationFromRangeVar(n.RangeVar), nil
	case *nodes.Node_TypeName:
		return parseRelationFromNodes(n.TypeName.Names)
	default:
		return nil, fmt.Errorf("unexpected node type: %T", n)
	}
}

func joinNodes(list []*nodes.Node, sep string) string {
	return strings.Join(stringSliceFromNodes(list), sep)
}

var errSkip = errors.New("skip stmt")

func Parse(query string) ([]Statement, error) {
	if query == "" {
		return nil, nil
	}

	query, err := pg_query.Normalize(query)
	if err != nil {
		return nil, fmt.Errorf("Error normalize query: %s, %w", query, err)
	}
	tree, err := pg_query.Parse(query)
	if err != nil {
		pErr := normalizeErr(err)
		return nil, pErr
	}

	var stmts []Statement
	for _, st := range tree.Stmts {
		s, err := convert(st.Stmt)
		if err == errSkip {
			continue
		}
		if err != nil {
			return nil, err
		}
		if s == nil {
			return nil, fmt.Errorf("unexpected nil node")
		}
		stmts = append(stmts, s)
	}
	return stmts, nil
}

