package parser

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// A Visitor's Visit method is invoked for each node encountered by Walk.
// If the result visitor w is not nil, Walk visits each of the children
// of node with the visitor w, followed by a call of w.Visit(nil).
type Visitor interface {
	Visit(*pg_query.Node) (v Visitor, err error)
	VisitEnd(*pg_query.Node) error
}

func walkList(v Visitor, l *pg_query.List) error {
	for _, item := range l.Items {
		if err := walkNode(v, item); err != nil {
			return err
		}
	}
	return nil
}

func walkSlice(v Visitor, nodes []*pg_query.Node) error {
	for _, n := range nodes {
		if err := walkNode(v, n); err != nil {
			return err
		}
	}
	return nil
}

func walkA_ArrayExpr(v Visitor, n *pg_query.A_ArrayExpr) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Elements); err != nil {
		return err
	}
	return nil
}

func walkA_Expr(v Visitor, n *pg_query.A_Expr) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Name); err != nil {
		return err
	}
	if err := walkNode(v, n.Lexpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Rexpr); err != nil {
		return err
	}
	return nil
}

func walkA_Indices(v Visitor, n *pg_query.A_Indices) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Lidx); err != nil {
		return err
	}
	if err := walkNode(v, n.Uidx); err != nil {
		return err
	}
	return nil
}

func walkA_Indirection(v Visitor, n *pg_query.A_Indirection) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	if err := walkSlice(v, n.Indirection); err != nil {
		return err
	}
	return nil
}

func walkAccessPriv(v Visitor, n *pg_query.AccessPriv) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Cols); err != nil {
		return err
	}
	return nil
}

func walkAggref(v Visitor, n *pg_query.Aggref) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Aggargtypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Aggdirectargs); err != nil {
		return nil
	}
	if err := walkSlice(v, n.Args); err != nil {
		return nil
	}
	if err := walkSlice(v, n.Aggorder); err != nil {
		return nil
	}
	if err := walkSlice(v, n.Aggdistinct); err != nil {
		return nil
	}
	if err := walkNode(v, n.Aggfilter); err != nil {
		return nil
	}
	return nil
}

func walkAlias(v Visitor, n *pg_query.Alias) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Colnames)
}

func walkAlterCollationStmt(v Visitor, n *pg_query.AlterCollationStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Collname); err != nil {
		return err
	}
	return nil
}

func walkAlterDatabaseSetStmt(v Visitor, n *pg_query.AlterDatabaseSetStmt) error {
	if n == nil {
		return nil
	}
	return walkVariableSetStmt(v, n.Setstmt)
}

func walkAlterDatabaseStmt(v Visitor, n *pg_query.AlterDatabaseStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkAlterDefaultPrivilegesStmt(v Visitor, n *pg_query.AlterDefaultPrivilegesStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	if err := walkGrantStmt(v, n.Action); err != nil {
		return err
	}
	return nil
}

func walkAlterDomainStmt(v Visitor, n *pg_query.AlterDomainStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.TypeName); err != nil {
		return err
	}
	if err := walkNode(v, n.Def); err != nil {
		return err
	}
	return nil
}

func walkAlterEnumStmt(v Visitor, n *pg_query.AlterEnumStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.TypeName)
}

func walkAlterExtensionContentsStmt(v Visitor, n *pg_query.AlterExtensionContentsStmt) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Object)
}

func walkAlterExtensionStmt(v Visitor, n *pg_query.AlterExtensionStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkAlterFdwStmt(v Visitor, n *pg_query.AlterFdwStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.FuncOptions); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkAlterForeignServerStmt(v Visitor, n *pg_query.AlterForeignServerStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkAlterFunctionStmt(v Visitor, n *pg_query.AlterFunctionStmt) error {
	if n == nil {
		return nil
	}

	if err := walkObjectWithArgs(v, n.Func); err != nil {
		return err
	}
	if err := walkSlice(v, n.Actions); err != nil {
		return err
	}
	return nil
}

func walkAlterObjectDependsStmt(v Visitor, n *pg_query.AlterObjectDependsStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkNode(v, n.Object); err != nil {
		return err
	}
	return nil
}

func walkAlterObjectSchemaStmt(v Visitor, n *pg_query.AlterObjectSchemaStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkNode(v, n.Object); err != nil {
		return err
	}
	return nil
}

func walkAlterOpFamilyStmt(v Visitor, n *pg_query.AlterOpFamilyStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Opfamilyname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Items); err != nil {
		return err
	}
	return nil
}

func walkAlterOperatorStmt(v Visitor, n *pg_query.AlterOperatorStmt) error {
	if n == nil {
		return nil
	}

	if err := walkObjectWithArgs(v, n.Opername); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkAlterSeqStmt(v Visitor, n *pg_query.AlterSeqStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Sequence); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkAlterSubscriptionStmt(v Visitor, n *pg_query.AlterSubscriptionStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Publication); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkAlterSystemStmt(v Visitor, n *pg_query.AlterSystemStmt) error {
	if n == nil {
		return nil
	}
	return walkVariableSetStmt(v, n.Setstmt)
}

func walkAlterTSConfigurationStmt(v Visitor, n *pg_query.AlterTSConfigurationStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Cfgname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Tokentype); err != nil {
		return err
	}
	if err := walkSlice(v, n.Dicts); err != nil {
		return err
	}
	return nil
}

func walkAlterTSDictionaryStmt(v Visitor, n *pg_query.AlterTSDictionaryStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Dictname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkAlterTableCmd(v Visitor, n *pg_query.AlterTableCmd) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Def)
}

func walkAlterTableSpaceOptionsStmt(v Visitor, n *pg_query.AlterTableSpaceOptionsStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkAlterTableStmt(v Visitor, n *pg_query.AlterTableStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.Cmds); err != nil {
		return err
	}
	return nil
}

func walkAlternativeSubPlan(v Visitor, n *pg_query.AlternativeSubPlan) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Subplans); err != nil {
		return err
	}
	return nil
}

func walkArrayCoerceExpr(v Visitor, n *pg_query.ArrayCoerceExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkArrayExpr(v Visitor, n *pg_query.ArrayExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Elements); err != nil {
		return err
	}
	return nil
}

func walkBoolExpr(v Visitor, n *pg_query.BoolExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkCallStmt(v Visitor, n *pg_query.CallStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Funccall.Funcname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccall.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccall.AggOrder); err != nil {
		return err
	}
	if err := walkNode(v, n.Funccall.AggFilter); err != nil {
		return err
	}
	if err := walkWindowDef(v, n.Funccall.Over); err != nil {
		return err
	}
	return nil
}

func walkCaseExpr(v Visitor, n *pg_query.CaseExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkNode(v, n.Defresult); err != nil {
		return err
	}
	return nil
}

