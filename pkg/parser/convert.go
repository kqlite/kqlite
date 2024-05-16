package parser

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

func convertFuncParamMode(m pg_query.FunctionParameterMode) (ast.FuncParamMode, error) {
	switch m {
	case pg_query.FunctionParameterMode_FUNC_PARAM_IN:
		return ast.FuncParamIn, nil
	case pg_query.FunctionParameterMode_FUNC_PARAM_OUT:
		return ast.FuncParamOut, nil
	case pg_query.FunctionParameterMode_FUNC_PARAM_INOUT:
		return ast.FuncParamInOut, nil
	case pg_query.FunctionParameterMode_FUNC_PARAM_VARIADIC:
		return ast.FuncParamVariadic, nil
	case pg_query.FunctionParameterMode_FUNC_PARAM_TABLE:
		return ast.FuncParamTable, nil
	case pg_query.FunctionParameterMode_FUNC_PARAM_DEFAULT:
		return ast.FuncParamDefault, nil
	default:
		return -1, fmt.Errorf("parse func param: invalid mode %v", m)
	}
}

func convertSubLinkType(t pg_query.SubLinkType) (ast.SubLinkType, error) {
	switch t {
	case pg_query.SubLinkType_EXISTS_SUBLINK:
		return ast.EXISTS_SUBLINK, nil
	case pg_query.SubLinkType_ALL_SUBLINK:
		return ast.ALL_SUBLINK, nil
	case pg_query.SubLinkType_ANY_SUBLINK:
		return ast.ANY_SUBLINK, nil
	case pg_query.SubLinkType_ROWCOMPARE_SUBLINK:
		return ast.ROWCOMPARE_SUBLINK, nil
	case pg_query.SubLinkType_EXPR_SUBLINK:
		return ast.EXPR_SUBLINK, nil
	case pg_query.SubLinkType_MULTIEXPR_SUBLINK:
		return ast.MULTIEXPR_SUBLINK, nil
	case pg_query.SubLinkType_ARRAY_SUBLINK:
		return ast.ARRAY_SUBLINK, nil
	case pg_query.SubLinkType_CTE_SUBLINK:
		return ast.CTE_SUBLINK, nil
	default:
		return 0, fmt.Errorf("parse sublink type: unknown type %s", t)
	}
}

func convertList(l *pg_query.List) *ast.List {
	out := &ast.List{}
	for _, item := range l.Items {
		out.Items = append(out.Items, convertNode(item))
	}
	return out
}

func convertSlice(nodes []*pg_query.Node) *ast.List {
	out := &ast.List{}
	for _, n := range nodes {
		out.Items = append(out.Items, convertNode(n))
	}
	return out
}

func convert(node *pg_query.Node) (Statement, error) {
	return convertNode(node)
}

func convertA_ArrayExpr(n *pg_query.A_ArrayExpr) *ast.A_ArrayExpr {
	if n == nil {
		return nil
	}
	return &ast.A_ArrayExpr{
		Elements: convertSlice(n.Elements),
		Location: int(n.Location),
	}
}

func convertA_Const(n *pg_query.A_Const) *ast.A_Const {
	if n == nil {
		return nil
	}
	var val ast.Node
	if n.Isnull {
		val = &ast.Null{}
	} else {
		switch v := n.Val.(type) {
		case *pg_query.A_Const_Boolval:
			val = convertBoolean(v.Boolval)
		case *pg_query.A_Const_Bsval:
			val = convertBitString(v.Bsval)
		case *pg_query.A_Const_Fval:
			val = convertFloat(v.Fval)
		case *pg_query.A_Const_Ival:
			val = convertInteger(v.Ival)
		case *pg_query.A_Const_Sval:
			val = convertString(v.Sval)
		}
	}
	return &ast.A_Const{
		Val:      val,
		Location: int(n.Location),
	}
}

func convertA_Expr(n *pg_query.A_Expr) *ast.A_Expr {
	if n == nil {
		return nil
	}
	return &ast.A_Expr{
		Kind:     ast.A_Expr_Kind(n.Kind),
		Name:     convertSlice(n.Name),
		Lexpr:    convertNode(n.Lexpr),
		Rexpr:    convertNode(n.Rexpr),
		Location: int(n.Location),
	}
}

func convertA_Indices(n *pg_query.A_Indices) *ast.A_Indices {
	if n == nil {
		return nil
	}
	return &ast.A_Indices{
		IsSlice: n.IsSlice,
		Lidx:    convertNode(n.Lidx),
		Uidx:    convertNode(n.Uidx),
	}
}

func convertA_Indirection(n *pg_query.A_Indirection) *ast.A_Indirection {
	if n == nil {
		return nil
	}
	return &ast.A_Indirection{
		Arg:         convertNode(n.Arg),
		Indirection: convertSlice(n.Indirection),
	}
}

func convertA_Star(n *pg_query.A_Star) *ast.A_Star {
	if n == nil {
		return nil
	}
	return &ast.A_Star{}
}

func convertAccessPriv(n *pg_query.AccessPriv) *ast.AccessPriv {
	if n == nil {
		return nil
	}
	return &ast.AccessPriv{
		PrivName: makeString(n.PrivName),
		Cols:     convertSlice(n.Cols),
	}
}

func convertAggref(n *pg_query.Aggref) *ast.Aggref {
	if n == nil {
		return nil
	}
	return &ast.Aggref{
		Xpr:           convertNode(n.Xpr),
		Aggfnoid:      ast.Oid(n.Aggfnoid),
		Aggtype:       ast.Oid(n.Aggtype),
		Aggcollid:     ast.Oid(n.Aggcollid),
		Inputcollid:   ast.Oid(n.Inputcollid),
		Aggargtypes:   convertSlice(n.Aggargtypes),
		Aggdirectargs: convertSlice(n.Aggdirectargs),
		Args:          convertSlice(n.Args),
		Aggorder:      convertSlice(n.Aggorder),
		Aggdistinct:   convertSlice(n.Aggdistinct),
		Aggfilter:     convertNode(n.Aggfilter),
		Aggstar:       n.Aggstar,
		Aggvariadic:   n.Aggvariadic,
		Aggkind:       makeByte(n.Aggkind),
		Agglevelsup:   ast.Index(n.Agglevelsup),
		Aggsplit:      ast.AggSplit(n.Aggsplit),
		Location:      int(n.Location),
	}
}

func convertAlias(n *pg_query.Alias) *ast.Alias {
	if n == nil {
		return nil
	}
	return &ast.Alias{
		Aliasname: makeString(n.Aliasname),
		Colnames:  convertSlice(n.Colnames),
	}
}

func convertAlterCollationStmt(n *pg_query.AlterCollationStmt) *ast.AlterCollationStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterCollationStmt{
		Collname: convertSlice(n.Collname),
	}
}

func convertAlterDatabaseSetStmt(n *pg_query.AlterDatabaseSetStmt) *ast.AlterDatabaseSetStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterDatabaseSetStmt{
		Dbname:  makeString(n.Dbname),
		Setstmt: convertVariableSetStmt(n.Setstmt),
	}
}

func convertAlterDatabaseStmt(n *pg_query.AlterDatabaseStmt) *ast.AlterDatabaseStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterDatabaseStmt{
		Dbname:  makeString(n.Dbname),
		Options: convertSlice(n.Options),
	}
}

func convertAlterDefaultPrivilegesStmt(n *pg_query.AlterDefaultPrivilegesStmt) *ast.AlterDefaultPrivilegesStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterDefaultPrivilegesStmt{
		Options: convertSlice(n.Options),
		Action:  convertGrantStmt(n.Action),
	}
}

func convertAlterDomainStmt(n *pg_query.AlterDomainStmt) *ast.AlterDomainStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterDomainStmt{
		Subtype:   makeByte(n.Subtype),
		TypeName:  convertSlice(n.TypeName),
		Name:      makeString(n.Name),
		Def:       convertNode(n.Def),
		Behavior:  ast.DropBehavior(n.Behavior),
		MissingOk: n.MissingOk,
	}
}

func convertAlterEnumStmt(n *pg_query.AlterEnumStmt) *ast.AlterEnumStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterEnumStmt{
		TypeName:           convertSlice(n.TypeName),
		OldVal:             makeString(n.OldVal),
		NewVal:             makeString(n.NewVal),
		NewValNeighbor:     makeString(n.NewValNeighbor),
		NewValIsAfter:      n.NewValIsAfter,
		SkipIfNewValExists: n.SkipIfNewValExists,
	}
}

func convertAlterEventTrigStmt(n *pg_query.AlterEventTrigStmt) *ast.AlterEventTrigStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterEventTrigStmt{
		Trigname:  makeString(n.Trigname),
		Tgenabled: makeByte(n.Tgenabled),
	}
}

func convertAlterExtensionContentsStmt(n *pg_query.AlterExtensionContentsStmt) *ast.AlterExtensionContentsStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterExtensionContentsStmt{
		Extname: makeString(n.Extname),
		Action:  int(n.Action),
		Objtype: ast.ObjectType(n.Objtype),
		Object:  convertNode(n.Object),
	}
}

func convertAlterExtensionStmt(n *pg_query.AlterExtensionStmt) *ast.AlterExtensionStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterExtensionStmt{
		Extname: makeString(n.Extname),
		Options: convertSlice(n.Options),
	}
}

func convertAlterFdwStmt(n *pg_query.AlterFdwStmt) *ast.AlterFdwStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterFdwStmt{
		Fdwname:     makeString(n.Fdwname),
		FuncOptions: convertSlice(n.FuncOptions),
		Options:     convertSlice(n.Options),
	}
}

func convertAlterForeignServerStmt(n *pg_query.AlterForeignServerStmt) *ast.AlterForeignServerStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterForeignServerStmt{
		Servername: makeString(n.Servername),
		Version:    makeString(n.Version),
		Options:    convertSlice(n.Options),
		HasVersion: n.HasVersion,
	}
}

func convertAlterFunctionStmt(n *pg_query.AlterFunctionStmt) *ast.AlterFunctionStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterFunctionStmt{
		Func:    convertObjectWithArgs(n.Func),
		Actions: convertSlice(n.Actions),
	}
}

func convertAlterObjectDependsStmt(n *pg_query.AlterObjectDependsStmt) *ast.AlterObjectDependsStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterObjectDependsStmt{
		ObjectType: ast.ObjectType(n.ObjectType),
		Relation:   convertRangeVar(n.Relation),
		Object:     convertNode(n.Object),
		Extname:    convertString(n.Extname),
	}
}

func convertAlterObjectSchemaStmt(n *pg_query.AlterObjectSchemaStmt) *ast.AlterObjectSchemaStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterObjectSchemaStmt{
		ObjectType: ast.ObjectType(n.ObjectType),
		Relation:   convertRangeVar(n.Relation),
		Object:     convertNode(n.Object),
		Newschema:  makeString(n.Newschema),
		MissingOk:  n.MissingOk,
	}
}

func convertAlterOpFamilyStmt(n *pg_query.AlterOpFamilyStmt) *ast.AlterOpFamilyStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterOpFamilyStmt{
		Opfamilyname: convertSlice(n.Opfamilyname),
		Amname:       makeString(n.Amname),
		IsDrop:       n.IsDrop,
		Items:        convertSlice(n.Items),
	}
}

func convertAlterOperatorStmt(n *pg_query.AlterOperatorStmt) *ast.AlterOperatorStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterOperatorStmt{
		Opername: convertObjectWithArgs(n.Opername),
		Options:  convertSlice(n.Options),
	}
}

func convertAlterOwnerStmt(n *pg_query.AlterOwnerStmt) *ast.AlterOwnerStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterOwnerStmt{
		ObjectType: ast.ObjectType(n.ObjectType),
		Relation:   convertRangeVar(n.Relation),
		Object:     convertNode(n.Object),
		Newowner:   convertRoleSpec(n.Newowner),
	}
}

func convertAlterPolicyStmt(n *pg_query.AlterPolicyStmt) *ast.AlterPolicyStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterPolicyStmt{
		PolicyName: makeString(n.PolicyName),
		Table:      convertRangeVar(n.Table),
		Roles:      convertSlice(n.Roles),
		Qual:       convertNode(n.Qual),
		WithCheck:  convertNode(n.WithCheck),
	}
}

func convertAlterPublicationStmt(n *pg_query.AlterPublicationStmt) *ast.AlterPublicationStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterPublicationStmt{
		Pubname:      makeString(n.Pubname),
		Options:      convertSlice(n.Options),
		Tables:       convertSlice(n.Pubobjects),
		ForAllTables: n.ForAllTables,
		TableAction:  ast.DefElemAction(n.Action),
	}
}

func convertAlterRoleSetStmt(n *pg_query.AlterRoleSetStmt) *ast.AlterRoleSetStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterRoleSetStmt{
		Role:     convertRoleSpec(n.Role),
		Database: makeString(n.Database),
		Setstmt:  convertVariableSetStmt(n.Setstmt),
	}
}

func convertAlterRoleStmt(n *pg_query.AlterRoleStmt) *ast.AlterRoleStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterRoleStmt{
		Role:    convertRoleSpec(n.Role),
		Options: convertSlice(n.Options),
		Action:  int(n.Action),
	}
}

func convertAlterSeqStmt(n *pg_query.AlterSeqStmt) *ast.AlterSeqStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterSeqStmt{
		Sequence:    convertRangeVar(n.Sequence),
		Options:     convertSlice(n.Options),
		ForIdentity: n.ForIdentity,
		MissingOk:   n.MissingOk,
	}
}

func convertAlterSubscriptionStmt(n *pg_query.AlterSubscriptionStmt) *ast.AlterSubscriptionStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterSubscriptionStmt{
		Kind:        ast.AlterSubscriptionType(n.Kind),
		Subname:     makeString(n.Subname),
		Conninfo:    makeString(n.Conninfo),
		Publication: convertSlice(n.Publication),
		Options:     convertSlice(n.Options),
	}
}

func convertAlterSystemStmt(n *pg_query.AlterSystemStmt) *ast.AlterSystemStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterSystemStmt{
		Setstmt: convertVariableSetStmt(n.Setstmt),
	}
}

func convertAlterTSConfigurationStmt(n *pg_query.AlterTSConfigurationStmt) *ast.AlterTSConfigurationStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterTSConfigurationStmt{
		Kind:      ast.AlterTSConfigType(n.Kind),
		Cfgname:   convertSlice(n.Cfgname),
		Tokentype: convertSlice(n.Tokentype),
		Dicts:     convertSlice(n.Dicts),
		Override:  n.Override,
		Replace:   n.Replace,
		MissingOk: n.MissingOk,
	}
}

func convertAlterTSDictionaryStmt(n *pg_query.AlterTSDictionaryStmt) *ast.AlterTSDictionaryStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterTSDictionaryStmt{
		Dictname: convertSlice(n.Dictname),
		Options:  convertSlice(n.Options),
	}
}

