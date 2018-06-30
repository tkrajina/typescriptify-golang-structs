package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"
	"time"
)

const TEMPLATE = `package main

import (
	"fmt"

	"{{ .ModelsPackage }}"
	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	t := typescriptify.New()
{{ range $key, $value := .InitParams }}	t.{{ $key }}={{ $value }}
{{ end }}
{{ range .Structs }}	t.Add({{ . }}{})
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
}

func main() {
	var packagePath, target, backupDir string
	flag.StringVar(&packagePath, "package", "", "Path of the package with models")
	flag.StringVar(&target, "target", "", "Target typescript file")
	flag.StringVar(&backupDir, "backup", "", "Directory where backup files are saved")
	flag.Parse()

	structs := []string{}
	for _, structOrGoFile := range flag.Args() {
		if strings.HasSuffix(structOrGoFile, ".go") {
			fmt.Println("Parsing:", structOrGoFile)
			fileStructs, err := GetGolangFileStructs(structOrGoFile)
			if err != nil {
				panic(fmt.Sprintf("Error loading/parsing golang file %s: %s", structOrGoFile, err.Error()))
			}
			for _, s := range fileStructs {
				structs = append(structs, s)
			}
		} else {
			structs = append(structs, structOrGoFile)
		}
	}

	if len(packagePath) == 0 {
		fmt.Fprintln(os.Stderr, "No package given")
		os.Exit(1)
	}
	if len(target) == 0 {
		fmt.Fprintln(os.Stderr, "No target file")
		os.Exit(1)
	}

	packageParts := strings.Split(packagePath, string(os.PathSeparator))
	pckg := packageParts[len(packageParts)-1]

	t := template.Must(template.New("").Parse(TEMPLATE))

	filename, err := ioutil.TempDir(os.TempDir(), "")
	handleErr(err)

	filename = fmt.Sprintf("%s/typescriptify_%d.go", filename, time.Now().Nanosecond())

	f, err := os.Create(filename)
	handleErr(err)
	defer f.Close()

	structsArr := make([]string, 0)
	for _, str := range structs {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			structsArr = append(structsArr, pckg+"."+str)
		}
	}

	p := Params{
		Structs:       structsArr,
		ModelsPackage: packagePath,
		TargetFile:    target,
		InitParams: map[string]interface{}{
			"BackupDir": fmt.Sprintf(`"%s"`, backupDir),
		},
	}
	err = t.Execute(f, p)
	handleErr(err)

	cmd := exec.Command("go", "run", filename)
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