func walkCaseTestExpr(v Visitor, n *pg_query.CaseTestExpr) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkCaseWhen(v Visitor, n *pg_query.CaseWhen) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Expr); err != nil {
		return err
	}
	if err := walkNode(v, n.Result); err != nil {
		return err
	}
	return nil
}

func walkClusterStmt(v Visitor, n *pg_query.ClusterStmt) error {
	if n == nil {
		return nil
	}
	return walkRangeVar(v, n.Relation)
}

func walkCoalesceExpr(v Visitor, n *pg_query.CoalesceExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkCoerceToDomain(v Visitor, n *pg_query.CoerceToDomain) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkCoerceToDomainValue(v Visitor, n *pg_query.CoerceToDomainValue) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkCoerceViaIO(v Visitor, n *pg_query.CoerceViaIO) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkCollateClause(v Visitor, n *pg_query.CollateClause) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	if err := walkSlice(v, n.Collname); err != nil {
		return err
	}
	return nil
}

func walkCollateExpr(v Visitor, n *pg_query.CollateExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkColumnDef(v Visitor, n *pg_query.ColumnDef) error {
	if n == nil {
		return nil
	}

	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	if err := walkNode(v, n.RawDefault); err != nil {
		return err
	}
	if err := walkNode(v, n.CookedDefault); err != nil {
		return err
	}
	if err := walkCollateClause(v, n.CollClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.Constraints); err != nil {
		return err
	}
	if err := walkSlice(v, n.Fdwoptions); err != nil {
		return err
	}
	return nil
}

func walkColumnRef(v Visitor, n *pg_query.ColumnRef) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Fields)
}

func walkCommentStmt(v Visitor, n *pg_query.CommentStmt) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Object)
}

func walkCommonTableExpr(v Visitor, n *pg_query.CommonTableExpr) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Aliascolnames); err != nil {
		return err
	}
	if err := walkNode(v, n.Ctequery); err != nil {
		return err
	}
	if err := walkSlice(v, n.Ctecolnames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Ctecoltypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Ctecoltypmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.Ctecolcollations); err != nil {
		return err
	}
	return nil
}

func walkConstraint(v Visitor, n *pg_query.Constraint) error {
	if n == nil {
		return nil
	}
	if err := walkNode(v, n.RawExpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Keys); err != nil {
		return err
	}
	if err := walkSlice(v, n.Exclusions); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkRangeVar(v, n.Pktable); err != nil {
		return err
	}
	if err := walkSlice(v, n.FkAttrs); err != nil {
		return err
	}
	if err := walkSlice(v, n.PkAttrs); err != nil {
		return err
	}
	if err := walkSlice(v, n.OldConpfeqop); err != nil {
		return err
	}
	return nil
}

func walkConstraintsSetStmt(v Visitor, n *pg_query.ConstraintsSetStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Constraints)
}

func walkRowtypeExpr(v Visitor, n *pg_query.ConvertRowtypeExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkCopyStmt(v Visitor, n *pg_query.CopyStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkNode(v, n.Query); err != nil {
		return err
	}
	if err := walkSlice(v, n.Attlist); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkCreateAmStmt(v Visitor, n *pg_query.CreateAmStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.HandlerName)
}

func walkCreateCastStmt(v Visitor, n *pg_query.CreateCastStmt) error {
	if n == nil {
		return nil
	}

	if err := walkTypeName(v, n.Sourcetype); err != nil {
		return err
	}
	if err := walkTypeName(v, n.Targettype); err != nil {
		return err
	}
	if err := walkObjectWithArgs(v, n.Func); err != nil {
		return err
	}
	return nil
}

func walkCreateConversionStmt(v Visitor, n *pg_query.CreateConversionStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.ConversionName); err != nil {
		return err
	}
	if err := walkSlice(v, n.FuncName); err != nil {
		return err
	}
	return nil
}

func walkCreateDomainStmt(v Visitor, n *pg_query.CreateDomainStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Domainname); err != nil {
		return err
	}
	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	if err := walkCollateClause(v, n.CollClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.Constraints); err != nil {
		return err
	}
	return nil
}

func walkCreateEnumStmt(v Visitor, n *pg_query.CreateEnumStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Vals)
}

func walkCreateEventTrigStmt(v Visitor, n *pg_query.CreateEventTrigStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Whenclause); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funcname); err != nil {
		return err
	}
	return nil
}

func walkCreateExtensionStmt(v Visitor, n *pg_query.CreateExtensionStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)

}

func walkCreateFdwStmt(v Visitor, n *pg_query.CreateFdwStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.FuncOptions); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkCreateForeignTableStmt(v Visitor, n *pg_query.CreateForeignTableStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkCreateFunctionStmt(v Visitor, n *pg_query.CreateFunctionStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Parameters); err != nil {
		return err
	}
	if err := walkTypeName(v, n.ReturnType); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkCreateOpClassItem(v Visitor, n *pg_query.CreateOpClassItem) error {
	if n == nil {
		return nil
	}

	if err := walkObjectWithArgs(v, n.Name); err != nil {
		return err
	}
	if err := walkSlice(v, n.OrderFamily); err != nil {
		return err
	}
	if err := walkSlice(v, n.ClassArgs); err != nil {
		return err
	}
	if err := walkTypeName(v, n.Storedtype); err != nil {
		return err
	}
	return nil
}

func walkCreateOpClassStmt(v Visitor, n *pg_query.CreateOpClassStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Opclassname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Opfamilyname); err != nil {
		return err
	}
	if err := walkTypeName(v, n.Datatype); err != nil {
		return err
	}
	if err := walkSlice(v, n.Items); err != nil {
		return err
	}
	return nil
}

func walkCreateOpFamilyStmt(v Visitor, n *pg_query.CreateOpFamilyStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Opfamilyname)
}

func walkCreatePLangStmt(v Visitor, n *pg_query.CreatePLangStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Plhandler); err != nil {
		return err
	}
	if err := walkSlice(v, n.Plinline); err != nil {
		return err
	}
	if err := walkSlice(v, n.Plvalidator); err != nil {
		return err
	}
	return nil
}

func walkCreatePublicationStmt(v Visitor, n *pg_query.CreatePublicationStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	if err := walkSlice(v, n.Pubobjects); err != nil {
		return err
	}
	return nil
}

func walkCreateRangeStmt(v Visitor, n *pg_query.CreateRangeStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.TypeName); err != nil {
		return err
	}
	if err := walkSlice(v, n.Params); err != nil {
		return err
	}
	return nil
}

func walkCreateStatsStmt(v Visitor, n *pg_query.CreateStatsStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Defnames); err != nil {
		return err
	}
	if err := walkSlice(v, n.StatTypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Exprs); err != nil {
		return err
	}
	if err := walkSlice(v, n.Relations); err != nil {
		return err
	}
	return nil
}