func convertAlterTableCmd(n *pg_query.AlterTableCmd) *ast.AlterTableCmd {
	if n == nil {
		return nil
	}
	def := convertNode(n.Def)
	columnDef := def.(*ast.ColumnDef)
	return &ast.AlterTableCmd{
		Subtype:   ast.AlterTableType(n.Subtype),
		Name:      makeString(n.Name),
		Newowner:  convertRoleSpec(n.Newowner),
		Def:       columnDef,
		Behavior:  ast.DropBehavior(n.Behavior),
		MissingOk: n.MissingOk,
	}
}

func convertAlterTableMoveAllStmt(n *pg_query.AlterTableMoveAllStmt) *ast.AlterTableMoveAllStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterTableMoveAllStmt{
		OrigTablespacename: makeString(n.OrigTablespacename),
		Objtype:            ast.ObjectType(n.Objtype),
		Roles:              convertSlice(n.Roles),
		NewTablespacename:  makeString(n.NewTablespacename),
		Nowait:             n.Nowait,
	}
}

func convertAlterTableSpaceOptionsStmt(n *pg_query.AlterTableSpaceOptionsStmt) *ast.AlterTableSpaceOptionsStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterTableSpaceOptionsStmt{
		Tablespacename: makeString(n.Tablespacename),
		Options:        convertSlice(n.Options),
		IsReset:        n.IsReset,
	}
}

func convertAlterTableStmt(n *pg_query.AlterTableStmt) *ast.AlterTableStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterTableStmt{
		Relation:  convertRangeVar(n.Relation),
		Cmds:      convertSlice(n.Cmds),
		Relkind:   ast.ObjectType(n.Objtype),
		MissingOk: n.MissingOk,
	}
}

func convertAlterUserMappingStmt(n *pg_query.AlterUserMappingStmt) *ast.AlterUserMappingStmt {
	if n == nil {
		return nil
	}
	return &ast.AlterUserMappingStmt{
		User:       convertRoleSpec(n.User),
		Servername: makeString(n.Servername),
		Options:    convertSlice(n.Options),
	}
}

func convertAlternativeSubPlan(n *pg_query.AlternativeSubPlan) *ast.AlternativeSubPlan {
	if n == nil {
		return nil
	}
	return &ast.AlternativeSubPlan{
		Xpr:      convertNode(n.Xpr),
		Subplans: convertSlice(n.Subplans),
	}
}

func convertArrayCoerceExpr(n *pg_query.ArrayCoerceExpr) *ast.ArrayCoerceExpr {
	if n == nil {
		return nil
	}
	return &ast.ArrayCoerceExpr{
		Xpr:          convertNode(n.Xpr),
		Arg:          convertNode(n.Arg),
		Resulttype:   ast.Oid(n.Resulttype),
		Resulttypmod: n.Resulttypmod,
		Resultcollid: ast.Oid(n.Resultcollid),
		Coerceformat: ast.CoercionForm(n.Coerceformat),
		Location:     int(n.Location),
	}
}

func convertArrayExpr(n *pg_query.ArrayExpr) *ast.ArrayExpr {
	if n == nil {
		return nil
	}
	return &ast.ArrayExpr{
		Xpr:           convertNode(n.Xpr),
		ArrayTypeid:   ast.Oid(n.ArrayTypeid),
		ArrayCollid:   ast.Oid(n.ArrayCollid),
		ElementTypeid: ast.Oid(n.ElementTypeid),
		Elements:      convertSlice(n.Elements),
		Multidims:     n.Multidims,
		Location:      int(n.Location),
	}
}

func convertBitString(n *pg_query.BitString) *ast.BitString {
	if n == nil {
		return nil
	}
	return &ast.BitString{
		Str: n.Bsval,
	}
}

func convertBoolExpr(n *pg_query.BoolExpr) *ast.BoolExpr {
	if n == nil {
		return nil
	}
	return &ast.BoolExpr{
		Xpr:      convertNode(n.Xpr),
		Boolop:   ast.BoolExprType(n.Boolop),
		Args:     convertSlice(n.Args),
		Location: int(n.Location),
	}
}

func convertBoolean(n *pg_query.Boolean) *ast.Boolean {
	if n == nil {
		return nil
	}
	return &ast.Boolean{
		Boolval: n.Boolval,
	}
}

func convertBooleanTest(n *pg_query.BooleanTest) *ast.BooleanTest {
	if n == nil {
		return nil
	}
	return &ast.BooleanTest{
		Xpr:          convertNode(n.Xpr),
		Arg:          convertNode(n.Arg),
		Booltesttype: ast.BoolTestType(n.Booltesttype),
		Location:     int(n.Location),
	}
}

func convertCallStmt(n *pg_query.CallStmt) *ast.CallStmt {
	if n == nil {
		return nil
	}
	rel, err := parseRelationFromNodes(n.Funccall.Funcname)
	if err != nil {
		// TODO: How should we handle errors?
		panic(err)
	}

	return &ast.CallStmt{
		FuncCall: &ast.FuncCall{
			Func:           rel.FuncName(),
			Funcname:       convertSlice(n.Funccall.Funcname),
			Args:           convertSlice(n.Funccall.Args),
			AggOrder:       convertSlice(n.Funccall.AggOrder),
			AggFilter:      convertNode(n.Funccall.AggFilter),
			AggWithinGroup: n.Funccall.AggWithinGroup,
			AggStar:        n.Funccall.AggStar,
			AggDistinct:    n.Funccall.AggDistinct,
			FuncVariadic:   n.Funccall.FuncVariadic,
			Over:           convertWindowDef(n.Funccall.Over),
			Location:       int(n.Funccall.Location),
		},
	}
}

func convertCaseExpr(n *pg_query.CaseExpr) *ast.CaseExpr {
	if n == nil {
		return nil
	}
	return &ast.CaseExpr{
		Xpr:        convertNode(n.Xpr),
		Casetype:   ast.Oid(n.Casetype),
		Casecollid: ast.Oid(n.Casecollid),
		Arg:        convertNode(n.Arg),
		Args:       convertSlice(n.Args),
		Defresult:  convertNode(n.Defresult),
		Location:   int(n.Location),
	}
}

func convertCaseTestExpr(n *pg_query.CaseTestExpr) *ast.CaseTestExpr {
	if n == nil {
		return nil
	}
	return &ast.CaseTestExpr{
		Xpr:       convertNode(n.Xpr),
		TypeId:    ast.Oid(n.TypeId),
		TypeMod:   n.TypeMod,
		Collation: ast.Oid(n.Collation),
	}
}

func convertCaseWhen(n *pg_query.CaseWhen) *ast.CaseWhen {
	if n == nil {
		return nil
	}
	return &ast.CaseWhen{
		Xpr:      convertNode(n.Xpr),
		Expr:     convertNode(n.Expr),
		Result:   convertNode(n.Result),
		Location: int(n.Location),
	}
}

func convertCheckPointStmt(n *pg_query.CheckPointStmt) *ast.CheckPointStmt {
	if n == nil {
		return nil
	}
	return &ast.CheckPointStmt{}
}

func convertClosePortalStmt(n *pg_query.ClosePortalStmt) *ast.ClosePortalStmt {
	if n == nil {
		return nil
	}
	return &ast.ClosePortalStmt{
		Portalname: makeString(n.Portalname),
	}
}

func convertClusterStmt(n *pg_query.ClusterStmt) *ast.ClusterStmt {
	if n == nil {
		return nil
	}
	return &ast.ClusterStmt{
		Relation:  convertRangeVar(n.Relation),
		Indexname: makeString(n.Indexname),
	}
}

func convertCoalesceExpr(n *pg_query.CoalesceExpr) *ast.CoalesceExpr {
	if n == nil {
		return nil
	}
	return &ast.CoalesceExpr{
		Xpr:            convertNode(n.Xpr),
		Coalescetype:   ast.Oid(n.Coalescetype),
		Coalescecollid: ast.Oid(n.Coalescecollid),
		Args:           convertSlice(n.Args),
		Location:       int(n.Location),
	}
}

func convertCoerceToDomain(n *pg_query.CoerceToDomain) *ast.CoerceToDomain {
	if n == nil {
		return nil
	}
	return &ast.CoerceToDomain{
		Xpr:            convertNode(n.Xpr),
		Arg:            convertNode(n.Arg),
		Resulttype:     ast.Oid(n.Resulttype),
		Resulttypmod:   n.Resulttypmod,
		Resultcollid:   ast.Oid(n.Resultcollid),
		Coercionformat: ast.CoercionForm(n.Coercionformat),
		Location:       int(n.Location),
	}
}

func convertCoerceToDomainValue(n *pg_query.CoerceToDomainValue) *ast.CoerceToDomainValue {
	if n == nil {
		return nil
	}
	return &ast.CoerceToDomainValue{
		Xpr:       convertNode(n.Xpr),
		TypeId:    ast.Oid(n.TypeId),
		TypeMod:   n.TypeMod,
		Collation: ast.Oid(n.Collation),
		Location:  int(n.Location),
	}
}

func convertCoerceViaIO(n *pg_query.CoerceViaIO) *ast.CoerceViaIO {
	if n == nil {
		return nil
	}
	return &ast.CoerceViaIO{
		Xpr:          convertNode(n.Xpr),
		Arg:          convertNode(n.Arg),
		Resulttype:   ast.Oid(n.Resulttype),
		Resultcollid: ast.Oid(n.Resultcollid),
		Coerceformat: ast.CoercionForm(n.Coerceformat),
		Location:     int(n.Location),
	}
}

func convertCollateClause(n *pg_query.CollateClause) *ast.CollateClause {
	if n == nil {
		return nil
	}
	return &ast.CollateClause{
		Arg:      convertNode(n.Arg),
		Collname: convertSlice(n.Collname),
		Location: int(n.Location),
	}
}

func convertCollateExpr(n *pg_query.CollateExpr) *ast.CollateExpr {
	if n == nil {
		return nil
	}
	return &ast.CollateExpr{
		Xpr:      convertNode(n.Xpr),
		Arg:      convertNode(n.Arg),
		CollOid:  ast.Oid(n.CollOid),
		Location: int(n.Location),
	}
}

func convertColumnDef(n *pg_query.ColumnDef) *ast.ColumnDef {
	if n == nil {
		return nil
	}
	return &ast.ColumnDef{
		Colname:       n.Colname,
		TypeName:      convertTypeName(n.TypeName),
		Inhcount:      int(n.Inhcount),
		IsLocal:       n.IsLocal,
		IsNotNull:     n.IsNotNull,
		IsFromType:    n.IsFromType,
		Storage:       makeByte(n.Storage),
		RawDefault:    convertNode(n.RawDefault),
		CookedDefault: convertNode(n.CookedDefault),
		Identity:      makeByte(n.Identity),
		CollClause:    convertCollateClause(n.CollClause),
		CollOid:       ast.Oid(n.CollOid),
		Constraints:   convertSlice(n.Constraints),
		Fdwoptions:    convertSlice(n.Fdwoptions),
		Location:      int(n.Location),
	}
}

func convertColumnRef(n *pg_query.ColumnRef) *ast.ColumnRef {
	if n == nil {
		return nil
	}
	return &ast.ColumnRef{
		Fields:   convertSlice(n.Fields),
		Location: int(n.Location),
	}
}

func convertCommentStmt(n *pg_query.CommentStmt) *ast.CommentStmt {
	if n == nil {
		return nil
	}
	return &ast.CommentStmt{
		Objtype: ast.ObjectType(n.Objtype),
		Object:  convertNode(n.Object),
		Comment: makeString(n.Comment),
	}
}

func convertCommonTableExpr(n *pg_query.CommonTableExpr) *ast.CommonTableExpr {
	if n == nil {
		return nil
	}
	return &ast.CommonTableExpr{
		Ctename:          makeString(n.Ctename),
		Aliascolnames:    convertSlice(n.Aliascolnames),
		Ctequery:         convertNode(n.Ctequery),
		Location:         int(n.Location),
		Cterecursive:     n.Cterecursive,
		Cterefcount:      int(n.Cterefcount),
		Ctecolnames:      convertSlice(n.Ctecolnames),
		Ctecoltypes:      convertSlice(n.Ctecoltypes),
		Ctecoltypmods:    convertSlice(n.Ctecoltypmods),
		Ctecolcollations: convertSlice(n.Ctecolcollations),
	}
}

func convertCompositeTypeStmt(n *pg_query.CompositeTypeStmt) *ast.CompositeTypeStmt {
	if n == nil {
		return nil
	}
	rel := parseRelationFromRangeVar(n.Typevar)
	return &ast.CompositeTypeStmt{
		TypeName: rel.TypeName(),
	}
}

func convertConstraint(n *pg_query.Constraint) *ast.Constraint {
	if n == nil {
		return nil
	}
	return &ast.Constraint{
		Contype:        ast.ConstrType(n.Contype),
		Conname:        makeString(n.Conname),
		Deferrable:     n.Deferrable,
		Initdeferred:   n.Initdeferred,
		Location:       int(n.Location),
		IsNoInherit:    n.IsNoInherit,
		RawExpr:        convertNode(n.RawExpr),
		CookedExpr:     makeString(n.CookedExpr),
		GeneratedWhen:  makeByte(n.GeneratedWhen),
		Keys:           convertSlice(n.Keys),
		Exclusions:     convertSlice(n.Exclusions),
		Options:        convertSlice(n.Options),
		Indexname:      makeString(n.Indexname),
		Indexspace:     makeString(n.Indexspace),
		AccessMethod:   makeString(n.AccessMethod),
		WhereClause:    convertNode(n.WhereClause),
		Pktable:        convertRangeVar(n.Pktable),
		FkAttrs:        convertSlice(n.FkAttrs),
		PkAttrs:        convertSlice(n.PkAttrs),
		FkMatchtype:    makeByte(n.FkMatchtype),
		FkUpdAction:    makeByte(n.FkUpdAction),
		FkDelAction:    makeByte(n.FkDelAction),
		OldConpfeqop:   convertSlice(n.OldConpfeqop),
		OldPktableOid:  ast.Oid(n.OldPktableOid),
		SkipValidation: n.SkipValidation,
		InitiallyValid: n.InitiallyValid,
	}
}

func convertConstraintsSetStmt(n *pg_query.ConstraintsSetStmt) *ast.ConstraintsSetStmt {
	if n == nil {
		return nil
	}
	return &ast.ConstraintsSetStmt{
		Constraints: convertSlice(n.Constraints),
		Deferred:    n.Deferred,
	}
}

func convertConvertRowtypeExpr(n *pg_query.ConvertRowtypeExpr) *ast.ConvertRowtypeExpr {
	if n == nil {
		return nil
	}
	return &ast.ConvertRowtypeExpr{
		Xpr:           convertNode(n.Xpr),
		Arg:           convertNode(n.Arg),
		Resulttype:    ast.Oid(n.Resulttype),
		Convertformat: ast.CoercionForm(n.Convertformat),
		Location:      int(n.Location),
	}
}

