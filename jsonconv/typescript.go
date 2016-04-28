package jsonconv

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
)

func init() {
	_ = fmt.Sprintf
	_ = errors.New
	_ = os.Stderr
	_ = html.EscapeString
}

// Generated code, do not edit!!!!

func TE__typescript(args TemplateArgs) (string, error) {
	__template__ := "typescript.tmpl"
	_ = __template__
	__escape__ := html.EscapeString
	_ = __escape__
	var result bytes.Buffer
	/*  */
	result.WriteString(`
`)
	/* !for _, entity := range args.Entities { */
	for _, entity := range args.Entities {

		/* class {{s entity.Name }} { */
		result.WriteString(fmt.Sprintf(`class %s {
`, __escape__(entity.Name)))
		/* !		for _, field := range entity.Fields { */
		for _, field := range entity.Fields {

			/* {{=s field.JsonName }}: {{=s args.JSONFieldTypeString(field) }}; */
			result.WriteString(fmt.Sprintf(`    %s: %s;
`, field.JsonName, args.JSONFieldTypeString(field)))
			/* !		} */
		}

		/* } */
		result.WriteString(`}
`)
		/* !} */
	}

	/*  */
	result.WriteString(``)

	return result.String(), nil
}

func T__typescript(args TemplateArgs) string {
	html, err := TE__typescript(args)
	if err != nil {
		os.Stderr.WriteString("Error running template typescript.tmpl:" + err.Error())
	}
	return html
}