func walkCreateStmt(v Visitor, n *pg_query.CreateStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.TableElts); err != nil {
		return err
	}
	if err := walkSlice(v, n.InhRelations); err != nil {
		return err
	}
	if err := walkPartitionBoundSpec(v, n.Partbound); err != nil {
		return err
	}
	if err := walkPartitionSpec(v, n.Partspec); err != nil {
		return err
	}
	if err := walkTypeName(v, n.OfTypename); err != nil {
		return err
	}
	if err := walkSlice(v, n.Constraints); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkCreateSubscriptionStmt(v Visitor, n *pg_query.CreateSubscriptionStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Publication); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkCreateTableAsStmt(v Visitor, n *pg_query.CreateTableAsStmt) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Query); err != nil {
		return err
	}
	if err := walkIntoClause(v, n.Into); err != nil {
		return err
	}
	return nil
}

func walkCreateTransformStmt(v Visitor, n *pg_query.CreateTransformStmt) error {
	if n == nil {
		return nil
	}

	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	if err := walkObjectWithArgs(v, n.Fromsql); err != nil {
		return err
	}
	if err := walkObjectWithArgs(v, n.Tosql); err != nil {
		return err
	}
	return nil
}

func walkCreateTrigStmt(v Visitor, n *pg_query.CreateTrigStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Funcname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.Columns); err != nil {
		return err
	}
	if err := walkNode(v, n.WhenClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.TransitionRels); err != nil {
		return err
	}
	return nil
}

func walkCreatedbStmt(v Visitor, n *pg_query.CreatedbStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkCurrentOfExpr(v Visitor, n *pg_query.CurrentOfExpr) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkDeclareCursorStmt(v Visitor, n *pg_query.DeclareCursorStmt) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Query)
}

func walkDefElem(v Visitor, n *pg_query.DefElem) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Arg)
}

func walkDefineStmt(v Visitor, n *pg_query.DefineStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Defnames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.Definition); err != nil {
		return err
	}
	return nil
}

func walkDeleteStmt(v Visitor, n *pg_query.DeleteStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.UsingClause); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.ReturningList); err != nil {
		return err
	}
	if err := walkWithClause(v, n.WithClause); err != nil {
		return err
	}
	return nil
}

func walkDoStmt(v Visitor, n *pg_query.DoStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Args)
}

func walkDropStmt(v Visitor, n *pg_query.DropStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Objects)
}

func walkExecuteStmt(v Visitor, n *pg_query.ExecuteStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Params)
}

func walkExplainStmt(v Visitor, n *pg_query.ExplainStmt) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Query); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkFieldSelect(v Visitor, n *pg_query.FieldSelect) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkFieldStore(v Visitor, n *pg_query.FieldStore) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	if err := walkSlice(v, n.Newvals); err != nil {
		return err
	}
	if err := walkSlice(v, n.Fieldnums); err != nil {
		return err
	}
	return nil
}

func walkFromExpr(v Visitor, n *pg_query.FromExpr) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Fromlist); err != nil {
		return err
	}
	if err := walkNode(v, n.Quals); err != nil {
		return err
	}
	return nil
}

func walkFuncCall(v Visitor, n *pg_query.FuncCall) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Funcname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.AggOrder); err != nil {
		return err
	}
	if err := walkNode(v, n.AggFilter); err != nil {
		return err
	}
	if err := walkWindowDef(v, n.Over); err != nil {
		return err
	}
	return nil
}

func walkFuncExpr(v Visitor, n *pg_query.FuncExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkFunctionParameter(v Visitor, n *pg_query.FunctionParameter) error {
	if n == nil {
		return nil
	}

	if err := walkTypeName(v, n.ArgType); err != nil {
		return err
	}
	if err := walkNode(v, n.Defexpr); err != nil {
		return err
	}
	return nil
}

func walkGrantStmt(v Visitor, n *pg_query.GrantStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Objects); err != nil {
		return err
	}
	if err := walkSlice(v, n.Privileges); err != nil {
		return err
	}
	if err := walkSlice(v, n.Grantees); err != nil {
		return err
	}
	return nil
}

func walkGroupingFunc(v Visitor, n *pg_query.GroupingFunc) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.Refs); err != nil {
		return err
	}
	return nil
}

func walkGroupingSet(v Visitor, n *pg_query.GroupingSet) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Content)
}

func walkIndexElem(v Visitor, n *pg_query.IndexElem) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Expr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Collation); err != nil {
		return err
	}
	if err := walkSlice(v, n.Opclass); err != nil {
		return err
	}
	return nil
}

func walkIndexStmt(v Visitor, n *pg_query.IndexStmt) error {
	if n == nil {
		return nil
	}
	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.IndexParams); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.ExcludeOpNames); err != nil {
		return err
	}
	return nil
}

func walkInferClause(v Visitor, n *pg_query.InferClause) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.IndexElems); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	return nil
}

func walkInferenceElem(v Visitor, n *pg_query.InferenceElem) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Expr); err != nil {
		return err
	}
	return nil

}

func walkInsertStmt(v Visitor, n *pg_query.InsertStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.Cols); err != nil {
		return err
	}
	if err := walkNode(v, n.SelectStmt); err != nil {
		return err
	}
	if err := walkOnConflictClause(v, n.OnConflictClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.ReturningList); err != nil {
		return err
	}
	if err := walkWithClause(v, n.WithClause); err != nil {
		return err
	}
	return nil
}

func walkIntoClause(v Visitor, n *pg_query.IntoClause) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Rel); err != nil {
		return err
	}
	if err := walkSlice(v, n.ColNames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	if err := walkNode(v, n.ViewQuery); err != nil {
		return err
	}
	return nil

}

func walkJoinExpr(v Visitor, n *pg_query.JoinExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Larg); err != nil {
		return err
	}
	if err := walkNode(v, n.Rarg); err != nil {
		return err
	}
	if err := walkSlice(v, n.UsingClause); err != nil {
		return err
	}
	if err := walkNode(v, n.Quals); err != nil {
		return err
	}
	if err := walkAlias(v, n.Alias); err != nil {
		return err
	}
	return nil
}

func walkLockStmt(v Visitor, n *pg_query.LockStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Relations)
}

func walkLockingClause(v Visitor, n *pg_query.LockingClause) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.LockedRels)
}

func walkMinMaxExpr(v Visitor, n *pg_query.MinMaxExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkMultiAssignRef(v Visitor, n *pg_query.MultiAssignRef) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Source)
}

