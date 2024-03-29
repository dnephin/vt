package tma

import (
	"fmt"
	"go/ast"
	"strings"
)

type msgResult struct {
	got        any
	want       any
	vtFuncName string
	callSource messageCallSource
}

func Got(got any) string {
	// TODO colorize when in supported terminal
	const vtFuncName = "tma.Got"

	result := msgResult{got: got, vtFuncName: vtFuncName}

	// TODO: check if got is an error type from this package and return early

	callSource, err := getCallSource()
	if err != nil {
		// TODO: include tips about how to prevent this
		return fmt.Sprintf("%v, but %v: %v", result.basicMsg(), vtFuncName, err)
	}
	if len(callSource.CallExpr.Args) != 1 {
		return result.msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}
	result.callSource = callSource

	switch v := got.(type) {
	case nil:
		// TODO: could be error comparison
	case string:
		// diff from cmp.Diff
		return handleSingleArgString(v, result, callSource)
	case error:
		// error comparison
		return handleSingleArgError(v, result, callSource)
	}
	// otherwise try for comparison to constant
	// TODO:

	return "TODO"
}

func GotWant(got any, want any) string {
	// TODO colorize when in supported terminal
	const vtFuncName = "tma.GotWant"

	result := msgResult{got: got}
	callSource, err := getCallSource()
	if err != nil {
		// TODO: include tips about how to prevent this
		return fmt.Sprintf("%v, but %v: %v", result.basicMsg(), vtFuncName, err)
	}
	if len(callSource.CallExpr.Args) != 2 {
		return result.msgUnexpectedAstNode(callSource.CallExpr, "wrong number of args")
	}

	// TODO: lookup comparison and switch on token

	// TODO:
	return "TODO"
}

func (r msgResult) basicMsg() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "got=%v", r.got)
	if r.want != nil {
		fmt.Fprintf(&buf, ", want=%v", r.want)
	}
	return buf.String()
}

func exprFromObjDecl(ident *ast.Ident) ast.Expr {
	switch v := ident.Obj.Decl.(type) {
	case *ast.AssignStmt:
		// TODO: handle multiple assignment
		return v.Rhs[0]

	case *ast.ValueSpec:
		// TODO: handle multiple declaration
		return v.Values[0]
	}
	// TODO: other cases?
	return nil
}

func (r msgResult) msgErrorFromExpr(err error, expr ast.Expr, want ast.Expr) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return r.msgUnexpectedAstNode(expr, "expected a function call for the variable declaration")
	}

	var buf strings.Builder

	// TODO: handle not an ident for the func
	// TODO: remove args if longer than x.
	n, _ := formatNode(call)

	fmt.Fprintf(&buf, "%v returned error: %v", n, err)
	if want != nil {
		n, _ = formatNode(want)
		fmt.Fprintf(&buf, ", wanted %v", n)
	}
	if len(r.callSource.CallComments) != 0 {
		buf.WriteString("\n")
		for _, item := range r.callSource.CallComments {
			buf.WriteString(item.Text())
		}
	}

	return buf.String()
}

func (r msgResult) msgUnexpectedAstNode(node ast.Node, reason string) string {
	// TODO: include request for a bug report
	n, _ := formatNode(node)
	return fmt.Sprintf("%v, %v: %v, got %T:\n%v",
		r.basicMsg(), r.vtFuncName, reason, node, n)
}
