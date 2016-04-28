package jsonconv

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"strings"
)

func init() {
	_ = fmt.Sprintf
	_ = errors.New
	_ = os.Stderr
	_ = html.EscapeString
}

// Generated code, do not edit!!!!

func TE__java(args TemplateArgs) (string, error) {
	__template__ := "java.tmpl"
	_ = __template__
	__escape__ := html.EscapeString
	_ = __escape__
	var result bytes.Buffer
	/*  */
	result.WriteString(`
`)
	/* !for _, entity := range args.Entities { */
	for _, entity := range args.Entities {

		/* public class {{s entity.Name }} { */
		result.WriteString(fmt.Sprintf(`public class %s {
`, __escape__(entity.Name)))
		/* !		for _, field := range entity.Fields { */
		for _, field := range entity.Fields {

			/* @JsonProperty("{{=s field.JsonName }}") */
			result.WriteString(fmt.Sprintf(`    @JsonProperty("%s")
`, field.JsonName))
			/* private {{=s args.JSONFieldTypeString(field) }} {{=s field.JsonName }}; */
			result.WriteString(fmt.Sprintf(`    private %s %s;
`, args.JSONFieldTypeString(field), field.JsonName))
			/*  */
			result.WriteString(`
`)
			/* !		} */
		}

		/*  */
		result.WriteString(`
`)
		/* !		for _, field := range entity.Fields { */
		for _, field := range entity.Fields {

			/* public set{{=s strings.Title(field.JsonName) }}(value {{=s args.JSONFieldTypeString(field) }}) { */
			result.WriteString(fmt.Sprintf(`    public set%s(value %s) {
`, strings.Title(field.JsonName), args.JSONFieldTypeString(field)))
			/* this.{{=s field.JsonName }} = value; */
			result.WriteString(fmt.Sprintf(`            this.%s = value;
`, field.JsonName))
			/* } */
			result.WriteString(`    }
`)
			/* public get{{=s strings.Title(field.JsonName) }}() { */
			result.WriteString(fmt.Sprintf(`    public get%s() {
`, strings.Title(field.JsonName)))
			/* return this.{{=s field.JsonName }}; */
			result.WriteString(fmt.Sprintf(`            return this.%s;
`, field.JsonName))
			/* } */
			result.WriteString(`    }
`)
			/*  */
			result.WriteString(`
`)
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

func T__java(args TemplateArgs) string {
	html, err := TE__java(args)
	if err != nil {
		os.Stderr.WriteString("Error running template java.tmpl:" + err.Error())
	}
	return html
}
