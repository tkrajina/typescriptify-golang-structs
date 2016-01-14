package main

import (
	"flag"
	"fmt"
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
}

func main() {
	var packagePath, target, stringExtension string
	flag.StringVar(&packagePath, "package", "", "Path of the package with models")
	flag.StringVar(&target, "target", "", "Target typescript file")
	flag.StringVar(&stringExtension, "extension", "", "")
	flag.Parse()

	structs := flag.Args()

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

	params := Params{Structs: structsArr, ModelsPackage: packagePath, TargetFile: target}
	err = t.Execute(f, params)
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

func handleErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
