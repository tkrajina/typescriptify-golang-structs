package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

type arrayImports []string

func (i *arrayImports) String() string {
	return "// custom imports:\n\n" + strings.Join(*i, "\n")
}

func (i *arrayImports) Set(value string) error {
	*i = append(*i, value)
	return nil
}

const TEMPLATE = `package main

import (
	"fmt"

	m "{{ .ModelsPackage }}"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	t := typescriptify.New()
	t.CreateInterface = {{ .Interface }}
{{ range $key, $value := .InitParams }}	t.{{ $key }}={{ $value }}
{{ end }}
{{ range .Structs }}	t.Add({{ . }}{})
{{ end }}
{{ range .CustomImports }}	t.AddImport("{{ . }}")
{{ end }}
	err := t.ConvertToFile("{{ .TargetFile }}")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("OK")
}`

type Params struct {
	ModelsPackage string
	TargetFile    string
	Structs       []string
	InitParams    map[string]interface{}
	CustomImports arrayImports
	Interface     bool
	Verbose       bool
}

func main() {
	var p Params
	var backupDir string
	flag.StringVar(&p.ModelsPackage, "package", "", "Path of the package with models")
	flag.StringVar(&p.TargetFile, "target", "", "Target typescript file")
	flag.StringVar(&backupDir, "backup", "", "Directory where backup files are saved")
	flag.BoolVar(&p.Interface, "interface", false, "Create interfaces (not classes)")
	flag.Var(&p.CustomImports, "import", "Typescript import for your custom type, repeat this option for each import needed")
	flag.BoolVar(&p.Verbose, "verbose", false, "Verbose logs")
	flag.Parse()

	structs := []string{}
	for _, structOrGoFile := range flag.Args() {
		if strings.HasSuffix(structOrGoFile, ".go") {
			fmt.Println("Parsing:", structOrGoFile)
			fileStructs, err := GetGolangFileStructs(structOrGoFile)
			if err != nil {
				panic(fmt.Sprintf("Error loading/parsing golang file %s: %s", structOrGoFile, err.Error()))
			}
			structs = append(structs, fileStructs...)
		} else {
			structs = append(structs, structOrGoFile)
		}
	}

	if len(p.ModelsPackage) == 0 {
		fmt.Fprintln(os.Stderr, "No package given")
		os.Exit(1)
	}
	if len(p.TargetFile) == 0 {
		fmt.Fprintln(os.Stderr, "No target file")
		os.Exit(1)
	}

	t := template.Must(template.New("").Parse(TEMPLATE))

	f, err := os.CreateTemp(os.TempDir(), "typescriptify_*.go")
	handleErr(err)
	defer f.Close()

	structsArr := make([]string, 0)
	for _, str := range structs {
		str = strings.TrimSpace(str)
		if strings.Contains(str, string(filepath.Separator)) {
			continue
		}
		if len(str) > 0 {
			structsArr = append(structsArr, "m."+str)
		}
	}

	p.Structs = structsArr
	p.InitParams = map[string]interface{}{
		"BackupDir": fmt.Sprintf(`"%s"`, backupDir),
	}
	err = t.Execute(f, p)
	handleErr(err)

	if p.Verbose {
		byts, err := os.ReadFile(f.Name())
		handleErr(err)
		fmt.Printf("\nCompiling generated code (%s):\n%s\n----------------------------------------------------------------------------------------------------\n", f.Name(), string(byts))
	}

	cmd := exec.Command("go", "run", f.Name())
	fmt.Println(strings.Join(cmd.Args, " "))
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		handleErr(err)
	}
	fmt.Println(string(output))
}

func GetGolangFileStructs(filename string) ([]string, error) {
	fset := token.NewFileSet() // positions are relative to fset

	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}

	v := &AVisitor{}
	ast.Walk(v, f)

	return v.structs, nil
}

type AVisitor struct {
	structNameCandidate string
	structs             []string
}

func (v *AVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		switch t := node.(type) {
		case *ast.Ident:
			v.structNameCandidate = t.Name
		case *ast.StructType:
			if len(v.structNameCandidate) > 0 {
				v.structs = append(v.structs, v.structNameCandidate)
				v.structNameCandidate = ""
			}
		default:
			v.structNameCandidate = ""
		}
	}
	return v
}

func handleErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