func convertCopyStmt(n *pg_query.CopyStmt) *ast.CopyStmt {
	if n == nil {
		return nil
	}
	return &ast.CopyStmt{
		Relation:  convertRangeVar(n.Relation),
		Query:     convertNode(n.Query),
		Attlist:   convertSlice(n.Attlist),
		IsFrom:    n.IsFrom,
		IsProgram: n.IsProgram,
		Filename:  makeString(n.Filename),
		Options:   convertSlice(n.Options),
	}
}

func convertCreateAmStmt(n *pg_query.CreateAmStmt) *ast.CreateAmStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateAmStmt{
		Amname:      makeString(n.Amname),
		HandlerName: convertSlice(n.HandlerName),
		Amtype:      makeByte(n.Amtype),
	}
}

func convertCreateCastStmt(n *pg_query.CreateCastStmt) *ast.CreateCastStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateCastStmt{
		Sourcetype: convertTypeName(n.Sourcetype),
		Targettype: convertTypeName(n.Targettype),
		Func:       convertObjectWithArgs(n.Func),
		Context:    ast.CoercionContext(n.Context),
		Inout:      n.Inout,
	}
}

func convertCreateConversionStmt(n *pg_query.CreateConversionStmt) *ast.CreateConversionStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateConversionStmt{
		ConversionName:  convertSlice(n.ConversionName),
		ForEncodingName: makeString(n.ForEncodingName),
		ToEncodingName:  makeString(n.ToEncodingName),
		FuncName:        convertSlice(n.FuncName),
		Def:             n.Def,
	}
}

func convertCreateDomainStmt(n *pg_query.CreateDomainStmt) *ast.CreateDomainStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateDomainStmt{
		Domainname:  convertSlice(n.Domainname),
		TypeName:    convertTypeName(n.TypeName),
		CollClause:  convertCollateClause(n.CollClause),
		Constraints: convertSlice(n.Constraints),
	}
}

func convertCreateEnumStmt(n *pg_query.CreateEnumStmt) *ast.CreateEnumStmt {
	if n == nil {
		return nil
	}
	rel, err := parseRelationFromNodes(n.TypeName)
	if err != nil {
		panic(err)
	}
	return &ast.CreateEnumStmt{
		TypeName: rel.TypeName(),
		Vals:     convertSlice(n.Vals),
	}
}

func convertCreateEventTrigStmt(n *pg_query.CreateEventTrigStmt) *ast.CreateEventTrigStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateEventTrigStmt{
		Trigname:   makeString(n.Trigname),
		Eventname:  makeString(n.Eventname),
		Whenclause: convertSlice(n.Whenclause),
		Funcname:   convertSlice(n.Funcname),
	}
}

func convertCreateExtensionStmt(n *pg_query.CreateExtensionStmt) *ast.CreateExtensionStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateExtensionStmt{
		Extname:     makeString(n.Extname),
		IfNotExists: n.IfNotExists,
		Options:     convertSlice(n.Options),
	}
}

func convertCreateFdwStmt(n *pg_query.CreateFdwStmt) *ast.CreateFdwStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateFdwStmt{
		Fdwname:     makeString(n.Fdwname),
		FuncOptions: convertSlice(n.FuncOptions),
		Options:     convertSlice(n.Options),
	}
}

func convertCreateForeignServerStmt(n *pg_query.CreateForeignServerStmt) *ast.CreateForeignServerStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateForeignServerStmt{
		Servername:  makeString(n.Servername),
		Servertype:  makeString(n.Servertype),
		Version:     makeString(n.Version),
		Fdwname:     makeString(n.Fdwname),
		IfNotExists: n.IfNotExists,
		Options:     convertSlice(n.Options),
	}
}

func convertCreateForeignTableStmt(n *pg_query.CreateForeignTableStmt) *ast.CreateForeignTableStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateForeignTableStmt{
		Servername: makeString(n.Servername),
		Options:    convertSlice(n.Options),
	}
}

func convertCreateFunctionStmt(n *pg_query.CreateFunctionStmt) *ast.CreateFunctionStmt {
	if n == nil {
		return nil
	}
	rel, err := parseRelationFromNodes(n.Funcname)
	if err != nil {
		panic(err)
	}
	return &ast.CreateFunctionStmt{
		Replace:    n.Replace,
		Func:       rel.FuncName(),
		Params:     convertSlice(n.Parameters),
		ReturnType: convertTypeName(n.ReturnType),
		Options:    convertSlice(n.Options),
	}
}

func convertCreateOpClassItem(n *pg_query.CreateOpClassItem) *ast.CreateOpClassItem {
	if n == nil {
		return nil
	}
	return &ast.CreateOpClassItem{
		Itemtype:    int(n.Itemtype),
		Name:        convertObjectWithArgs(n.Name),
		Number:      int(n.Number),
		OrderFamily: convertSlice(n.OrderFamily),
		ClassArgs:   convertSlice(n.ClassArgs),
		Storedtype:  convertTypeName(n.Storedtype),
	}
}

func convertCreateOpClassStmt(n *pg_query.CreateOpClassStmt) *ast.CreateOpClassStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateOpClassStmt{
		Opclassname:  convertSlice(n.Opclassname),
		Opfamilyname: convertSlice(n.Opfamilyname),
		Amname:       makeString(n.Amname),
		Datatype:     convertTypeName(n.Datatype),
		Items:        convertSlice(n.Items),
		IsDefault:    n.IsDefault,
	}
}

func convertCreateOpFamilyStmt(n *pg_query.CreateOpFamilyStmt) *ast.CreateOpFamilyStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateOpFamilyStmt{
		Opfamilyname: convertSlice(n.Opfamilyname),
		Amname:       makeString(n.Amname),
	}
}

func convertCreatePLangStmt(n *pg_query.CreatePLangStmt) *ast.CreatePLangStmt {
	if n == nil {
		return nil
	}
	return &ast.CreatePLangStmt{
		Replace:     n.Replace,
		Plname:      makeString(n.Plname),
		Plhandler:   convertSlice(n.Plhandler),
		Plinline:    convertSlice(n.Plinline),
		Plvalidator: convertSlice(n.Plvalidator),
		Pltrusted:   n.Pltrusted,
	}
}

func convertCreatePolicyStmt(n *pg_query.CreatePolicyStmt) *ast.CreatePolicyStmt {
	if n == nil {
		return nil
	}
	return &ast.CreatePolicyStmt{
		PolicyName: makeString(n.PolicyName),
		Table:      convertRangeVar(n.Table),
		CmdName:    makeString(n.CmdName),
		Permissive: n.Permissive,
		Roles:      convertSlice(n.Roles),
		Qual:       convertNode(n.Qual),
		WithCheck:  convertNode(n.WithCheck),
	}
}

func convertCreatePublicationStmt(n *pg_query.CreatePublicationStmt) *ast.CreatePublicationStmt {
	if n == nil {
		return nil
	}
	return &ast.CreatePublicationStmt{
		Pubname:      makeString(n.Pubname),
		Options:      convertSlice(n.Options),
		Tables:       convertSlice(n.Pubobjects),
		ForAllTables: n.ForAllTables,
	}
}

func convertCreateRangeStmt(n *pg_query.CreateRangeStmt) *ast.CreateRangeStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateRangeStmt{
		TypeName: convertSlice(n.TypeName),
		Params:   convertSlice(n.Params),
	}
}

func convertCreateRoleStmt(n *pg_query.CreateRoleStmt) *ast.CreateRoleStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateRoleStmt{
		StmtType: ast.RoleStmtType(n.StmtType),
		Role:     makeString(n.Role),
		Options:  convertSlice(n.Options),
	}
}

func convertCreateSchemaStmt(n *pg_query.CreateSchemaStmt) *ast.CreateSchemaStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateSchemaStmt{
		Name:        makeString(n.Schemaname),
		Authrole:    convertRoleSpec(n.Authrole),
		SchemaElts:  convertSlice(n.SchemaElts),
		IfNotExists: n.IfNotExists,
	}
}

func convertCreateSeqStmt(n *pg_query.CreateSeqStmt) *ast.CreateSeqStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateSeqStmt{
		Sequence:    convertRangeVar(n.Sequence),
		Options:     convertSlice(n.Options),
		OwnerId:     ast.Oid(n.OwnerId),
		ForIdentity: n.ForIdentity,
		IfNotExists: n.IfNotExists,
	}
}

func convertCreateStatsStmt(n *pg_query.CreateStatsStmt) *ast.CreateStatsStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateStatsStmt{
		Defnames:    convertSlice(n.Defnames),
		StatTypes:   convertSlice(n.StatTypes),
		Exprs:       convertSlice(n.Exprs),
		Relations:   convertSlice(n.Relations),
		IfNotExists: n.IfNotExists,
	}
}

func convertCreateStmt(n *pg_query.CreateStmt) *ast.CreateStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateStmt{
		Relation:       convertRangeVar(n.Relation),
		TableElts:      convertSlice(n.TableElts),
		InhRelations:   convertSlice(n.InhRelations),
		Partbound:      convertPartitionBoundSpec(n.Partbound),
		Partspec:       convertPartitionSpec(n.Partspec),
		OfTypename:     convertTypeName(n.OfTypename),
		Constraints:    convertSlice(n.Constraints),
		Options:        convertSlice(n.Options),
		Oncommit:       ast.OnCommitAction(n.Oncommit),
		Tablespacename: makeString(n.Tablespacename),
		IfNotExists:    n.IfNotExists,
	}
}

func convertCreateSubscriptionStmt(n *pg_query.CreateSubscriptionStmt) *ast.CreateSubscriptionStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateSubscriptionStmt{
		Subname:     makeString(n.Subname),
		Conninfo:    makeString(n.Conninfo),
		Publication: convertSlice(n.Publication),
		Options:     convertSlice(n.Options),
	}
}

func convertCreateTableAsStmt(n *pg_query.CreateTableAsStmt) *ast.CreateTableAsStmt {
	if n == nil {
		return nil
	}
	res := &ast.CreateTableAsStmt{
		Query:        convertNode(n.Query),
		Into:         convertIntoClause(n.Into),
		Relkind:      ast.ObjectType(n.Objtype),
		IsSelectInto: n.IsSelectInto,
		IfNotExists:  n.IfNotExists,
	}
	return res
}

func convertCreateTableSpaceStmt(n *pg_query.CreateTableSpaceStmt) *ast.CreateTableSpaceStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateTableSpaceStmt{
		Tablespacename: makeString(n.Tablespacename),
		Owner:          convertRoleSpec(n.Owner),
		Location:       makeString(n.Location),
		Options:        convertSlice(n.Options),
	}
}

func convertCreateTransformStmt(n *pg_query.CreateTransformStmt) *ast.CreateTransformStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateTransformStmt{
		Replace:  n.Replace,
		TypeName: convertTypeName(n.TypeName),
		Lang:     makeString(n.Lang),
		Fromsql:  convertObjectWithArgs(n.Fromsql),
		Tosql:    convertObjectWithArgs(n.Tosql),
	}
}

func convertCreateTrigStmt(n *pg_query.CreateTrigStmt) *ast.CreateTrigStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateTrigStmt{
		Trigname:       makeString(n.Trigname),
		Relation:       convertRangeVar(n.Relation),
		Funcname:       convertSlice(n.Funcname),
		Args:           convertSlice(n.Args),
		Row:            n.Row,
		Timing:         int16(n.Timing),
		Events:         int16(n.Events),
		Columns:        convertSlice(n.Columns),
		WhenClause:     convertNode(n.WhenClause),
		Isconstraint:   n.Isconstraint,
		TransitionRels: convertSlice(n.TransitionRels),
		Deferrable:     n.Deferrable,
		Initdeferred:   n.Initdeferred,
		Constrrel:      convertRangeVar(n.Constrrel),
	}
}

func convertCreateUserMappingStmt(n *pg_query.CreateUserMappingStmt) *ast.CreateUserMappingStmt {
	if n == nil {
		return nil
	}
	return &ast.CreateUserMappingStmt{
		User:        convertRoleSpec(n.User),
		Servername:  makeString(n.Servername),
		IfNotExists: n.IfNotExists,
		Options:     convertSlice(n.Options),
	}
}

func convertCreatedbStmt(n *pg_query.CreatedbStmt) *ast.CreatedbStmt {
	if n == nil {
		return nil
	}
	return &ast.CreatedbStmt{
		Dbname:  makeString(n.Dbname),
		Options: convertSlice(n.Options),
	}
}

func convertCurrentOfExpr(n *pg_query.CurrentOfExpr) *ast.CurrentOfExpr {
	if n == nil {
		return nil
	}
	return &ast.CurrentOfExpr{
		Xpr:         convertNode(n.Xpr),
		Cvarno:      ast.Index(n.Cvarno),
		CursorName:  makeString(n.CursorName),
		CursorParam: int(n.CursorParam),
	}
}

func convertDeallocateStmt(n *pg_query.DeallocateStmt) *ast.DeallocateStmt {
	if n == nil {
		return nil
	}
	return &ast.DeallocateStmt{
		Name: makeString(n.Name),
	}
}

func convertDeclareCursorStmt(n *pg_query.DeclareCursorStmt) *ast.DeclareCursorStmt {
	if n == nil {
		return nil
	}
	return &ast.DeclareCursorStmt{
		Portalname: makeString(n.Portalname),
		Options:    int(n.Options),
		Query:      convertNode(n.Query),
	}
}

func convertDefElem(n *pg_query.DefElem) *ast.DefElem {
	if n == nil {
		return nil
	}
	return &ast.DefElem{
		Defnamespace: makeString(n.Defnamespace),
		Defname:      makeString(n.Defname),
		Arg:          convertNode(n.Arg),
		Defaction:    ast.DefElemAction(n.Defaction),
		Location:     int(n.Location),
	}
}

func convertDefineStmt(n *pg_query.DefineStmt) *ast.DefineStmt {
	if n == nil {
		return nil
	}
	return &ast.DefineStmt{
		Kind:        ast.ObjectType(n.Kind),
		Oldstyle:    n.Oldstyle,
		Defnames:    convertSlice(n.Defnames),
		Args:        convertSlice(n.Args),
		Definition:  convertSlice(n.Definition),
		IfNotExists: n.IfNotExists,
	}
}

func convertDeleteStmt(n *pg_query.DeleteStmt) *ast.DeleteStmt {
	if n == nil {
		return nil
	}
	return &ast.DeleteStmt{
		Relations: &ast.List{
			Items: []ast.Node{convertRangeVar(n.Relation)},
		},
		UsingClause:   convertSlice(n.UsingClause),
		WhereClause:   convertNode(n.WhereClause),
		ReturningList: convertSlice(n.ReturningList),
		WithClause:    convertWithClause(n.WithClause),
	}
}