func walkNamedArgExpr(v Visitor, n *pg_query.NamedArgExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkNextValueExpr(v Visitor, n *pg_query.NextValueExpr) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkNullTest(v Visitor, n *pg_query.NullTest) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkObjectWithArgs(v Visitor, n *pg_query.ObjectWithArgs) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Objname); err != nil {
		return err
	}
	if err := walkSlice(v, n.Objargs); err != nil {
		return err
	}
	return nil
}

func walkOnConflictClause(v Visitor, n *pg_query.OnConflictClause) error {
	if n == nil {
		return nil
	}
	if err := walkInferClause(v, n.Infer); err != nil {
		return err
	}
	if err := walkSlice(v, n.TargetList); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	return nil
}

func walkOnConflictExpr(v Visitor, n *pg_query.OnConflictExpr) error {
	if n == nil {
		return nil
	}
	if err := walkSlice(v, n.ArbiterElems); err != nil {
		return err
	}
	if err := walkNode(v, n.ArbiterWhere); err != nil {
		return err
	}
	if err := walkSlice(v, n.OnConflictSet); err != nil {
		return err
	}
	if err := walkNode(v, n.OnConflictWhere); err != nil {
		return err
	}
	if err := walkSlice(v, n.ExclRelTlist); err != nil {
		return err
	}
	return nil
}

func walkOpExpr(v Visitor, n *pg_query.OpExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil

}

func walkParam(v Visitor, n *pg_query.Param) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)

}

func walkPartitionBoundSpec(v Visitor, n *pg_query.PartitionBoundSpec) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Listdatums); err != nil {
		return err
	}
	if err := walkSlice(v, n.Lowerdatums); err != nil {
		return err
	}
	if err := walkSlice(v, n.Upperdatums); err != nil {
		return err
	}
	return nil
}

func walkPartitionSpec(v Visitor, n *pg_query.PartitionSpec) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.PartParams)
}

func walkPrepareStmt(v Visitor, n *pg_query.PrepareStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Argtypes); err != nil {
		return err
	}
	if err := walkNode(v, n.Query); err != nil {
		return err
	}
	return nil
}

func walkQuery(v Visitor, n *pg_query.Query) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.UtilityStmt); err != nil {
		return err
	}
	if err := walkSlice(v, n.CteList); err != nil {
		return err
	}
	if err := walkSlice(v, n.Rtable); err != nil {
		return err
	}
	if err := walkFromExpr(v, n.Jointree); err != nil {
		return err
	}
	if err := walkSlice(v, n.TargetList); err != nil {
		return err
	}
	if err := walkOnConflictExpr(v, n.OnConflict); err != nil {
		return err
	}
	if err := walkSlice(v, n.ReturningList); err != nil {
		return err
	}
	if err := walkSlice(v, n.GroupClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.GroupingSets); err != nil {
		return err
	}
	if err := walkNode(v, n.HavingQual); err != nil {
		return err
	}
	if err := walkSlice(v, n.WindowClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.DistinctClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.SortClause); err != nil {
		return err
	}
	if err := walkNode(v, n.LimitOffset); err != nil {
		return err
	}
	if err := walkNode(v, n.LimitCount); err != nil {
		return err
	}
	if err := walkSlice(v, n.RowMarks); err != nil {
		return err
	}
	if err := walkNode(v, n.SetOperations); err != nil {
		return err
	}
	if err := walkSlice(v, n.ConstraintDeps); err != nil {
		return err
	}
	if err := walkSlice(v, n.WithCheckOptions); err != nil {
		return err
	}
	return nil
}

func walkRangeFunction(v Visitor, n *pg_query.RangeFunction) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Functions); err != nil {
		return err
	}
	if err := walkAlias(v, n.Alias); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coldeflist); err != nil {
		return err
	}
	return nil
}

func walkRangeSubselect(v Visitor, n *pg_query.RangeSubselect) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Subquery); err != nil {
		return err
	}
	if err := walkAlias(v, n.Alias); err != nil {
		return err
	}
	return nil
}

func walkRangeTableFunc(v Visitor, n *pg_query.RangeTableFunc) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Docexpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Rowexpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Namespaces); err != nil {
		return err
	}
	if err := walkSlice(v, n.Columns); err != nil {
		return err
	}
	if err := walkAlias(v, n.Alias); err != nil {
		return err
	}
	return nil
}

func walkRangeTableFuncCol(v Visitor, n *pg_query.RangeTableFuncCol) error {
	if n == nil {
		return nil
	}

	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	if err := walkNode(v, n.Colexpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Coldefexpr); err != nil {
		return err
	}
	return nil
}

func walkRangeTableSample(v Visitor, n *pg_query.RangeTableSample) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.Method); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkNode(v, n.Repeatable); err != nil {
		return err
	}
	return nil
}

func walkRangeTblEntry(v Visitor, n *pg_query.RangeTblEntry) error {
	if n == nil {
		return nil
	}
	if err := walkTableSampleClause(v, n.Tablesample); err != nil {
		return err
	}
	if err := walkQuery(v, n.Subquery); err != nil {
		return err
	}
	if err := walkSlice(v, n.Joinaliasvars); err != nil {
		return err
	}
	if err := walkSlice(v, n.Functions); err != nil {
		return err
	}
	if err := walkTableFunc(v, n.Tablefunc); err != nil {
		return err
	}
	if err := walkSlice(v, n.ValuesLists); err != nil {
		return err
	}
	if err := walkSlice(v, n.Joinaliasvars); err != nil {
		return err
	}
	if err := walkSlice(v, n.Functions); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coltypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coltypmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.Colcollations); err != nil {
		return err
	}
	if err := walkAlias(v, n.Alias); err != nil {
		return err
	}
	if err := walkAlias(v, n.Eref); err != nil {
		return err
	}
	if err := walkSlice(v, n.SecurityQuals); err != nil {
		return err
	}
	return nil
}

func walkRangeTblFunction(v Visitor, n *pg_query.RangeTblFunction) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Funcexpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccolnames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccoltypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccoltypmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.Funccolcollations); err != nil {
		return err
	}
	return nil
}

func walkRawStmt(v Visitor, n *pg_query.RawStmt) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Stmt)
}

func walkRefreshMatViewStmt(v Visitor, n *pg_query.RefreshMatViewStmt) error {
	if n == nil {
		return nil
	}
	return walkRangeVar(v, n.Relation)
}

func walkReindexStmt(v Visitor, n *pg_query.ReindexStmt) error {
	if n == nil {
		return nil
	}
	return walkRangeVar(v, n.Relation)
}

func walkRelabelType(v Visitor, n *pg_query.RelabelType) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	return nil
}

func walkRenameStmt(v Visitor, n *pg_query.RenameStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkNode(v, n.Object); err != nil {
		return err
	}
	return nil
}

func walkResTarget(v Visitor, n *pg_query.ResTarget) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Indirection); err != nil {
		return err
	}
	if err := walkNode(v, n.Val); err != nil {
		return err
	}
	return nil
}

