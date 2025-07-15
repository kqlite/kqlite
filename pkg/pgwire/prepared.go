package pgwire

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgerrcode"

	"github.com/kqlite/kqlite/pkg/store"
	"github.com/kqlite/kqlite/pkg/util/pgerror"
)

const (
	// PrepareStatementType represents a prepared statement type.
	PrepareStatementType byte = 'S'
	// PreparePortalType represents a portal message type.
	PreparePortalType byte = 'P'
)

// PreparedPortal is a PreparedStatement that has been bound with query arguments.
type PreparedPortal struct {
	Name     string
	Prepared *PreparedStatement

	// Query arguments.
	Qargs []any
}

// PreparedStatement is a SQL statement that has been parsed and the types
// of arguments and results have been determined.
type PreparedStatement struct {
	Name string
	Stmt *store.Statement

	// Statement param types.
	ParamOIDs []uint32

	// Statement result field types.
	Fields []*sql.ColumnType
}

// addPreparedStmt creates a new PreparedStatement with the provided name, DB statement and statement argument types (OIDs).
// The new prepared statement added is also returned.
// It is illegal to call this when a statement with that name
// already exists (even for anonymous prepared statements).
func (conn *ClientConn) addPreparedStatement(
	name string, stmt *store.Statement, paramOids []uint32,
) (*PreparedStatement, error) {
	if _, ok := conn.prepStmts[name]; ok {
		return nil, pgerror.New(pgerrcode.DuplicatePreparedStatement, fmt.Sprintf("prepared statement %q already exists", name))
	}

	preparedStmt := &PreparedStatement{
		Name:      name,
		Stmt:      stmt,
		ParamOIDs: paramOids,
	}

	// Add statement to connection cache.
	conn.prepStmts[name] = preparedStmt

	return preparedStmt, nil
}

// addPortal creates a new PreparedPortal in the client session cache.
// It is illegal to call this when a portal with that name already exists (even
// for anonymous portals).
func (conn *ClientConn) addPortal(portalName string, pareparedStmt *PreparedStatement, parameterValues []any) error {
	if _, ok := conn.portals[portalName]; ok {
		return fmt.Errorf("portal already exists: %q", portalName)
	}

	portal := &PreparedPortal{
		Name:     portalName,
		Qargs:    parameterValues,
		Prepared: pareparedStmt,
	}

	// Add portal to connection cache.
	conn.portals[portalName] = portal

	return nil
}

func (conn *ClientConn) deletePreparedStmt(name string) {
	_, found := conn.prepStmts[name]
	if !found {
		return
	}
	delete(conn.prepStmts, name)
}

func (conn *ClientConn) deletePortal(portalName string) {
	_, found := conn.portals[portalName]
	if !found {
		return
	}
	delete(conn.portals, portalName)
}