func convertDiscardStmt(n *pg_query.DiscardStmt) *ast.DiscardStmt {
	if n == nil {
		return nil
	}
	return &ast.DiscardStmt{
		Target: ast.DiscardMode(n.Target),
	}
}

func convertDoStmt(n *pg_query.DoStmt) *ast.DoStmt {
	if n == nil {
		return nil
	}
	return &ast.DoStmt{
		Args: convertSlice(n.Args),
	}
}

func convertDropOwnedStmt(n *pg_query.DropOwnedStmt) *ast.DropOwnedStmt {
	if n == nil {
		return nil
	}
	return &ast.DropOwnedStmt{
		Roles:    convertSlice(n.Roles),
		Behavior: ast.DropBehavior(n.Behavior),
	}
}

func convertDropRoleStmt(n *pg_query.DropRoleStmt) *ast.DropRoleStmt {
	if n == nil {
		return nil
	}
	return &ast.DropRoleStmt{
		Roles:     convertSlice(n.Roles),
		MissingOk: n.MissingOk,
	}
}

func convertDropStmt(n *pg_query.DropStmt) *ast.DropStmt {
	if n == nil {
		return nil
	}
	return &ast.DropStmt{
		Objects:    convertSlice(n.Objects),
		RemoveType: ast.ObjectType(n.RemoveType),
		Behavior:   ast.DropBehavior(n.Behavior),
		MissingOk:  n.MissingOk,
		Concurrent: n.Concurrent,
	}
}

func convertDropSubscriptionStmt(n *pg_query.DropSubscriptionStmt) *ast.DropSubscriptionStmt {
	if n == nil {
		return nil
	}
	return &ast.DropSubscriptionStmt{
		Subname:   makeString(n.Subname),
		MissingOk: n.MissingOk,
		Behavior:  ast.DropBehavior(n.Behavior),
	}
}

func convertDropTableSpaceStmt(n *pg_query.DropTableSpaceStmt) *ast.DropTableSpaceStmt {
	if n == nil {
		return nil
	}
	return &ast.DropTableSpaceStmt{
		Tablespacename: makeString(n.Tablespacename),
		MissingOk:      n.MissingOk,
	}
}

func convertDropUserMappingStmt(n *pg_query.DropUserMappingStmt) *ast.DropUserMappingStmt {
	if n == nil {
		return nil
	}
	return &ast.DropUserMappingStmt{
		User:       convertRoleSpec(n.User),
		Servername: makeString(n.Servername),
		MissingOk:  n.MissingOk,
	}
}

func convertDropdbStmt(n *pg_query.DropdbStmt) *ast.DropdbStmt {
	if n == nil {
		return nil
	}
	return &ast.DropdbStmt{
		Dbname:    makeString(n.Dbname),
		MissingOk: n.MissingOk,
	}
}

func convertExecuteStmt(n *pg_query.ExecuteStmt) *ast.ExecuteStmt {
	if n == nil {
		return nil
	}
	return &ast.ExecuteStmt{
		Name:   makeString(n.Name),
		Params: convertSlice(n.Params),
	}
}

func convertExplainStmt(n *pg_query.ExplainStmt) *ast.ExplainStmt {
	if n == nil {
		return nil
	}
	return &ast.ExplainStmt{
		Query:   convertNode(n.Query),
		Options: convertSlice(n.Options),
	}
}

func convertFetchStmt(n *pg_query.FetchStmt) *ast.FetchStmt {
	if n == nil {
		return nil
	}
	return &ast.FetchStmt{
		Direction:  ast.FetchDirection(n.Direction),
		HowMany:    n.HowMany,
		Portalname: makeString(n.Portalname),
		Ismove:     n.Ismove,
	}
}

func convertFieldSelect(n *pg_query.FieldSelect) *ast.FieldSelect {
	if n == nil {
		return nil
	}
	return &ast.FieldSelect{
		Xpr:          convertNode(n.Xpr),
		Arg:          convertNode(n.Arg),
		Fieldnum:     ast.AttrNumber(n.Fieldnum),
		Resulttype:   ast.Oid(n.Resulttype),
		Resulttypmod: n.Resulttypmod,
		Resultcollid: ast.Oid(n.Resultcollid),
	}
}

func convertFieldStore(n *pg_query.FieldStore) *ast.FieldStore {
	if n == nil {
		return nil
	}
	return &ast.FieldStore{
		Xpr:        convertNode(n.Xpr),
		Arg:        convertNode(n.Arg),
		Newvals:    convertSlice(n.Newvals),
		Fieldnums:  convertSlice(n.Fieldnums),
		Resulttype: ast.Oid(n.Resulttype),
	}
}

func convertFloat(n *pg_query.Float) *ast.Float {
	if n == nil {
		return nil
	}
	return &ast.Float{
		Str: n.Fval,
	}
}

func convertFromExpr(n *pg_query.FromExpr) *ast.FromExpr {
	if n == nil {
		return nil
	}
	return &ast.FromExpr{
		Fromlist: convertSlice(n.Fromlist),
		Quals:    convertNode(n.Quals),
	}
}

func convertFuncCall(n *pg_query.FuncCall) *ast.FuncCall {
	if n == nil {
		return nil
	}
	rel, err := parseRelationFromNodes(n.Funcname)
	if err != nil {
		// TODO: How should we handle errors?
		panic(err)
	}
	return &ast.FuncCall{
		Func:           rel.FuncName(),
		Funcname:       convertSlice(n.Funcname),
		Args:           convertSlice(n.Args),
		AggOrder:       convertSlice(n.AggOrder),
		AggFilter:      convertNode(n.AggFilter),
		AggWithinGroup: n.AggWithinGroup,
		AggStar:        n.AggStar,
		AggDistinct:    n.AggDistinct,
		FuncVariadic:   n.FuncVariadic,
		Over:           convertWindowDef(n.Over),
		Location:       int(n.Location),
	}
}

func convertFuncExpr(n *pg_query.FuncExpr) *ast.FuncExpr {
	if n == nil {
		return nil
	}
	return &ast.FuncExpr{
		Xpr:            convertNode(n.Xpr),
		Funcid:         ast.Oid(n.Funcid),
		Funcresulttype: ast.Oid(n.Funcresulttype),
		Funcretset:     n.Funcretset,
		Funcvariadic:   n.Funcvariadic,
		Funcformat:     ast.CoercionForm(n.Funcformat),
		Funccollid:     ast.Oid(n.Funccollid),
		Inputcollid:    ast.Oid(n.Inputcollid),
		Args:           convertSlice(n.Args),
		Location:       int(n.Location),
	}
}

func convertFunctionParameter(n *pg_query.FunctionParameter) *ast.FunctionParameter {
	if n == nil {
		return nil
	}
	return &ast.FunctionParameter{
		Name:    makeString(n.Name),
		ArgType: convertTypeName(n.ArgType),
		Mode:    ast.FunctionParameterMode(n.Mode),
		Defexpr: convertNode(n.Defexpr),
	}
}

func convertGrantRoleStmt(n *pg_query.GrantRoleStmt) *ast.GrantRoleStmt {
	if n == nil {
		return nil
	}
	return &ast.GrantRoleStmt{
		GrantedRoles: convertSlice(n.GrantedRoles),
		GranteeRoles: convertSlice(n.GranteeRoles),
		IsGrant:      n.IsGrant,
		Grantor:      convertRoleSpec(n.Grantor),
		Behavior:     ast.DropBehavior(n.Behavior),
	}
}

func convertGrantStmt(n *pg_query.GrantStmt) *ast.GrantStmt {
	if n == nil {
		return nil
	}
	return &ast.GrantStmt{
		IsGrant:     n.IsGrant,
		Targtype:    ast.GrantTargetType(n.Targtype),
		Objtype:     ast.GrantObjectType(n.Objtype),
		Objects:     convertSlice(n.Objects),
		Privileges:  convertSlice(n.Privileges),
		Grantees:    convertSlice(n.Grantees),
		GrantOption: n.GrantOption,
		Behavior:    ast.DropBehavior(n.Behavior),
	}
}

func convertGroupingFunc(n *pg_query.GroupingFunc) *ast.GroupingFunc {
	if n == nil {
		return nil
	}
	return &ast.GroupingFunc{
		Xpr:         convertNode(n.Xpr),
		Args:        convertSlice(n.Args),
		Refs:        convertSlice(n.Refs),
		Agglevelsup: ast.Index(n.Agglevelsup),
		Location:    int(n.Location),
	}
}

func convertGroupingSet(n *pg_query.GroupingSet) *ast.GroupingSet {
	if n == nil {
		return nil
	}
	return &ast.GroupingSet{
		Kind:     ast.GroupingSetKind(n.Kind),
		Content:  convertSlice(n.Content),
		Location: int(n.Location),
	}
}

func convertImportForeignSchemaStmt(n *pg_query.ImportForeignSchemaStmt) *ast.ImportForeignSchemaStmt {
	if n == nil {
		return nil
	}
	return &ast.ImportForeignSchemaStmt{
		ServerName:   makeString(n.ServerName),
		RemoteSchema: makeString(n.RemoteSchema),
		LocalSchema:  makeString(n.LocalSchema),
		ListType:     ast.ImportForeignSchemaType(n.ListType),
		TableList:    convertSlice(n.TableList),
		Options:      convertSlice(n.Options),
	}
}

func convertIndexElem(n *pg_query.IndexElem) *ast.IndexElem {
	if n == nil {
		return nil
	}
	return &ast.IndexElem{
		Name:          makeString(n.Name),
		Expr:          convertNode(n.Expr),
		Indexcolname:  makeString(n.Indexcolname),
		Collation:     convertSlice(n.Collation),
		Opclass:       convertSlice(n.Opclass),
		Ordering:      ast.SortByDir(n.Ordering),
		NullsOrdering: ast.SortByNulls(n.NullsOrdering),
	}
}

func convertIndexStmt(n *pg_query.IndexStmt) *ast.IndexStmt {
	if n == nil {
		return nil
	}
	return &ast.IndexStmt{
		Idxname:        makeString(n.Idxname),
		Relation:       convertRangeVar(n.Relation),
		AccessMethod:   makeString(n.AccessMethod),
		TableSpace:     makeString(n.TableSpace),
		IndexParams:    convertSlice(n.IndexParams),
		Options:        convertSlice(n.Options),
		WhereClause:    convertNode(n.WhereClause),
		ExcludeOpNames: convertSlice(n.ExcludeOpNames),
		Idxcomment:     makeString(n.Idxcomment),
		IndexOid:       ast.Oid(n.IndexOid),
		Unique:         n.Unique,
		Primary:        n.Primary,
		Isconstraint:   n.Isconstraint,
		Deferrable:     n.Deferrable,
		Initdeferred:   n.Initdeferred,
		Transformed:    n.Transformed,
		Concurrent:     n.Concurrent,
		IfNotExists:    n.IfNotExists,
	}
}

func convertInferClause(n *pg_query.InferClause) *ast.InferClause {
	if n == nil {
		return nil
	}
	return &ast.InferClause{
		IndexElems:  convertSlice(n.IndexElems),
		WhereClause: convertNode(n.WhereClause),
		Conname:     makeString(n.Conname),
		Location:    int(n.Location),
	}
}

func convertInferenceElem(n *pg_query.InferenceElem) *ast.InferenceElem {
	if n == nil {
		return nil
	}
	return &ast.InferenceElem{
		Xpr:          convertNode(n.Xpr),
		Expr:         convertNode(n.Expr),
		Infercollid:  ast.Oid(n.Infercollid),
		Inferopclass: ast.Oid(n.Inferopclass),
	}
}

func convertInlineCodeBlock(n *pg_query.InlineCodeBlock) *ast.InlineCodeBlock {
	if n == nil {
		return nil
	}
	return &ast.InlineCodeBlock{
		SourceText:    makeString(n.SourceText),
		LangOid:       ast.Oid(n.LangOid),
		LangIsTrusted: n.LangIsTrusted,
	}
}

func convertInsertStmt(n *pg_query.InsertStmt) *ast.InsertStmt {
	if n == nil {
		return nil
	}
	return &ast.InsertStmt{
		Relation:         convertRangeVar(n.Relation),
		Cols:             convertSlice(n.Cols),
		SelectStmt:       convertNode(n.SelectStmt),
		OnConflictClause: convertOnConflictClause(n.OnConflictClause),
		ReturningList:    convertSlice(n.ReturningList),
		WithClause:       convertWithClause(n.WithClause),
		Override:         ast.OverridingKind(n.Override),
	}
}

func convertInteger(n *pg_query.Integer) *ast.Integer {
	if n == nil {
		return nil
	}
	return &ast.Integer{
		Ival: int64(n.Ival),
	}
}

func convertJoinExpr(n *pg_query.JoinExpr) *ast.JoinExpr {
	if n == nil {
		return nil
	}
	return &ast.JoinExpr{
		Jointype:    ast.JoinType(n.Jointype),
		IsNatural:   n.IsNatural,
		Larg:        convertNode(n.Larg),
		Rarg:        convertNode(n.Rarg),
		UsingClause: convertSlice(n.UsingClause),
		Quals:       convertNode(n.Quals),
		Alias:       convertAlias(n.Alias),
		Rtindex:     int(n.Rtindex),
	}
}

func convertListenStmt(n *pg_query.ListenStmt) *ast.ListenStmt {
	if n == nil {
		return nil
	}
	return &ast.ListenStmt{
		Conditionname: makeString(n.Conditionname),
	}
}

func convertLoadStmt(n *pg_query.LoadStmt) *ast.LoadStmt {
	if n == nil {
		return nil
	}
	return &ast.LoadStmt{
		Filename: makeString(n.Filename),
	}
}

func convertLockStmt(n *pg_query.LockStmt) *ast.LockStmt {
	if n == nil {
		return nil
	}
	return &ast.LockStmt{
		Relations: convertSlice(n.Relations),
		Mode:      int(n.Mode),
		Nowait:    n.Nowait,
	}
}

func convertLockingClause(n *pg_query.LockingClause) *ast.LockingClause {
	if n == nil {
		return nil
	}
	return &ast.LockingClause{
		LockedRels: convertSlice(n.LockedRels),
		Strength:   ast.LockClauseStrength(n.Strength),
		WaitPolicy: ast.LockWaitPolicy(n.WaitPolicy),
	}
}

func convertMinMaxExpr(n *pg_query.MinMaxExpr) *ast.MinMaxExpr {
	if n == nil {
		return nil
	}
	return &ast.MinMaxExpr{
		Xpr:          convertNode(n.Xpr),
		Minmaxtype:   ast.Oid(n.Minmaxtype),
		Minmaxcollid: ast.Oid(n.Minmaxcollid),
		Inputcollid:  ast.Oid(n.Inputcollid),
		Op:           ast.MinMaxOp(n.Op),
		Args:         convertSlice(n.Args),
		Location:     int(n.Location),
	}
}