func walkRowCompareExpr(v Visitor, n *pg_query.RowCompareExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Opnos); err != nil {
		return err
	}
	if err := walkSlice(v, n.Opfamilies); err != nil {
		return err
	}
	if err := walkSlice(v, n.Inputcollids); err != nil {
		return err
	}
	if err := walkSlice(v, n.Largs); err != nil {
		return err
	}
	if err := walkSlice(v, n.Rargs); err != nil {
		return err
	}
	return nil
}

func walkRowExpr(v Visitor, n *pg_query.RowExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkSlice(v, n.Colnames); err != nil {
		return err
	}
	return nil
}

func walkRuleStmt(v Visitor, n *pg_query.RuleStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.Actions); err != nil {
		return err
	}
	return nil
}

func walkSQLValueFunction(v Visitor, n *pg_query.SQLValueFunction) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkScalarArrayOpExpr(v Visitor, n *pg_query.ScalarArrayOpExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkSecLabelStmt(v Visitor, n *pg_query.SecLabelStmt) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Object)
}

func walkSelectStmt(v Visitor, n *pg_query.SelectStmt) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.DistinctClause); err != nil {
		return err
	}
	if err := walkIntoClause(v, n.IntoClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.TargetList); err != nil {
		return err
	}
	if err := walkSlice(v, n.FromClause); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.GroupClause); err != nil {
		return err
	}
	if err := walkNode(v, n.HavingClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.WindowClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.ValuesLists); err != nil {
		return err
	}
	if err := walkSlice(v, n.SortClause); err != nil {
		return err
	}
	if err := walkNode(v, n.LimitOffset); err != nil {
		return err
	}
	if err := walkNode(v, n.LimitCount); err != nil {
		return err
	}
	if err := walkSlice(v, n.LockingClause); err != nil {
		return err
	}
	if err := walkWithClause(v, n.WithClause); err != nil {
		return err
	}
	if err := walkSelectStmt(v, n.Larg); err != nil {
		return err
	}
	if err := walkSelectStmt(v, n.Rarg); err != nil {
		return err
	}
	return nil
}

func walkSetOperationStmt(v Visitor, n *pg_query.SetOperationStmt) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Larg); err != nil {
		return err
	}
	if err := walkNode(v, n.Rarg); err != nil {
		return err
	}
	if err := walkSlice(v, n.ColTypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.ColTypmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.ColCollations); err != nil {
		return err
	}
	if err := walkSlice(v, n.GroupClauses); err != nil {
		return err
	}
	return nil
}

func walkSetToDefault(v Visitor, n *pg_query.SetToDefault) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkSortBy(v Visitor, n *pg_query.SortBy) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Node); err != nil {
		return err
	}
	if err := walkSlice(v, n.UseOp); err != nil {
		return err
	}
	return nil

}

func walkSubLink(v Visitor, n *pg_query.SubLink) error {
	if n == nil {
		return nil
	}
	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Testexpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.OperName); err != nil {
		return err
	}
	if err := walkNode(v, n.Subselect); err != nil {
		return err
	}
	return nil
}

func walkSubPlan(v Visitor, n *pg_query.SubPlan) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Testexpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.ParamIds); err != nil {
		return err
	}
	if err := walkSlice(v, n.SetParam); err != nil {
		return err
	}
	if err := walkSlice(v, n.ParParam); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkTableFunc(v Visitor, n *pg_query.TableFunc) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.NsUris); err != nil {
		return err
	}
	if err := walkSlice(v, n.NsNames); err != nil {
		return err
	}
	if err := walkNode(v, n.Docexpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Rowexpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Colnames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coltypes); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coltypmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.Colcollations); err != nil {
		return err
	}
	if err := walkSlice(v, n.Colexprs); err != nil {
		return err
	}
	if err := walkSlice(v, n.Coldefexprs); err != nil {
		return err
	}
	return nil
}

func walkTableLikeClause(v Visitor, n *pg_query.TableLikeClause) error {
	if n == nil {
		return nil
	}
	return walkRangeVar(v, n.Relation)
}

func walkTableSampleClause(v Visitor, n *pg_query.TableSampleClause) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkNode(v, n.Repeatable); err != nil {
		return err
	}
	return nil
}

func walkTargetEntry(v Visitor, n *pg_query.TargetEntry) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkNode(v, n.Expr); err != nil {
		return err
	}
	return nil
}

func walkTransactionStmt(v Visitor, n *pg_query.TransactionStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Options)
}

func walkTruncateStmt(v Visitor, n *pg_query.TruncateStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Relations)
}

func walkTypeCast(v Visitor, n *pg_query.TypeCast) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Arg); err != nil {
		return err
	}
	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	return nil
}

func walkTypeName(v Visitor, n *pg_query.TypeName) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.Names); err != nil {
		return err
	}
	if err := walkSlice(v, n.Typmods); err != nil {
		return err
	}
	if err := walkSlice(v, n.ArrayBounds); err != nil {
		return err
	}
	return nil
}

func walkRangeVar(v Visitor, n *pg_query.RangeVar) error {
	if n == nil {
		return nil
	}
	return walkAlias(v, n.Alias)
}

func walkUpdateStmt(v Visitor, n *pg_query.UpdateStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.Relation); err != nil {
		return err
	}
	if err := walkSlice(v, n.TargetList); err != nil {
		return err
	}
	if err := walkNode(v, n.WhereClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.FromClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.ReturningList); err != nil {
		return err
	}
	if err := walkWithClause(v, n.WithClause); err != nil {
		return err
	}
	return nil
}

func walkVar(v Visitor, n *pg_query.Var) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Xpr)
}

func walkVariableSetStmt(v Visitor, n *pg_query.VariableSetStmt) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Args)
}

func walkViewStmt(v Visitor, n *pg_query.ViewStmt) error {
	if n == nil {
		return nil
	}

	if err := walkRangeVar(v, n.View); err != nil {
		return err
	}
	if err := walkSlice(v, n.Aliases); err != nil {
		return err
	}
	if err := walkNode(v, n.Query); err != nil {
		return err
	}
	if err := walkSlice(v, n.Options); err != nil {
		return err
	}
	return nil
}

func walkWindowClause(v Visitor, n *pg_query.WindowClause) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.PartitionClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.OrderClause); err != nil {
		return err
	}
	if err := walkNode(v, n.StartOffset); err != nil {
		return err
	}
	if err := walkNode(v, n.EndOffset); err != nil {
		return err
	}
	return nil
}

