package parser

import "testing"

func TestName(t *testing.T) {
	ast := DjotAst([]byte(`hello *wo_hey_rld*\  
`))
	t.Log(ast)
}