func convertMultiAssignRef(n *pg_query.MultiAssignRef) *ast.MultiAssignRef {
	if n == nil {
		return nil
	}
	return &ast.MultiAssignRef{
		Source:   convertNode(n.Source),
		Colno:    int(n.Colno),
		Ncolumns: int(n.Ncolumns),
	}
}

func convertNamedArgExpr(n *pg_query.NamedArgExpr) *ast.NamedArgExpr {
	if n == nil {
		return nil
	}
	return &ast.NamedArgExpr{
		Xpr:       convertNode(n.Xpr),
		Arg:       convertNode(n.Arg),
		Name:      makeString(n.Name),
		Argnumber: int(n.Argnumber),
		Location:  int(n.Location),
	}
}

func convertNextValueExpr(n *pg_query.NextValueExpr) *ast.NextValueExpr {
	if n == nil {
		return nil
	}
	return &ast.NextValueExpr{
		Xpr:    convertNode(n.Xpr),
		Seqid:  ast.Oid(n.Seqid),
		TypeId: ast.Oid(n.TypeId),
	}
}

func convertNotifyStmt(n *pg_query.NotifyStmt) *ast.NotifyStmt {
	if n == nil {
		return nil
	}
	return &ast.NotifyStmt{
		Conditionname: makeString(n.Conditionname),
		Payload:       makeString(n.Payload),
	}
}

func convertNullTest(n *pg_query.NullTest) *ast.NullTest {
	if n == nil {
		return nil
	}
	return &ast.NullTest{
		Xpr:          convertNode(n.Xpr),
		Arg:          convertNode(n.Arg),
		Nulltesttype: ast.NullTestType(n.Nulltesttype),
		Argisrow:     n.Argisrow,
		Location:     int(n.Location),
	}
}

func convertObjectWithArgs(n *pg_query.ObjectWithArgs) *ast.ObjectWithArgs {
	if n == nil {
		return nil
	}
	return &ast.ObjectWithArgs{
		Objname:         convertSlice(n.Objname),
		Objargs:         convertSlice(n.Objargs),
		ArgsUnspecified: n.ArgsUnspecified,
	}
}

func convertOnConflictClause(n *pg_query.OnConflictClause) *ast.OnConflictClause {
	if n == nil {
		return nil
	}
	return &ast.OnConflictClause{
		Action:      ast.OnConflictAction(n.Action),
		Infer:       convertInferClause(n.Infer),
		TargetList:  convertSlice(n.TargetList),
		WhereClause: convertNode(n.WhereClause),
		Location:    int(n.Location),
	}
}

func convertOnConflictExpr(n *pg_query.OnConflictExpr) *ast.OnConflictExpr {
	if n == nil {
		return nil
	}
	return &ast.OnConflictExpr{
		Action:          ast.OnConflictAction(n.Action),
		ArbiterElems:    convertSlice(n.ArbiterElems),
		ArbiterWhere:    convertNode(n.ArbiterWhere),
		Constraint:      ast.Oid(n.Constraint),
		OnConflictSet:   convertSlice(n.OnConflictSet),
		OnConflictWhere: convertNode(n.OnConflictWhere),
		ExclRelIndex:    int(n.ExclRelIndex),
		ExclRelTlist:    convertSlice(n.ExclRelTlist),
	}
}

func convertOpExpr(n *pg_query.OpExpr) *ast.OpExpr {
	if n == nil {
		return nil
	}
	return &ast.OpExpr{
		Xpr:          convertNode(n.Xpr),
		Opno:         ast.Oid(n.Opno),
		Opresulttype: ast.Oid(n.Opresulttype),
		Opretset:     n.Opretset,
		Opcollid:     ast.Oid(n.Opcollid),
		Inputcollid:  ast.Oid(n.Inputcollid),
		Args:         convertSlice(n.Args),
		Location:     int(n.Location),
	}
}

func convertParam(n *pg_query.Param) *ast.Param {
	if n == nil {
		return nil
	}
	return &ast.Param{
		Xpr:         convertNode(n.Xpr),
		Paramkind:   ast.ParamKind(n.Paramkind),
		Paramid:     int(n.Paramid),
		Paramtype:   ast.Oid(n.Paramtype),
		Paramtypmod: n.Paramtypmod,
		Paramcollid: ast.Oid(n.Paramcollid),
		Location:    int(n.Location),
	}
}

func convertParamRef(n *pg_query.ParamRef) *ast.ParamRef {
	if n == nil {
		return nil
	}
	var dollar bool
	if n.Number != 0 {
		dollar = true
	}
	return &ast.ParamRef{
		Dollar:   dollar,
		Number:   int(n.Number),
		Location: int(n.Location),
	}
}

func convertPartitionBoundSpec(n *pg_query.PartitionBoundSpec) *ast.PartitionBoundSpec {
	if n == nil {
		return nil
	}
	return &ast.PartitionBoundSpec{
		Strategy:    makeByte(n.Strategy),
		Listdatums:  convertSlice(n.Listdatums),
		Lowerdatums: convertSlice(n.Lowerdatums),
		Upperdatums: convertSlice(n.Upperdatums),
		Location:    int(n.Location),
	}
}

func convertPartitionCmd(n *pg_query.PartitionCmd) *ast.PartitionCmd {
	if n == nil {
		return nil
	}
	return &ast.PartitionCmd{
		Name:  convertRangeVar(n.Name),
		Bound: convertPartitionBoundSpec(n.Bound),
	}
}

func convertPartitionElem(n *pg_query.PartitionElem) *ast.PartitionElem {
	if n == nil {
		return nil
	}
	return &ast.PartitionElem{
		Name:      makeString(n.Name),
		Expr:      convertNode(n.Expr),
		Collation: convertSlice(n.Collation),
		Opclass:   convertSlice(n.Opclass),
		Location:  int(n.Location),
	}
}

func convertPartitionRangeDatum(n *pg_query.PartitionRangeDatum) *ast.PartitionRangeDatum {
	if n == nil {
		return nil
	}
	return &ast.PartitionRangeDatum{
		Kind:     ast.PartitionRangeDatumKind(n.Kind),
		Value:    convertNode(n.Value),
		Location: int(n.Location),
	}
}

func convertPartitionSpec(n *pg_query.PartitionSpec) *ast.PartitionSpec {
	if n == nil {
		return nil
	}
	return &ast.PartitionSpec{
		Strategy:   makeString(n.Strategy.String()),
		PartParams: convertSlice(n.PartParams),
		Location:   int(n.Location),
	}
}

func convertPrepareStmt(n *pg_query.PrepareStmt) *ast.PrepareStmt {
	if n == nil {
		return nil
	}
	return &ast.PrepareStmt{
		Name:     makeString(n.Name),
		Argtypes: convertSlice(n.Argtypes),
		Query:    convertNode(n.Query),
	}
}

func convertQuery(n *pg_query.Query) *ast.Query {
	if n == nil {
		return nil
	}
	return &ast.Query{
		CommandType:      ast.CmdType(n.CommandType),
		QuerySource:      ast.QuerySource(n.QuerySource),
		CanSetTag:        n.CanSetTag,
		UtilityStmt:      convertNode(n.UtilityStmt),
		ResultRelation:   int(n.ResultRelation),
		HasAggs:          n.HasAggs,
		HasWindowFuncs:   n.HasWindowFuncs,
		HasTargetSrfs:    n.HasTargetSrfs,
		HasSubLinks:      n.HasSubLinks,
		HasDistinctOn:    n.HasDistinctOn,
		HasRecursive:     n.HasRecursive,
		HasModifyingCte:  n.HasModifyingCte,
		HasForUpdate:     n.HasForUpdate,
		HasRowSecurity:   n.HasRowSecurity,
		CteList:          convertSlice(n.CteList),
		Rtable:           convertSlice(n.Rtable),
		Jointree:         convertFromExpr(n.Jointree),
		TargetList:       convertSlice(n.TargetList),
		Override:         ast.OverridingKind(n.Override),
		OnConflict:       convertOnConflictExpr(n.OnConflict),
		ReturningList:    convertSlice(n.ReturningList),
		GroupClause:      convertSlice(n.GroupClause),
		GroupingSets:     convertSlice(n.GroupingSets),
		HavingQual:       convertNode(n.HavingQual),
		WindowClause:     convertSlice(n.WindowClause),
		DistinctClause:   convertSlice(n.DistinctClause),
		SortClause:       convertSlice(n.SortClause),
		LimitOffset:      convertNode(n.LimitOffset),
		LimitCount:       convertNode(n.LimitCount),
		RowMarks:         convertSlice(n.RowMarks),
		SetOperations:    convertNode(n.SetOperations),
		ConstraintDeps:   convertSlice(n.ConstraintDeps),
		WithCheckOptions: convertSlice(n.WithCheckOptions),
		StmtLocation:     int(n.StmtLocation),
		StmtLen:          int(n.StmtLen),
	}
}

func convertRangeFunction(n *pg_query.RangeFunction) *ast.RangeFunction {
	if n == nil {
		return nil
	}
	return &ast.RangeFunction{
		Lateral:    n.Lateral,
		Ordinality: n.Ordinality,
		IsRowsfrom: n.IsRowsfrom,
		Functions:  convertSlice(n.Functions),
		Alias:      convertAlias(n.Alias),
		Coldeflist: convertSlice(n.Coldeflist),
	}
}

func convertRangeSubselect(n *pg_query.RangeSubselect) *ast.RangeSubselect {
	if n == nil {
		return nil
	}
	return &ast.RangeSubselect{
		Lateral:  n.Lateral,
		Subquery: convertNode(n.Subquery),
		Alias:    convertAlias(n.Alias),
	}
}

func convertRangeTableFunc(n *pg_query.RangeTableFunc) *ast.RangeTableFunc {
	if n == nil {
		return nil
	}
	return &ast.RangeTableFunc{
		Lateral:    n.Lateral,
		Docexpr:    convertNode(n.Docexpr),
		Rowexpr:    convertNode(n.Rowexpr),
		Namespaces: convertSlice(n.Namespaces),
		Columns:    convertSlice(n.Columns),
		Alias:      convertAlias(n.Alias),
		Location:   int(n.Location),
	}
}

func convertRangeTableFuncCol(n *pg_query.RangeTableFuncCol) *ast.RangeTableFuncCol {
	if n == nil {
		return nil
	}
	return &ast.RangeTableFuncCol{
		Colname:       makeString(n.Colname),
		TypeName:      convertTypeName(n.TypeName),
		ForOrdinality: n.ForOrdinality,
		IsNotNull:     n.IsNotNull,
		Colexpr:       convertNode(n.Colexpr),
		Coldefexpr:    convertNode(n.Coldefexpr),
		Location:      int(n.Location),
	}
}

func convertRangeTableSample(n *pg_query.RangeTableSample) *ast.RangeTableSample {
	if n == nil {
		return nil
	}
	return &ast.RangeTableSample{
		Relation:   convertNode(n.Relation),
		Method:     convertSlice(n.Method),
		Args:       convertSlice(n.Args),
		Repeatable: convertNode(n.Repeatable),
		Location:   int(n.Location),
	}
}

func convertRangeTblEntry(n *pg_query.RangeTblEntry) *ast.RangeTblEntry {
	if n == nil {
		return nil
	}
	return &ast.RangeTblEntry{
		Rtekind:         ast.RTEKind(n.Rtekind),
		Relid:           ast.Oid(n.Relid),
		Relkind:         makeByte(n.Relkind),
		Tablesample:     convertTableSampleClause(n.Tablesample),
		Subquery:        convertQuery(n.Subquery),
		SecurityBarrier: n.SecurityBarrier,
		Jointype:        ast.JoinType(n.Jointype),
		Joinaliasvars:   convertSlice(n.Joinaliasvars),
		Functions:       convertSlice(n.Functions),
		Funcordinality:  n.Funcordinality,
		Tablefunc:       convertTableFunc(n.Tablefunc),
		ValuesLists:     convertSlice(n.ValuesLists),
		Ctename:         makeString(n.Ctename),
		Ctelevelsup:     ast.Index(n.Ctelevelsup),
		SelfReference:   n.SelfReference,
		Coltypes:        convertSlice(n.Coltypes),
		Coltypmods:      convertSlice(n.Coltypmods),
		Colcollations:   convertSlice(n.Colcollations),
		Enrname:         makeString(n.Enrname),
		Enrtuples:       n.Enrtuples,
		Alias:           convertAlias(n.Alias),
		Eref:            convertAlias(n.Eref),
		Lateral:         n.Lateral,
		Inh:             n.Inh,
		InFromCl:        n.InFromCl,
		SecurityQuals:   convertSlice(n.SecurityQuals),
	}
}

func convertRangeTblFunction(n *pg_query.RangeTblFunction) *ast.RangeTblFunction {
	if n == nil {
		return nil
	}
	return &ast.RangeTblFunction{
		Funcexpr:          convertNode(n.Funcexpr),
		Funccolcount:      int(n.Funccolcount),
		Funccolnames:      convertSlice(n.Funccolnames),
		Funccoltypes:      convertSlice(n.Funccoltypes),
		Funccoltypmods:    convertSlice(n.Funccoltypmods),
		Funccolcollations: convertSlice(n.Funccolcollations),
		Funcparams:        makeUint32Slice(n.Funcparams),
	}
}

func convertRangeTblRef(n *pg_query.RangeTblRef) *ast.RangeTblRef {
	if n == nil {
		return nil
	}
	return &ast.RangeTblRef{
		Rtindex: int(n.Rtindex),
	}
}

func convertRangeVar(n *pg_query.RangeVar) *ast.RangeVar {
	if n == nil {
		return nil
	}
	return &ast.RangeVar{
		Catalogname:    makeString(n.Catalogname),
		Schemaname:     makeString(n.Schemaname),
		Relname:        makeString(n.Relname),
		Inh:            n.Inh,
		Relpersistence: makeByte(n.Relpersistence),
		Alias:          convertAlias(n.Alias),
		Location:       int(n.Location),
	}
}

func convertRawStmt(n *pg_query.RawStmt) *ast.RawStmt {
	if n == nil {
		return nil
	}
	return &ast.RawStmt{
		Stmt:         convertNode(n.Stmt),
		StmtLocation: int(n.StmtLocation),
		StmtLen:      int(n.StmtLen),
	}
}

func convertReassignOwnedStmt(n *pg_query.ReassignOwnedStmt) *ast.ReassignOwnedStmt {
	if n == nil {
		return nil
	}
	return &ast.ReassignOwnedStmt{
		Roles:   convertSlice(n.Roles),
		Newrole: convertRoleSpec(n.Newrole),
	}
}

func convertRefreshMatViewStmt(n *pg_query.RefreshMatViewStmt) *ast.RefreshMatViewStmt {
	if n == nil {
		return nil
	}
	return &ast.RefreshMatViewStmt{
		Concurrent: n.Concurrent,
		SkipData:   n.SkipData,
		Relation:   convertRangeVar(n.Relation),
	}
}