func walkWindowDef(v Visitor, n *pg_query.WindowDef) error {
	if n == nil {
		return nil
	}

	if err := walkSlice(v, n.PartitionClause); err != nil {
		return err
	}
	if err := walkSlice(v, n.OrderClause); err != nil {
		return err
	}
	if err := walkNode(v, n.StartOffset); err != nil {
		return err
	}
	if err := walkNode(v, n.EndOffset); err != nil {
		return err
	}
	return nil
}

func walkWindowFunc(v Visitor, n *pg_query.WindowFunc) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	if err := walkNode(v, n.Aggfilter); err != nil {
		return err
	}
	return nil
}

func walkWithCheckOption(v Visitor, n *pg_query.WithCheckOption) error {
	if n == nil {
		return nil
	}
	return walkNode(v, n.Qual)
}

func walkWithClause(v Visitor, n *pg_query.WithClause) error {
	if n == nil {
		return nil
	}
	return walkSlice(v, n.Ctes)
}

func walkXmlExpr(v Visitor, n *pg_query.XmlExpr) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Xpr); err != nil {
		return err
	}
	if err := walkSlice(v, n.NamedArgs); err != nil {
		return err
	}
	if err := walkSlice(v, n.ArgNames); err != nil {
		return err
	}
	if err := walkSlice(v, n.Args); err != nil {
		return err
	}
	return nil
}

func walkXmlSerialize(v Visitor, n *pg_query.XmlSerialize) error {
	if n == nil {
		return nil
	}

	if err := walkNode(v, n.Expr); err != nil {
		return err
	}
	if err := walkTypeName(v, n.TypeName); err != nil {
		return err
	}
	return nil
}

func Walk(v Visitor, node *pg_query.Node) error {
	if v != nil {
		return walkNode(v, node)
	}
	return nil
}

