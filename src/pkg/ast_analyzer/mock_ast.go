// pkg/ast_analyzer/mock_ast.go

package ast_analyzer

type mockNode struct {
    typ      string
    children []Node
}

func (m *mockNode) Type() string     { return m.typ }
func (m *mockNode) Children() []Node { return m.children }

type mockFunction struct {
    mockNode
    name       string
    parameters []string
}

func (m *mockFunction) Name() string         { return m.name }
func (m *mockFunction) Parameters() []string { return m.parameters }

type mockClass struct {
    mockNode
    name    string
    methods []Function
}

func (m *mockClass) Name() string       { return m.name }
func (m *mockClass) Methods() []Function { return m.methods }

type mockVariable struct {
    mockNode
    name     string
    typeInfo string
}

func (m *mockVariable) Name() string     { return m.name }
func (m *mockVariable) TypeInfo() string { return m.typeInfo }

func mockParse(filePath string) (Node, error) {
    // Create a mock AST structure for testing
    root := &mockNode{typ: "File", children: []Node{}}

    function := &mockFunction{
        mockNode:   mockNode{typ: "Function", children: []Node{}},
        name:       "exampleFunction",
        parameters: []string{"param1", "param2"},
    }
    root.children = append(root.children, function)

    class := &mockClass{
        mockNode: mockNode{typ: "Class", children: []Node{}},
        name:     "ExampleClass",
        methods:  []Function{function},
    }
    root.children = append(root.children, class)

    variable := &mockVariable{
        mockNode: mockNode{typ: "Variable", children: []Node{}},
        name:     "exampleVariable",
        typeInfo: "int",
    }
    root.children = append(root.children, variable)

    return root, nil
}