func convertReindexStmt(n *pg_query.ReindexStmt) *ast.ReindexStmt {
	if n == nil {
		return nil
	}
	return &ast.ReindexStmt{
		Kind:     ast.ReindexObjectType(n.Kind),
		Relation: convertRangeVar(n.Relation),
		Name:     makeString(n.Name),
		// Options:  int(n.Options), TODO: Support params
	}
}

func convertRelabelType(n *pg_query.RelabelType) *ast.RelabelType {
	if n == nil {
		return nil
	}
	return &ast.RelabelType{
		Xpr:           convertNode(n.Xpr),
		Arg:           convertNode(n.Arg),
		Resulttype:    ast.Oid(n.Resulttype),
		Resulttypmod:  n.Resulttypmod,
		Resultcollid:  ast.Oid(n.Resultcollid),
		Relabelformat: ast.CoercionForm(n.Relabelformat),
		Location:      int(n.Location),
	}
}

func convertRenameStmt(n *pg_query.RenameStmt) *ast.RenameStmt {
	if n == nil {
		return nil
	}
	return &ast.RenameStmt{
		RenameType:   ast.ObjectType(n.RenameType),
		RelationType: ast.ObjectType(n.RelationType),
		Relation:     convertRangeVar(n.Relation),
		Object:       convertNode(n.Object),
		Subname:      makeString(n.Subname),
		Newname:      makeString(n.Newname),
		Behavior:     ast.DropBehavior(n.Behavior),
		MissingOk:    n.MissingOk,
	}
}

func convertReplicaIdentityStmt(n *pg_query.ReplicaIdentityStmt) *ast.ReplicaIdentityStmt {
	if n == nil {
		return nil
	}
	return &ast.ReplicaIdentityStmt{
		IdentityType: makeByte(n.IdentityType),
		Name:         makeString(n.Name),
	}
}

func convertResTarget(n *pg_query.ResTarget) *ast.ResTarget {
	if n == nil {
		return nil
	}
	return &ast.ResTarget{
		Name:        makeString(n.Name),
		Indirection: convertSlice(n.Indirection),
		Val:         convertNode(n.Val),
		Location:    int(n.Location),
	}
}

func convertRoleSpec(n *pg_query.RoleSpec) *ast.RoleSpec {
	if n == nil {
		return nil
	}
	return &ast.RoleSpec{
		Roletype: ast.RoleSpecType(n.Roletype),
		Rolename: makeString(n.Rolename),
		Location: int(n.Location),
	}
}

func convertRowCompareExpr(n *pg_query.RowCompareExpr) *ast.RowCompareExpr {
	if n == nil {
		return nil
	}
	return &ast.RowCompareExpr{
		Xpr:          convertNode(n.Xpr),
		Rctype:       ast.RowCompareType(n.Rctype),
		Opnos:        convertSlice(n.Opnos),
		Opfamilies:   convertSlice(n.Opfamilies),
		Inputcollids: convertSlice(n.Inputcollids),
		Largs:        convertSlice(n.Largs),
		Rargs:        convertSlice(n.Rargs),
	}
}

func convertRowExpr(n *pg_query.RowExpr) *ast.RowExpr {
	if n == nil {
		return nil
	}
	return &ast.RowExpr{
		Xpr:       convertNode(n.Xpr),
		Args:      convertSlice(n.Args),
		RowTypeid: ast.Oid(n.RowTypeid),
		RowFormat: ast.CoercionForm(n.RowFormat),
		Colnames:  convertSlice(n.Colnames),
		Location:  int(n.Location),
	}
}

func convertRowMarkClause(n *pg_query.RowMarkClause) *ast.RowMarkClause {
	if n == nil {
		return nil
	}
	return &ast.RowMarkClause{
		Rti:        ast.Index(n.Rti),
		Strength:   ast.LockClauseStrength(n.Strength),
		WaitPolicy: ast.LockWaitPolicy(n.WaitPolicy),
		PushedDown: n.PushedDown,
	}
}

func convertRuleStmt(n *pg_query.RuleStmt) *ast.RuleStmt {
	if n == nil {
		return nil
	}
	return &ast.RuleStmt{
		Relation:    convertRangeVar(n.Relation),
		Rulename:    makeString(n.Rulename),
		WhereClause: convertNode(n.WhereClause),
		Event:       ast.CmdType(n.Event),
		Instead:     n.Instead,
		Actions:     convertSlice(n.Actions),
		Replace:     n.Replace,
	}
}

func convertSQLValueFunction(n *pg_query.SQLValueFunction) *ast.SQLValueFunction {
	if n == nil {
		return nil
	}
	return &ast.SQLValueFunction{
		Xpr:      convertNode(n.Xpr),
		Op:       ast.SQLValueFunctionOp(n.Op),
		Type:     ast.Oid(n.Type),
		Typmod:   n.Typmod,
		Location: int(n.Location),
	}
}

func convertScalarArrayOpExpr(n *pg_query.ScalarArrayOpExpr) *ast.ScalarArrayOpExpr {
	if n == nil {
		return nil
	}
	return &ast.ScalarArrayOpExpr{
		Xpr:         convertNode(n.Xpr),
		Opno:        ast.Oid(n.Opno),
		UseOr:       n.UseOr,
		Inputcollid: ast.Oid(n.Inputcollid),
		Args:        convertSlice(n.Args),
		Location:    int(n.Location),
	}
}

func convertSecLabelStmt(n *pg_query.SecLabelStmt) *ast.SecLabelStmt {
	if n == nil {
		return nil
	}
	return &ast.SecLabelStmt{
		Objtype:  ast.ObjectType(n.Objtype),
		Object:   convertNode(n.Object),
		Provider: makeString(n.Provider),
		Label:    makeString(n.Label),
	}
}

func convertSelectStmt(n *pg_query.SelectStmt) (Statement, error) {
	if n == nil {
		return nil
	}
	return &ast.SelectStmt{
		// TODO
		//TargetList:     convertSlice(n.TargetList),
		FromClause:     convertSlice(n.FromClause),
		WhereClause:    convertNode(n.WhereClause),
		GroupClause:    convertSlice(n.GroupClause),
		HavingClause:   convertNode(n.HavingClause),
		WindowClause:   convertSlice(n.WindowClause),
		ValuesLists:    convertSlice(n.ValuesLists),
		SortClause:     convertSlice(n.SortClause),
		LimitOffset:    convertNode(n.LimitOffset),
		LimitCount:     convertNode(n.LimitCount),
		LockingClause:  convertSlice(n.LockingClause),
		WithClause:     convertWithClause(n.WithClause),
		Op:             op,
		All:            n.All,
		Larg:           convertSelectStmt(n.Larg),
		Rarg:           convertSelectStmt(n.Rarg),
	}
}

func convertSetOperationStmt(n *pg_query.SetOperationStmt) *ast.SetOperationStmt {
	if n == nil {
		return nil
	}
	op, err := convertSetOperation(n.Op)
	if err != nil {
		panic(err)
	}
	return &ast.SetOperationStmt{
		Op:            op,
		All:           n.All,
		Larg:          convertNode(n.Larg),
		Rarg:          convertNode(n.Rarg),
		ColTypes:      convertSlice(n.ColTypes),
		ColTypmods:    convertSlice(n.ColTypmods),
		ColCollations: convertSlice(n.ColCollations),
		GroupClauses:  convertSlice(n.GroupClauses),
	}
}

func convertSetToDefault(n *pg_query.SetToDefault) *ast.SetToDefault {
	if n == nil {
		return nil
	}
	return &ast.SetToDefault{
		Xpr:       convertNode(n.Xpr),
		TypeId:    ast.Oid(n.TypeId),
		TypeMod:   n.TypeMod,
		Collation: ast.Oid(n.Collation),
		Location:  int(n.Location),
	}
}

func convertSortBy(n *pg_query.SortBy) *ast.SortBy {
	if n == nil {
		return nil
	}
	return &ast.SortBy{
		Node:        convertNode(n.Node),
		SortbyDir:   ast.SortByDir(n.SortbyDir),
		SortbyNulls: ast.SortByNulls(n.SortbyNulls),
		UseOp:       convertSlice(n.UseOp),
		Location:    int(n.Location),
	}
}

func convertSortGroupClause(n *pg_query.SortGroupClause) *ast.SortGroupClause {
	if n == nil {
		return nil
	}
	return &ast.SortGroupClause{
		TleSortGroupRef: ast.Index(n.TleSortGroupRef),
		Eqop:            ast.Oid(n.Eqop),
		Sortop:          ast.Oid(n.Sortop),
		NullsFirst:      n.NullsFirst,
		Hashable:        n.Hashable,
	}
}

func convertString(n *pg_query.String) *ast.String {
	if n == nil {
		return nil
	}
	return &ast.String{
		Str: n.Sval,
	}
}

func convertSubLink(n *pg_query.SubLink) *ast.SubLink {
	if n == nil {
		return nil
	}
	slt, err := convertSubLinkType(n.SubLinkType)
	if err != nil {
		panic(err)
	}
	return &ast.SubLink{
		Xpr:         convertNode(n.Xpr),
		SubLinkType: slt,
		SubLinkId:   int(n.SubLinkId),
		Testexpr:    convertNode(n.Testexpr),
		OperName:    convertSlice(n.OperName),
		Subselect:   convertNode(n.Subselect),
		Location:    int(n.Location),
	}
}

func convertSubPlan(n *pg_query.SubPlan) *ast.SubPlan {
	if n == nil {
		return nil
	}
	slt, err := convertSubLinkType(n.SubLinkType)
	if err != nil {
		panic(err)
	}
	return &ast.SubPlan{
		Xpr:               convertNode(n.Xpr),
		SubLinkType:       slt,
		Testexpr:          convertNode(n.Testexpr),
		ParamIds:          convertSlice(n.ParamIds),
		PlanId:            int(n.PlanId),
		PlanName:          makeString(n.PlanName),
		FirstColType:      ast.Oid(n.FirstColType),
		FirstColTypmod:    n.FirstColTypmod,
		FirstColCollation: ast.Oid(n.FirstColCollation),
		UseHashTable:      n.UseHashTable,
		UnknownEqFalse:    n.UnknownEqFalse,
		ParallelSafe:      n.ParallelSafe,
		SetParam:          convertSlice(n.SetParam),
		ParParam:          convertSlice(n.ParParam),
		Args:              convertSlice(n.Args),
		StartupCost:       ast.Cost(n.StartupCost),
		PerCallCost:       ast.Cost(n.PerCallCost),
	}
}

func convertTableFunc(n *pg_query.TableFunc) *ast.TableFunc {
	if n == nil {
		return nil
	}
	return &ast.TableFunc{
		NsUris:        convertSlice(n.NsUris),
		NsNames:       convertSlice(n.NsNames),
		Docexpr:       convertNode(n.Docexpr),
		Rowexpr:       convertNode(n.Rowexpr),
		Colnames:      convertSlice(n.Colnames),
		Coltypes:      convertSlice(n.Coltypes),
		Coltypmods:    convertSlice(n.Coltypmods),
		Colcollations: convertSlice(n.Colcollations),
		Colexprs:      convertSlice(n.Colexprs),
		Coldefexprs:   convertSlice(n.Coldefexprs),
		Notnulls:      makeUint32Slice(n.Notnulls),
		Ordinalitycol: int(n.Ordinalitycol),
		Location:      int(n.Location),
	}
}

func convertTableLikeClause(n *pg_query.TableLikeClause) *ast.TableLikeClause {
	if n == nil {
		return nil
	}
	return &ast.TableLikeClause{
		Relation: convertRangeVar(n.Relation),
		Options:  n.Options,
	}
}

func convertTableSampleClause(n *pg_query.TableSampleClause) *ast.TableSampleClause {
	if n == nil {
		return nil
	}
	return &ast.TableSampleClause{
		Tsmhandler: ast.Oid(n.Tsmhandler),
		Args:       convertSlice(n.Args),
		Repeatable: convertNode(n.Repeatable),
	}
}

func convertTargetEntry(n *pg_query.TargetEntry) *ast.TargetEntry {
	if n == nil {
		return nil
	}
	return &ast.TargetEntry{
		Xpr:             convertNode(n.Xpr),
		Expr:            convertNode(n.Expr),
		Resno:           ast.AttrNumber(n.Resno),
		Resname:         makeString(n.Resname),
		Ressortgroupref: ast.Index(n.Ressortgroupref),
		Resorigtbl:      ast.Oid(n.Resorigtbl),
		Resorigcol:      ast.AttrNumber(n.Resorigcol),
		Resjunk:         n.Resjunk,
	}
}

func convertTransactionStmt(n *pg_query.TransactionStmt) *ast.TransactionStmt {
	if n == nil {
		return nil
	}
	return &ast.TransactionStmt{
		Kind:    ast.TransactionStmtKind(n.Kind),
		Options: convertSlice(n.Options),
		Gid:     makeString(n.Gid),
	}
}

func convertTriggerTransition(n *pg_query.TriggerTransition) *ast.TriggerTransition {
	if n == nil {
		return nil
	}
	return &ast.TriggerTransition{
		Name:    makeString(n.Name),
		IsNew:   n.IsNew,
		IsTable: n.IsTable,
	}
}

func convertTruncateStmt(n *pg_query.TruncateStmt) *ast.TruncateStmt {
	if n == nil {
		return nil
	}
	return &ast.TruncateStmt{
		Relations:   convertSlice(n.Relations),
		RestartSeqs: n.RestartSeqs,
		Behavior:    ast.DropBehavior(n.Behavior),
	}
}

func convertTypeCast(n *pg_query.TypeCast) *ast.TypeCast {
	if n == nil {
		return nil
	}
	return &ast.TypeCast{
		Arg:      convertNode(n.Arg),
		TypeName: convertTypeName(n.TypeName),
		Location: int(n.Location),
	}
}

func convertTypeName(n *pg_query.TypeName) *ast.TypeName {
	if n == nil {
		return nil
	}
	return &ast.TypeName{
		Names:       convertSlice(n.Names),
		TypeOid:     ast.Oid(n.TypeOid),
		Setof:       n.Setof,
		PctType:     n.PctType,
		Typmods:     convertSlice(n.Typmods),
		Typemod:     n.Typemod,
		ArrayBounds: convertSlice(n.ArrayBounds),
		Location:    int(n.Location),
	}
}

func convertUnlistenStmt(n *pg_query.UnlistenStmt) *ast.UnlistenStmt {
	if n == nil {
		return nil
	}
	return &ast.UnlistenStmt{
		Conditionname: makeString(n.Conditionname),
	}
}

func convertUpdateStmt(n *pg_query.UpdateStmt) *ast.UpdateStmt {
	if n == nil {
		return nil
	}

	return &ast.UpdateStmt{
		Relations: &ast.List{
			Items: []ast.Node{convertRangeVar(n.Relation)},
		},
		TargetList:    convertSlice(n.TargetList),
		WhereClause:   convertNode(n.WhereClause),
		FromClause:    convertSlice(n.FromClause),
		ReturningList: convertSlice(n.ReturningList),
		WithClause:    convertWithClause(n.WithClause),
	}
}