func walkNode(v Visitor, node *pg_query.Node) error {
	if node == nil || node.Node == nil {
		return fmt.Errorf("Invalid input node")
	}

	// Visit the node itself
	if v, err := v.Visit(node); err != nil {
		return err
	} else if v == nil {
		return nil
	}

	switch n := node.Node.(type) {
	case *pg_query.Node_AArrayExpr:
		return walkA_ArrayExpr(v, n.AArrayExpr)

	case *pg_query.Node_AConst:
		// Handled by Visitor.
		return nil

	case *pg_query.Node_AExpr:
		return walkA_Expr(v, n.AExpr)

	case *pg_query.Node_AIndices:
		return walkA_Indices(v, n.AIndices)

	case *pg_query.Node_AIndirection:
		return walkA_Indirection(v, n.AIndirection)

	case *pg_query.Node_AccessPriv:
		return walkAccessPriv(v, n.AccessPriv)

	case *pg_query.Node_Aggref:
		return walkAggref(v, n.Aggref)

	case *pg_query.Node_Alias:
		return walkAlias(v, n.Alias)

	case *pg_query.Node_AlterCollationStmt:
		return walkAlterCollationStmt(v, n.AlterCollationStmt)

	case *pg_query.Node_AlterDatabaseSetStmt:
		return walkAlterDatabaseSetStmt(v, n.AlterDatabaseSetStmt)

	case *pg_query.Node_AlterDatabaseStmt:
		return walkAlterDatabaseStmt(v, n.AlterDatabaseStmt)

	case *pg_query.Node_AlterDefaultPrivilegesStmt:
		return walkAlterDefaultPrivilegesStmt(v, n.AlterDefaultPrivilegesStmt)

	case *pg_query.Node_AlterDomainStmt:
		return walkAlterDomainStmt(v, n.AlterDomainStmt)

	case *pg_query.Node_AlterEnumStmt:
		return walkAlterEnumStmt(v, n.AlterEnumStmt)

	case *pg_query.Node_AlterExtensionContentsStmt:
		return walkAlterExtensionContentsStmt(v, n.AlterExtensionContentsStmt)

	case *pg_query.Node_AlterExtensionStmt:
		return walkAlterExtensionStmt(v, n.AlterExtensionStmt)

	case *pg_query.Node_AlterFdwStmt:
		return walkAlterFdwStmt(v, n.AlterFdwStmt)

	case *pg_query.Node_AlterForeignServerStmt:
		return walkAlterForeignServerStmt(v, n.AlterForeignServerStmt)

	case *pg_query.Node_AlterFunctionStmt:
		return walkAlterFunctionStmt(v, n.AlterFunctionStmt)

	case *pg_query.Node_AlterObjectDependsStmt:
		return walkAlterObjectDependsStmt(v, n.AlterObjectDependsStmt)

	case *pg_query.Node_AlterObjectSchemaStmt:
		return walkAlterObjectSchemaStmt(v, n.AlterObjectSchemaStmt)

	case *pg_query.Node_AlterOpFamilyStmt:
		return walkAlterOpFamilyStmt(v, n.AlterOpFamilyStmt)

	case *pg_query.Node_AlterOperatorStmt:
		return walkAlterOperatorStmt(v, n.AlterOperatorStmt)

	case *pg_query.Node_AlterSeqStmt:
		return walkAlterSeqStmt(v, n.AlterSeqStmt)

	case *pg_query.Node_AlterSubscriptionStmt:
		return walkAlterSubscriptionStmt(v, n.AlterSubscriptionStmt)

	case *pg_query.Node_AlterSystemStmt:
		return walkAlterSystemStmt(v, n.AlterSystemStmt)

	case *pg_query.Node_AlterTsconfigurationStmt:
		return walkAlterTSConfigurationStmt(v, n.AlterTsconfigurationStmt)

	case *pg_query.Node_AlterTsdictionaryStmt:
		return walkAlterTSDictionaryStmt(v, n.AlterTsdictionaryStmt)

	case *pg_query.Node_AlterTableCmd:
		return walkAlterTableCmd(v, n.AlterTableCmd)

	case *pg_query.Node_AlterTableStmt:
		return walkAlterTableStmt(v, n.AlterTableStmt)

	case *pg_query.Node_AlternativeSubPlan:
		return walkAlternativeSubPlan(v, n.AlternativeSubPlan)

	case *pg_query.Node_ArrayCoerceExpr:
		return walkArrayCoerceExpr(v, n.ArrayCoerceExpr)

	case *pg_query.Node_ArrayExpr:
		return walkArrayExpr(v, n.ArrayExpr)

	case *pg_query.Node_BoolExpr:
		return walkBoolExpr(v, n.BoolExpr)

	case *pg_query.Node_CallStmt:
		return walkCallStmt(v, n.CallStmt)

	case *pg_query.Node_CaseExpr:
		return walkCaseExpr(v, n.CaseExpr)

	case *pg_query.Node_CaseTestExpr:
		return walkCaseTestExpr(v, n.CaseTestExpr)

	case *pg_query.Node_CaseWhen:
		return walkCaseWhen(v, n.CaseWhen)

	case *pg_query.Node_ClusterStmt:
		return walkClusterStmt(v, n.ClusterStmt)

	case *pg_query.Node_CoalesceExpr:
		return walkCoalesceExpr(v, n.CoalesceExpr)

	case *pg_query.Node_CoerceToDomain:
		return walkCoerceToDomain(v, n.CoerceToDomain)

	case *pg_query.Node_CoerceToDomainValue:
		return walkCoerceToDomainValue(v, n.CoerceToDomainValue)

	case *pg_query.Node_CoerceViaIo:
		return walkCoerceViaIO(v, n.CoerceViaIo)

	case *pg_query.Node_CollateClause:
		return walkCollateClause(v, n.CollateClause)

	case *pg_query.Node_CollateExpr:
		return walkCollateExpr(v, n.CollateExpr)

	case *pg_query.Node_ColumnDef:
		return walkColumnDef(v, n.ColumnDef)

	case *pg_query.Node_ColumnRef:
		return walkColumnRef(v, n.ColumnRef)

	case *pg_query.Node_CommentStmt:
		return walkCommentStmt(v, n.CommentStmt)

	case *pg_query.Node_CommonTableExpr:
		return walkCommonTableExpr(v, n.CommonTableExpr)

	case *pg_query.Node_Constraint:
		return walkConstraint(v, n.Constraint)

	case *pg_query.Node_ConstraintsSetStmt:
		return walkConstraintsSetStmt(v, n.ConstraintsSetStmt)

	case *pg_query.Node_ConvertRowtypeExpr:
		return walkRowtypeExpr(v, n.ConvertRowtypeExpr)

	case *pg_query.Node_CopyStmt:
		return walkCopyStmt(v, n.CopyStmt)

	case *pg_query.Node_CreateAmStmt:
		return walkCreateAmStmt(v, n.CreateAmStmt)

	case *pg_query.Node_CreateCastStmt:
		return walkCreateCastStmt(v, n.CreateCastStmt)

	case *pg_query.Node_CreateConversionStmt:
		return walkCreateConversionStmt(v, n.CreateConversionStmt)

	case *pg_query.Node_CreateDomainStmt:
		return walkCreateDomainStmt(v, n.CreateDomainStmt)

	case *pg_query.Node_CreateEnumStmt:
		return walkCreateEnumStmt(v, n.CreateEnumStmt)

	case *pg_query.Node_CreateEventTrigStmt:
		return walkCreateEventTrigStmt(v, n.CreateEventTrigStmt)

	case *pg_query.Node_CreateExtensionStmt:
		return walkCreateExtensionStmt(v, n.CreateExtensionStmt)

	case *pg_query.Node_CreateFdwStmt:
		return walkCreateFdwStmt(v, n.CreateFdwStmt)

	case *pg_query.Node_CreateForeignTableStmt:
		return walkCreateForeignTableStmt(v, n.CreateForeignTableStmt)

	case *pg_query.Node_CreateFunctionStmt:
		return walkCreateFunctionStmt(v, n.CreateFunctionStmt)

	case *pg_query.Node_CreateOpClassItem:
		return walkCreateOpClassItem(v, n.CreateOpClassItem)

	case *pg_query.Node_CreateOpClassStmt:
		return walkCreateOpClassStmt(v, n.CreateOpClassStmt)

	case *pg_query.Node_CreateOpFamilyStmt:
		return walkCreateOpFamilyStmt(v, n.CreateOpFamilyStmt)

	case *pg_query.Node_CreatePlangStmt:
		return walkCreatePLangStmt(v, n.CreatePlangStmt)

	case *pg_query.Node_CreatePublicationStmt:
		return walkCreatePublicationStmt(v, n.CreatePublicationStmt)

	case *pg_query.Node_CreateRangeStmt:
		return walkCreateRangeStmt(v, n.CreateRangeStmt)

	case *pg_query.Node_CreateStatsStmt:
		return walkCreateStatsStmt(v, n.CreateStatsStmt)

	case *pg_query.Node_CreateStmt:
		return walkCreateStmt(v, n.CreateStmt)

	case *pg_query.Node_CreateSubscriptionStmt:
		return walkCreateSubscriptionStmt(v, n.CreateSubscriptionStmt)

	case *pg_query.Node_CreateTableAsStmt:
		return walkCreateTableAsStmt(v, n.CreateTableAsStmt)

	case *pg_query.Node_CreateTransformStmt:
		return walkCreateTransformStmt(v, n.CreateTransformStmt)

	case *pg_query.Node_CreateTrigStmt:
		return walkCreateTrigStmt(v, n.CreateTrigStmt)

	case *pg_query.Node_CreatedbStmt:
		return walkCreatedbStmt(v, n.CreatedbStmt)

	case *pg_query.Node_CurrentOfExpr:
		return walkCurrentOfExpr(v, n.CurrentOfExpr)

	case *pg_query.Node_DeclareCursorStmt:
		return walkDeclareCursorStmt(v, n.DeclareCursorStmt)

	case *pg_query.Node_DefElem:
		return walkDefElem(v, n.DefElem)

	case *pg_query.Node_DefineStmt:
		return walkDefineStmt(v, n.DefineStmt)

	case *pg_query.Node_DeleteStmt:
		return walkDeleteStmt(v, n.DeleteStmt)

	case *pg_query.Node_DoStmt:
		return walkDoStmt(v, n.DoStmt)

	case *pg_query.Node_DropStmt:
		return walkDropStmt(v, n.DropStmt)

	case *pg_query.Node_ExecuteStmt:
		return walkExecuteStmt(v, n.ExecuteStmt)

	case *pg_query.Node_ExplainStmt:
		return walkExplainStmt(v, n.ExplainStmt)

	case *pg_query.Node_FieldSelect:
		return walkFieldSelect(v, n.FieldSelect)

	case *pg_query.Node_FieldStore:
		return walkFieldStore(v, n.FieldStore)

	case *pg_query.Node_FromExpr:
		return walkFromExpr(v, n.FromExpr)

	case *pg_query.Node_FuncCall:
		return walkFuncCall(v, n.FuncCall)

	case *pg_query.Node_FuncExpr:
		return walkFuncExpr(v, n.FuncExpr)

	case *pg_query.Node_FunctionParameter:
		return walkFunctionParameter(v, n.FunctionParameter)

	case *pg_query.Node_GroupingFunc:
		return walkGroupingFunc(v, n.GroupingFunc)

	case *pg_query.Node_GroupingSet:
		return walkGroupingSet(v, n.GroupingSet)

	case *pg_query.Node_IndexElem:
		return walkIndexElem(v, n.IndexElem)

	case *pg_query.Node_IndexStmt:
		return walkIndexStmt(v, n.IndexStmt)

	case *pg_query.Node_InferClause:
		return walkInferClause(v, n.InferClause)

	case *pg_query.Node_InferenceElem:
		return walkInferenceElem(v, n.InferenceElem)

	case *pg_query.Node_InsertStmt:
		return walkInsertStmt(v, n.InsertStmt)

	case *pg_query.Node_IntoClause:
		return walkIntoClause(v, n.IntoClause)

	case *pg_query.Node_JoinExpr:
		return walkJoinExpr(v, n.JoinExpr)

	case *pg_query.Node_List:
		return walkList(v, n.List)

	case *pg_query.Node_LockStmt:
		return walkLockStmt(v, n.LockStmt)

	case *pg_query.Node_LockingClause:
		return walkLockingClause(v, n.LockingClause)

	case *pg_query.Node_MinMaxExpr:
		return walkMinMaxExpr(v, n.MinMaxExpr)

	case *pg_query.Node_MultiAssignRef:
		return walkMultiAssignRef(v, n.MultiAssignRef)

	case *pg_query.Node_NamedArgExpr:
		return walkNamedArgExpr(v, n.NamedArgExpr)

	case *pg_query.Node_NextValueExpr:
		return walkNextValueExpr(v, n.NextValueExpr)

	case *pg_query.Node_NullTest:
		return walkNullTest(v, n.NullTest)

	case *pg_query.Node_ObjectWithArgs:
		return walkObjectWithArgs(v, n.ObjectWithArgs)

	case *pg_query.Node_OnConflictClause:
		return walkOnConflictClause(v, n.OnConflictClause)

	case *pg_query.Node_OnConflictExpr:
		return walkOnConflictExpr(v, n.OnConflictExpr)

	case *pg_query.Node_OpExpr:
		return walkOpExpr(v, n.OpExpr)

	case *pg_query.Node_Param:
		return walkParam(v, n.Param)

	case *pg_query.Node_PrepareStmt:
		return walkPrepareStmt(v, n.PrepareStmt)

	case *pg_query.Node_Query:
		return walkQuery(v, n.Query)

	case *pg_query.Node_RangeFunction:
		return walkRangeFunction(v, n.RangeFunction)

	case *pg_query.Node_RangeSubselect:
		return walkRangeSubselect(v, n.RangeSubselect)

	case *pg_query.Node_RangeTableFunc:
		return walkRangeTableFunc(v, n.RangeTableFunc)

	case *pg_query.Node_RangeTableFuncCol:
		return walkRangeTableFuncCol(v, n.RangeTableFuncCol)

	case *pg_query.Node_RangeTableSample:
		return walkRangeTableSample(v, n.RangeTableSample)

	case *pg_query.Node_RangeTblEntry:
		return walkRangeTblEntry(v, n.RangeTblEntry)

	case *pg_query.Node_RangeTblFunction:
		return walkRangeTblFunction(v, n.RangeTblFunction)

	case *pg_query.Node_RawStmt:
		return walkRawStmt(v, n.RawStmt)

	case *pg_query.Node_RefreshMatViewStmt:
		return walkRefreshMatViewStmt(v, n.RefreshMatViewStmt)

	case *pg_query.Node_ReindexStmt:
		return walkReindexStmt(v, n.ReindexStmt)

	case *pg_query.Node_RelabelType:
		return walkRelabelType(v, n.RelabelType)

	case *pg_query.Node_RenameStmt:
		return walkRenameStmt(v, n.RenameStmt)

	case *pg_query.Node_ResTarget:
		return walkResTarget(v, n.ResTarget)

	case *pg_query.Node_RowCompareExpr:
		return walkRowCompareExpr(v, n.RowCompareExpr)

	case *pg_query.Node_RowExpr:
		return walkRowExpr(v, n.RowExpr)

	case *pg_query.Node_RuleStmt:
		return walkRuleStmt(v, n.RuleStmt)

	case *pg_query.Node_SqlvalueFunction:
		return walkSQLValueFunction(v, n.SqlvalueFunction)

	case *pg_query.Node_ScalarArrayOpExpr:
		return walkScalarArrayOpExpr(v, n.ScalarArrayOpExpr)

	case *pg_query.Node_SecLabelStmt:
		return walkSecLabelStmt(v, n.SecLabelStmt)

	case *pg_query.Node_SelectStmt:
		return walkSelectStmt(v, n.SelectStmt)

	case *pg_query.Node_SetOperationStmt:
		return walkSetOperationStmt(v, n.SetOperationStmt)

	case *pg_query.Node_SetToDefault:
		return walkSetToDefault(v, n.SetToDefault)

	case *pg_query.Node_SortBy:
		return walkSortBy(v, n.SortBy)

	case *pg_query.Node_SubLink:
		return walkSubLink(v, n.SubLink)

	case *pg_query.Node_SubPlan:
		return walkSubPlan(v, n.SubPlan)

	case *pg_query.Node_TableFunc:
		return walkTableFunc(v, n.TableFunc)

	case *pg_query.Node_TableLikeClause:
		return walkTableLikeClause(v, n.TableLikeClause)

	case *pg_query.Node_TableSampleClause:
		return walkTableSampleClause(v, n.TableSampleClause)

	case *pg_query.Node_TargetEntry:
		return walkTargetEntry(v, n.TargetEntry)

	case *pg_query.Node_TransactionStmt:
		return walkTransactionStmt(v, n.TransactionStmt)

	case *pg_query.Node_TruncateStmt:
		return walkTruncateStmt(v, n.TruncateStmt)

	case *pg_query.Node_TypeCast:
		return walkTypeCast(v, n.TypeCast)

	case *pg_query.Node_TypeName:
		return walkTypeName(v, n.TypeName)

	case *pg_query.Node_UpdateStmt:
		return walkUpdateStmt(v, n.UpdateStmt)

	case *pg_query.Node_Var:
		return walkVar(v, n.Var)

	case *pg_query.Node_VariableSetStmt:
		return walkVariableSetStmt(v, n.VariableSetStmt)

	case *pg_query.Node_ViewStmt:
		return walkViewStmt(v, n.ViewStmt)

	case *pg_query.Node_WindowClause:
		return walkWindowClause(v, n.WindowClause)

	case *pg_query.Node_WindowDef:
		return walkWindowDef(v, n.WindowDef)

	case *pg_query.Node_WindowFunc:
		return walkWindowFunc(v, n.WindowFunc)

	case *pg_query.Node_WithCheckOption:
		return walkWithCheckOption(v, n.WithCheckOption)

	case *pg_query.Node_WithClause:
		return walkWithClause(v, n.WithClause)

	case *pg_query.Node_XmlExpr:
		return walkXmlExpr(v, n.XmlExpr)

	case *pg_query.Node_XmlSerialize:
		return walkXmlSerialize(v, n.XmlSerialize)
	}
	// Revisit original node after its children have been processed.
	return v.VisitEnd(node)
}
