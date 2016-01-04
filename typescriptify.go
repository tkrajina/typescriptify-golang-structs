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

	"github.com/tkrajina/typescriptify-golang-structs/typescriptify"
)

func main() {
	t := typescriptify.New()
{{ range .Structs }}	t.Add({{ . }}{})
{{ end }}
	result, err := t.ConvertTo(os.Stdout)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(result))
}`

type Params struct {
	Structs []string
}

func main() {
	var stringExtension string
	var structs string
	flag.StringVar(&structs, "structs", "", "List of (comma-delimited) structs to be typescriptified")
	flag.StringVar(&stringExtension, "extension", "", "")
	flag.Parse()

	t := template.Must(template.New("").Parse(TEMPLATE))

	filename, err := ioutil.TempDir(os.TempDir(), "")
	handleErr(err)

	filename = fmt.Sprintf("%s/typescriptify_%d.go", filename, time.Now().Nanosecond())

	f, err := os.Create(filename)
	handleErr(err)
	defer f.Close()

	structsArr := make([]string, 0)
	for _, str := range strings.Split(structs, ",") {
		str = strings.TrimSpace(str)
		if len(str) > 0 {
			structsArr = append(structsArr, str)
		}
	}

	params := Params{Structs: structsArr}
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