func convertVacuumStmt(n *pg_query.VacuumStmt) *ast.VacuumStmt {
	if n == nil {
		return nil
	}
	return &ast.VacuumStmt{
		// FIXME: The VacuumStmt node has changed quite a bit
		// Options:  n.Options
		// Relation: convertRangeVar(n.Relation),
		// VaCols:   convertSlice(n.VaCols),
	}
}

func convertVar(n *pg_query.Var) *ast.Var {
	if n == nil {
		return nil
	}
	return &ast.Var{
		Xpr:         convertNode(n.Xpr),
		Varno:       ast.Index(n.Varno),
		Varattno:    ast.AttrNumber(n.Varattno),
		Vartype:     ast.Oid(n.Vartype),
		Vartypmod:   n.Vartypmod,
		Varcollid:   ast.Oid(n.Varcollid),
		Varlevelsup: ast.Index(n.Varlevelsup),
		Location:    int(n.Location),
	}
}

func convertVariableSetStmt(n *pg_query.VariableSetStmt) *ast.VariableSetStmt {
	if n == nil {
		return nil
	}
	return &ast.VariableSetStmt{
		Kind:    ast.VariableSetKind(n.Kind),
		Name:    makeString(n.Name),
		Args:    convertSlice(n.Args),
		IsLocal: n.IsLocal,
	}
}

func convertVariableShowStmt(n *pg_query.VariableShowStmt) *ast.VariableShowStmt {
	if n == nil {
		return nil
	}
	return &ast.VariableShowStmt{
		Name: makeString(n.Name),
	}
}

func convertViewStmt(n *pg_query.ViewStmt) *ast.ViewStmt {
	if n == nil {
		return nil
	}
	return &ast.ViewStmt{
		View:            convertRangeVar(n.View),
		Aliases:         convertSlice(n.Aliases),
		Query:           convertNode(n.Query),
		Replace:         n.Replace,
		Options:         convertSlice(n.Options),
		WithCheckOption: ast.ViewCheckOption(n.WithCheckOption),
	}
}

func convertWindowClause(n *pg_query.WindowClause) *ast.WindowClause {
	if n == nil {
		return nil
	}
	return &ast.WindowClause{
		Name:            makeString(n.Name),
		Refname:         makeString(n.Refname),
		PartitionClause: convertSlice(n.PartitionClause),
		OrderClause:     convertSlice(n.OrderClause),
		FrameOptions:    int(n.FrameOptions),
		StartOffset:     convertNode(n.StartOffset),
		EndOffset:       convertNode(n.EndOffset),
		Winref:          ast.Index(n.Winref),
		CopiedOrder:     n.CopiedOrder,
	}
}

func convertWindowDef(n *pg_query.WindowDef) *ast.WindowDef {
	if n == nil {
		return nil
	}
	return &ast.WindowDef{
		Name:            makeString(n.Name),
		Refname:         makeString(n.Refname),
		PartitionClause: convertSlice(n.PartitionClause),
		OrderClause:     convertSlice(n.OrderClause),
		FrameOptions:    int(n.FrameOptions),
		StartOffset:     convertNode(n.StartOffset),
		EndOffset:       convertNode(n.EndOffset),
		Location:        int(n.Location),
	}
}

func convertWindowFunc(n *pg_query.WindowFunc) *ast.WindowFunc {
	if n == nil {
		return nil
	}
	return &ast.WindowFunc{
		Xpr:         convertNode(n.Xpr),
		Winfnoid:    ast.Oid(n.Winfnoid),
		Wintype:     ast.Oid(n.Wintype),
		Wincollid:   ast.Oid(n.Wincollid),
		Inputcollid: ast.Oid(n.Inputcollid),
		Args:        convertSlice(n.Args),
		Aggfilter:   convertNode(n.Aggfilter),
		Winref:      ast.Index(n.Winref),
		Winstar:     n.Winstar,
		Winagg:      n.Winagg,
		Location:    int(n.Location),
	}
}

func convertWithCheckOption(n *pg_query.WithCheckOption) *ast.WithCheckOption {
	if n == nil {
		return nil
	}
	return &ast.WithCheckOption{
		Kind:     ast.WCOKind(n.Kind),
		Relname:  makeString(n.Relname),
		Polname:  makeString(n.Polname),
		Qual:     convertNode(n.Qual),
		Cascaded: n.Cascaded,
	}
}

func convertWithClause(n *pg_query.WithClause) *ast.WithClause {
	if n == nil {
		return nil
	}
	return &ast.WithClause{
		Ctes:      convertSlice(n.Ctes),
		Recursive: n.Recursive,
		Location:  int(n.Location),
	}
}

func convertXmlExpr(n *pg_query.XmlExpr) *ast.XmlExpr {
	if n == nil {
		return nil
	}
	return &ast.XmlExpr{
		Xpr:       convertNode(n.Xpr),
		Op:        ast.XmlExprOp(n.Op),
		Name:      makeString(n.Name),
		NamedArgs: convertSlice(n.NamedArgs),
		ArgNames:  convertSlice(n.ArgNames),
		Args:      convertSlice(n.Args),
		Xmloption: ast.XmlOptionType(n.Xmloption),
		Type:      ast.Oid(n.Type),
		Typmod:    n.Typmod,
		Location:  int(n.Location),
	}
}

func convertXmlSerialize(n *pg_query.XmlSerialize) *ast.XmlSerialize {
	if n == nil {
		return nil
	}
	return &ast.XmlSerialize{
		Xmloption: ast.XmlOptionType(n.Xmloption),
		Expr:      convertNode(n.Expr),
		TypeName:  convertTypeName(n.TypeName),
		Location:  int(n.Location),
	}
}

func convertNode(node *pg_query.Node) (Statement, error) {
	if node == nil || node.Node == nil {
		return Statement{}, fmt.Errorf("Invalid Node object")
	}

	switch n := node.Node.(type) {

	case *pg_query.Node_AArrayExpr:
		return convertA_ArrayExpr(n.AArrayExpr)

	case *pg_query.Node_AConst:
		return convertA_Const(n.AConst)

	case *pg_query.Node_AExpr:
		return convertA_Expr(n.AExpr)

	case *pg_query.Node_AIndices:
		return convertA_Indices(n.AIndices)

	case *pg_query.Node_AIndirection:
		return convertA_Indirection(n.AIndirection)

	case *pg_query.Node_AStar:
		return convertA_Star(n.AStar)

	case *pg_query.Node_AccessPriv:
		return convertAccessPriv(n.AccessPriv)

	case *pg_query.Node_Aggref:
		return convertAggref(n.Aggref)

	case *pg_query.Node_Alias:
		return convertAlias(n.Alias)

	case *pg_query.Node_AlterCollationStmt:
		return convertAlterCollationStmt(n.AlterCollationStmt)

	case *pg_query.Node_AlterDatabaseSetStmt:
		return convertAlterDatabaseSetStmt(n.AlterDatabaseSetStmt)

	case *pg_query.Node_AlterDatabaseStmt:
		return convertAlterDatabaseStmt(n.AlterDatabaseStmt)

	case *pg_query.Node_AlterDefaultPrivilegesStmt:
		return convertAlterDefaultPrivilegesStmt(n.AlterDefaultPrivilegesStmt)

	case *pg_query.Node_AlterDomainStmt:
		return convertAlterDomainStmt(n.AlterDomainStmt)

	case *pg_query.Node_AlterEnumStmt:
		return convertAlterEnumStmt(n.AlterEnumStmt)

	case *pg_query.Node_AlterEventTrigStmt:
		return convertAlterEventTrigStmt(n.AlterEventTrigStmt)

	case *pg_query.Node_AlterExtensionContentsStmt:
		return convertAlterExtensionContentsStmt(n.AlterExtensionContentsStmt)

	case *pg_query.Node_AlterExtensionStmt:
		return convertAlterExtensionStmt(n.AlterExtensionStmt)

	case *pg_query.Node_AlterFdwStmt:
		return convertAlterFdwStmt(n.AlterFdwStmt)

	case *pg_query.Node_AlterForeignServerStmt:
		return convertAlterForeignServerStmt(n.AlterForeignServerStmt)

	case *pg_query.Node_AlterFunctionStmt:
		return convertAlterFunctionStmt(n.AlterFunctionStmt)

	case *pg_query.Node_AlterObjectDependsStmt:
		return convertAlterObjectDependsStmt(n.AlterObjectDependsStmt)

	case *pg_query.Node_AlterObjectSchemaStmt:
		return convertAlterObjectSchemaStmt(n.AlterObjectSchemaStmt)

	case *pg_query.Node_AlterOpFamilyStmt:
		return convertAlterOpFamilyStmt(n.AlterOpFamilyStmt)

	case *pg_query.Node_AlterOperatorStmt:
		return convertAlterOperatorStmt(n.AlterOperatorStmt)

	case *pg_query.Node_AlterOwnerStmt:
		return convertAlterOwnerStmt(n.AlterOwnerStmt)

	case *pg_query.Node_AlterPolicyStmt:
		return convertAlterPolicyStmt(n.AlterPolicyStmt)

	case *pg_query.Node_AlterPublicationStmt:
		return convertAlterPublicationStmt(n.AlterPublicationStmt)

	case *pg_query.Node_AlterRoleSetStmt:
		return convertAlterRoleSetStmt(n.AlterRoleSetStmt)

	case *pg_query.Node_AlterRoleStmt:
		return convertAlterRoleStmt(n.AlterRoleStmt)

	case *pg_query.Node_AlterSeqStmt:
		return convertAlterSeqStmt(n.AlterSeqStmt)

	case *pg_query.Node_AlterSubscriptionStmt:
		return convertAlterSubscriptionStmt(n.AlterSubscriptionStmt)

	case *pg_query.Node_AlterSystemStmt:
		return convertAlterSystemStmt(n.AlterSystemStmt)

	case *pg_query.Node_AlterTsconfigurationStmt:
		return convertAlterTSConfigurationStmt(n.AlterTsconfigurationStmt)

	case *pg_query.Node_AlterTsdictionaryStmt:
		return convertAlterTSDictionaryStmt(n.AlterTsdictionaryStmt)

	case *pg_query.Node_AlterTableCmd:
		return convertAlterTableCmd(n.AlterTableCmd)

	case *pg_query.Node_AlterTableMoveAllStmt:
		return convertAlterTableMoveAllStmt(n.AlterTableMoveAllStmt)

	case *pg_query.Node_AlterTableSpaceOptionsStmt:
		return convertAlterTableSpaceOptionsStmt(n.AlterTableSpaceOptionsStmt)

	case *pg_query.Node_AlterTableStmt:
		return convertAlterTableStmt(n.AlterTableStmt)

	case *pg_query.Node_AlterUserMappingStmt:
		return convertAlterUserMappingStmt(n.AlterUserMappingStmt)

	case *pg_query.Node_AlternativeSubPlan:
		return convertAlternativeSubPlan(n.AlternativeSubPlan)

	case *pg_query.Node_ArrayCoerceExpr:
		return convertArrayCoerceExpr(n.ArrayCoerceExpr)

	case *pg_query.Node_ArrayExpr:
		return convertArrayExpr(n.ArrayExpr)

	case *pg_query.Node_BitString:
		return convertBitString(n.BitString)

	case *pg_query.Node_BoolExpr:
		return convertBoolExpr(n.BoolExpr)

	case *pg_query.Node_Boolean:
		return convertBoolean(n.Boolean)

	case *pg_query.Node_BooleanTest:
		return convertBooleanTest(n.BooleanTest)

	case *pg_query.Node_CallStmt:
		return convertCallStmt(n.CallStmt)

	case *pg_query.Node_CaseExpr:
		return convertCaseExpr(n.CaseExpr)

	case *pg_query.Node_CaseTestExpr:
		return convertCaseTestExpr(n.CaseTestExpr)

	case *pg_query.Node_CaseWhen:
		return convertCaseWhen(n.CaseWhen)

	case *pg_query.Node_CheckPointStmt:
		return convertCheckPointStmt(n.CheckPointStmt)

	case *pg_query.Node_ClosePortalStmt:
		return convertClosePortalStmt(n.ClosePortalStmt)

	case *pg_query.Node_ClusterStmt:
		return convertClusterStmt(n.ClusterStmt)

	case *pg_query.Node_CoalesceExpr:
		return convertCoalesceExpr(n.CoalesceExpr)

	case *pg_query.Node_CoerceToDomain:
		return convertCoerceToDomain(n.CoerceToDomain)

	case *pg_query.Node_CoerceToDomainValue:
		return convertCoerceToDomainValue(n.CoerceToDomainValue)

	case *pg_query.Node_CoerceViaIo:
		return convertCoerceViaIO(n.CoerceViaIo)

	case *pg_query.Node_CollateClause:
		return convertCollateClause(n.CollateClause)

	case *pg_query.Node_CollateExpr:
		return convertCollateExpr(n.CollateExpr)

	case *pg_query.Node_ColumnDef:
		return convertColumnDef(n.ColumnDef)

	case *pg_query.Node_ColumnRef:
		return convertColumnRef(n.ColumnRef)

	case *pg_query.Node_CommentStmt:
		return convertCommentStmt(n.CommentStmt)

	case *pg_query.Node_CommonTableExpr:
		return convertCommonTableExpr(n.CommonTableExpr)

	case *pg_query.Node_CompositeTypeStmt:
		return convertCompositeTypeStmt(n.CompositeTypeStmt)

	case *pg_query.Node_Constraint:
		return convertConstraint(n.Constraint)

	case *pg_query.Node_ConstraintsSetStmt:
		return convertConstraintsSetStmt(n.ConstraintsSetStmt)

	case *pg_query.Node_ConvertRowtypeExpr:
		return convertConvertRowtypeExpr(n.ConvertRowtypeExpr)

	case *pg_query.Node_CopyStmt:
		return convertCopyStmt(n.CopyStmt)

	case *pg_query.Node_CreateAmStmt:
		return convertCreateAmStmt(n.CreateAmStmt)

	case *pg_query.Node_CreateCastStmt:
		return convertCreateCastStmt(n.CreateCastStmt)

	case *pg_query.Node_CreateConversionStmt:
		return convertCreateConversionStmt(n.CreateConversionStmt)

	case *pg_query.Node_CreateDomainStmt:
		return convertCreateDomainStmt(n.CreateDomainStmt)

	case *pg_query.Node_CreateEnumStmt:
		return convertCreateEnumStmt(n.CreateEnumStmt)

	case *pg_query.Node_CreateEventTrigStmt:
		return convertCreateEventTrigStmt(n.CreateEventTrigStmt)

	case *pg_query.Node_CreateExtensionStmt:
		return convertCreateExtensionStmt(n.CreateExtensionStmt)

	case *pg_query.Node_CreateFdwStmt:
		return convertCreateFdwStmt(n.CreateFdwStmt)

	case *pg_query.Node_CreateForeignServerStmt:
		return convertCreateForeignServerStmt(n.CreateForeignServerStmt)

	case *pg_query.Node_CreateForeignTableStmt:
		return convertCreateForeignTableStmt(n.CreateForeignTableStmt)

	case *pg_query.Node_CreateFunctionStmt:
		return convertCreateFunctionStmt(n.CreateFunctionStmt)

	case *pg_query.Node_CreateOpClassItem:
		return convertCreateOpClassItem(n.CreateOpClassItem)

	case *pg_query.Node_CreateOpClassStmt:
		return convertCreateOpClassStmt(n.CreateOpClassStmt)

	case *pg_query.Node_CreateOpFamilyStmt:
		return convertCreateOpFamilyStmt(n.CreateOpFamilyStmt)

	case *pg_query.Node_CreatePlangStmt:
		return convertCreatePLangStmt(n.CreatePlangStmt)

	case *pg_query.Node_CreatePolicyStmt:
		return convertCreatePolicyStmt(n.CreatePolicyStmt)

	case *pg_query.Node_CreatePublicationStmt:
		return convertCreatePublicationStmt(n.CreatePublicationStmt)

	case *pg_query.Node_CreateRangeStmt:
		return convertCreateRangeStmt(n.CreateRangeStmt)

	case *pg_query.Node_CreateRoleStmt:
		return convertCreateRoleStmt(n.CreateRoleStmt)

	case *pg_query.Node_CreateSchemaStmt:
		return convertCreateSchemaStmt(n.CreateSchemaStmt)

	case *pg_query.Node_CreateSeqStmt:
		return convertCreateSeqStmt(n.CreateSeqStmt)

	case *pg_query.Node_CreateStatsStmt:
		return convertCreateStatsStmt(n.CreateStatsStmt)

	case *pg_query.Node_CreateStmt:
		return convertCreateStmt(n.CreateStmt)

	case *pg_query.Node_CreateSubscriptionStmt:
		return convertCreateSubscriptionStmt(n.CreateSubscriptionStmt)

	case *pg_query.Node_CreateTableAsStmt:
		return convertCreateTableAsStmt(n.CreateTableAsStmt)

	case *pg_query.Node_CreateTableSpaceStmt:
		return convertCreateTableSpaceStmt(n.CreateTableSpaceStmt)

	case *pg_query.Node_CreateTransformStmt:
		return convertCreateTransformStmt(n.CreateTransformStmt)

	case *pg_query.Node_CreateTrigStmt:
		return convertCreateTrigStmt(n.CreateTrigStmt)

	case *pg_query.Node_CreateUserMappingStmt:
		return convertCreateUserMappingStmt(n.CreateUserMappingStmt)

	case *pg_query.Node_CreatedbStmt:
		return convertCreatedbStmt(n.CreatedbStmt)

	case *pg_query.Node_CurrentOfExpr:
		return convertCurrentOfExpr(n.CurrentOfExpr)

	case *pg_query.Node_DeallocateStmt:
		return convertDeallocateStmt(n.DeallocateStmt)

	case *pg_query.Node_DeclareCursorStmt:
		return convertDeclareCursorStmt(n.DeclareCursorStmt)

	case *pg_query.Node_DefElem:
		return convertDefElem(n.DefElem)

	case *pg_query.Node_DefineStmt:
		return convertDefineStmt(n.DefineStmt)

	case *pg_query.Node_DeleteStmt:
		return convertDeleteStmt(n.DeleteStmt)

	case *pg_query.Node_DiscardStmt:
		return convertDiscardStmt(n.DiscardStmt)

	case *pg_query.Node_DoStmt:
		return convertDoStmt(n.DoStmt)

	case *pg_query.Node_DropOwnedStmt:
		return convertDropOwnedStmt(n.DropOwnedStmt)

	case *pg_query.Node_DropRoleStmt:
		return convertDropRoleStmt(n.DropRoleStmt)

	case *pg_query.Node_DropStmt:
		return convertDropStmt(n.DropStmt)

	case *pg_query.Node_DropSubscriptionStmt:
		return convertDropSubscriptionStmt(n.DropSubscriptionStmt)

	case *pg_query.Node_DropTableSpaceStmt:
		return convertDropTableSpaceStmt(n.DropTableSpaceStmt)

	case *pg_query.Node_DropUserMappingStmt:
		return convertDropUserMappingStmt(n.DropUserMappingStmt)

	case *pg_query.Node_DropdbStmt:
		return convertDropdbStmt(n.DropdbStmt)

	case *pg_query.Node_ExecuteStmt:
		return convertExecuteStmt(n.ExecuteStmt)

	case *pg_query.Node_ExplainStmt:
		return convertExplainStmt(n.ExplainStmt)

	case *pg_query.Node_FetchStmt:
		return convertFetchStmt(n.FetchStmt)

	case *pg_query.Node_FieldSelect:
		return convertFieldSelect(n.FieldSelect)

	case *pg_query.Node_FieldStore:
		return convertFieldStore(n.FieldStore)

	case *pg_query.Node_Float:
		return convertFloat(n.Float)

	case *pg_query.Node_FromExpr:
		return convertFromExpr(n.FromExpr)

	case *pg_query.Node_FuncCall:
		return convertFuncCall(n.FuncCall)

	case *pg_query.Node_FuncExpr:
		return convertFuncExpr(n.FuncExpr)

	case *pg_query.Node_FunctionParameter:
		return convertFunctionParameter(n.FunctionParameter)

	case *pg_query.Node_GrantRoleStmt:
		return convertGrantRoleStmt(n.GrantRoleStmt)

	case *pg_query.Node_GrantStmt:
		return convertGrantStmt(n.GrantStmt)

	case *pg_query.Node_GroupingFunc:
		return convertGroupingFunc(n.GroupingFunc)

	case *pg_query.Node_GroupingSet:
		return convertGroupingSet(n.GroupingSet)

	case *pg_query.Node_ImportForeignSchemaStmt:
		return convertImportForeignSchemaStmt(n.ImportForeignSchemaStmt)

	case *pg_query.Node_IndexElem:
		return convertIndexElem(n.IndexElem)

	case *pg_query.Node_IndexStmt:
		return convertIndexStmt(n.IndexStmt)

	case *pg_query.Node_InferClause:
		return convertInferClause(n.InferClause)

	case *pg_query.Node_InferenceElem:
		return convertInferenceElem(n.InferenceElem)

	case *pg_query.Node_InlineCodeBlock:
		return convertInlineCodeBlock(n.InlineCodeBlock)

	case *pg_query.Node_InsertStmt:
		return convertInsertStmt(n.InsertStmt)

	case *pg_query.Node_Integer:
		return convertInteger(n.Integer)

	case *pg_query.Node_IntoClause:
		return convertIntoClause(n.IntoClause)

	case *pg_query.Node_JoinExpr:
		return convertJoinExpr(n.JoinExpr)

	case *pg_query.Node_List:
		return convertList(n.List)

	case *pg_query.Node_ListenStmt:
		return convertListenStmt(n.ListenStmt)

	case *pg_query.Node_LoadStmt:
		return convertLoadStmt(n.LoadStmt)

	case *pg_query.Node_LockStmt:
		return convertLockStmt(n.LockStmt)

	case *pg_query.Node_LockingClause:
		return convertLockingClause(n.LockingClause)

	case *pg_query.Node_MinMaxExpr:
		return convertMinMaxExpr(n.MinMaxExpr)

	case *pg_query.Node_MultiAssignRef:
		return convertMultiAssignRef(n.MultiAssignRef)

	case *pg_query.Node_NamedArgExpr:
		return convertNamedArgExpr(n.NamedArgExpr)

	case *pg_query.Node_NextValueExpr:
		return convertNextValueExpr(n.NextValueExpr)

	case *pg_query.Node_NotifyStmt:
		return convertNotifyStmt(n.NotifyStmt)

	case *pg_query.Node_NullTest:
		return convertNullTest(n.NullTest)

	case *pg_query.Node_ObjectWithArgs:
		return convertObjectWithArgs(n.ObjectWithArgs)

	case *pg_query.Node_OnConflictClause:
		return convertOnConflictClause(n.OnConflictClause)

	case *pg_query.Node_OnConflictExpr:
		return convertOnConflictExpr(n.OnConflictExpr)

	case *pg_query.Node_OpExpr:
		return convertOpExpr(n.OpExpr)

	case *pg_query.Node_Param:
		return convertParam(n.Param)

	case *pg_query.Node_ParamRef:
		return convertParamRef(n.ParamRef)

	case *pg_query.Node_PartitionBoundSpec:
		return convertPartitionBoundSpec(n.PartitionBoundSpec)

	case *pg_query.Node_PartitionCmd:
		return convertPartitionCmd(n.PartitionCmd)

	case *pg_query.Node_PartitionElem:
		return convertPartitionElem(n.PartitionElem)

	case *pg_query.Node_PartitionRangeDatum:
		return convertPartitionRangeDatum(n.PartitionRangeDatum)

	case *pg_query.Node_PartitionSpec:
		return convertPartitionSpec(n.PartitionSpec)

	case *pg_query.Node_PrepareStmt:
		return convertPrepareStmt(n.PrepareStmt)

	case *pg_query.Node_Query:
		return convertQuery(n.Query)

	case *pg_query.Node_RangeFunction:
		return convertRangeFunction(n.RangeFunction)

	case *pg_query.Node_RangeSubselect:
		return convertRangeSubselect(n.RangeSubselect)

	case *pg_query.Node_RangeTableFunc:
		return convertRangeTableFunc(n.RangeTableFunc)

	case *pg_query.Node_RangeTableFuncCol:
		return convertRangeTableFuncCol(n.RangeTableFuncCol)

	case *pg_query.Node_RangeTableSample:
		return convertRangeTableSample(n.RangeTableSample)

	case *pg_query.Node_RangeTblEntry:
		return convertRangeTblEntry(n.RangeTblEntry)

	case *pg_query.Node_RangeTblFunction:
		return convertRangeTblFunction(n.RangeTblFunction)

	case *pg_query.Node_RangeTblRef:
		return convertRangeTblRef(n.RangeTblRef)

	case *pg_query.Node_RangeVar:
		return convertRangeVar(n.RangeVar)

	case *pg_query.Node_RawStmt:
		return convertRawStmt(n.RawStmt)

	case *pg_query.Node_ReassignOwnedStmt:
		return convertReassignOwnedStmt(n.ReassignOwnedStmt)

	case *pg_query.Node_RefreshMatViewStmt:
		return convertRefreshMatViewStmt(n.RefreshMatViewStmt)

	case *pg_query.Node_ReindexStmt:
		return convertReindexStmt(n.ReindexStmt)

	case *pg_query.Node_RelabelType:
		return convertRelabelType(n.RelabelType)

	case *pg_query.Node_RenameStmt:
		return convertRenameStmt(n.RenameStmt)

	case *pg_query.Node_ReplicaIdentityStmt:
		return convertReplicaIdentityStmt(n.ReplicaIdentityStmt)

	case *pg_query.Node_ResTarget:
		return convertResTarget(n.ResTarget)

	case *pg_query.Node_RoleSpec:
		return convertRoleSpec(n.RoleSpec)

	case *pg_query.Node_RowCompareExpr:
		return convertRowCompareExpr(n.RowCompareExpr)

	case *pg_query.Node_RowExpr:
		return convertRowExpr(n.RowExpr)

	case *pg_query.Node_RowMarkClause:
		return convertRowMarkClause(n.RowMarkClause)

	case *pg_query.Node_RuleStmt:
		return convertRuleStmt(n.RuleStmt)

	case *pg_query.Node_SqlvalueFunction:
		return convertSQLValueFunction(n.SqlvalueFunction)

	case *pg_query.Node_ScalarArrayOpExpr:
		return convertScalarArrayOpExpr(n.ScalarArrayOpExpr)

	case *pg_query.Node_SecLabelStmt:
		return convertSecLabelStmt(n.SecLabelStmt)

	case *pg_query.Node_SelectStmt:
		return convertSelectStmt(n.SelectStmt)

	case *pg_query.Node_SetOperationStmt:
		return convertSetOperationStmt(n.SetOperationStmt)

	case *pg_query.Node_SetToDefault:
		return convertSetToDefault(n.SetToDefault)

	case *pg_query.Node_SortBy:
		return convertSortBy(n.SortBy)

	case *pg_query.Node_SortGroupClause:
		return convertSortGroupClause(n.SortGroupClause)

	case *pg_query.Node_String_:
		return convertString(n.String_)

	case *pg_query.Node_SubLink:
		return convertSubLink(n.SubLink)

	case *pg_query.Node_SubPlan:
		return convertSubPlan(n.SubPlan)

	case *pg_query.Node_TableFunc:
		return convertTableFunc(n.TableFunc)

	case *pg_query.Node_TableLikeClause:
		return convertTableLikeClause(n.TableLikeClause)

	case *pg_query.Node_TableSampleClause:
		return convertTableSampleClause(n.TableSampleClause)

	case *pg_query.Node_TargetEntry:
		return convertTargetEntry(n.TargetEntry)

	case *pg_query.Node_TransactionStmt:
		return convertTransactionStmt(n.TransactionStmt)

	case *pg_query.Node_TriggerTransition:
		return convertTriggerTransition(n.TriggerTransition)

	case *pg_query.Node_TruncateStmt:
		return convertTruncateStmt(n.TruncateStmt)

	case *pg_query.Node_TypeCast:
		return convertTypeCast(n.TypeCast)

	case *pg_query.Node_TypeName:
		return convertTypeName(n.TypeName)

	case *pg_query.Node_UnlistenStmt:
		return convertUnlistenStmt(n.UnlistenStmt)

	case *pg_query.Node_UpdateStmt:
		return convertUpdateStmt(n.UpdateStmt)

	case *pg_query.Node_VacuumStmt:
		return convertVacuumStmt(n.VacuumStmt)

	case *pg_query.Node_Var:
		return convertVar(n.Var)

	case *pg_query.Node_VariableSetStmt:
		return convertVariableSetStmt(n.VariableSetStmt)

	case *pg_query.Node_VariableShowStmt:
		return convertVariableShowStmt(n.VariableShowStmt)

	case *pg_query.Node_ViewStmt:
		return convertViewStmt(n.ViewStmt)

	case *pg_query.Node_WindowClause:
		return convertWindowClause(n.WindowClause)

	case *pg_query.Node_WindowDef:
		return convertWindowDef(n.WindowDef)

	case *pg_query.Node_WindowFunc:
		return convertWindowFunc(n.WindowFunc)

	case *pg_query.Node_WithCheckOption:
		return convertWithCheckOption(n.WithCheckOption)

	case *pg_query.Node_WithClause:
		return convertWithClause(n.WithClause)

	case *pg_query.Node_XmlExpr:
		return convertXmlExpr(n.XmlExpr)

	case *pg_query.Node_XmlSerialize:
		return convertXmlSerialize(n.XmlSerialize)

	default:
		return &ast.TODO{}
	}
}
