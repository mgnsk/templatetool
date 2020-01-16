package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type vartype int

// Template variables types.
const (
	TypeString vartype = iota
	TypeJSON
)

func (tp vartype) String() string {
	switch tp {
	case TypeString:
		return "String"
	case TypeJSON:
		return "JSON"
	}
	panic("invalid type")
}

// TODO currently a global func map must be declared for all templates.
var funcMap = template.FuncMap{
	"Title": strings.Title,
}

func main() {
	rootCmd := &cobra.Command{
		Short: "Template lister and renderer.",
	}

	glob := os.Getenv("TPL_GLOB")
	if glob == "" {
		log.Fatal("TPL_GLOB must not be empty")
	}

	absGlob, err := filepath.Abs(glob)
	if err != nil {
		log.Fatal(err)
	}

	for _, t := range mustGlobTemplate(absGlob).Templates() {
		tvars := parseTemplateVars(t)
		if len(tvars) == 0 {
			continue
		}

		cmd := &cobra.Command{
			Use: t.Name(),
			Run: renderer(t),
		}

		// Declare template variables as command flags.
		for varName, tp := range tvars {
			cmd.Flags().String(varName, "", tp.String())
			cmd.MarkFlagRequired(varName)
		}

		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func mustGlobTemplate(glob string) *template.Template {
	tpls, err := template.New("").Funcs(funcMap).ParseGlob(glob)
	if err != nil {
		log.Fatalf("Error parsing templates from glob: %s", glob)
	}
	return tpls
}

func parseCommandNodes(result map[string]vartype, tp vartype, cmds ...*parse.CommandNode) {
	for _, c := range cmds {
		for _, a := range c.Args {
			if f, ok := a.(*parse.FieldNode); ok {
				for _, ident := range f.Ident {
					if _, exists := result[ident]; !exists {
						result[ident] = tp
					}
				}
			}
		}
	}
}

func parseTemplateVars(t *template.Template) map[string]vartype {
	varMap := map[string]vartype{}

	for _, node := range t.Tree.Root.Nodes {
		switch n := node.(type) {
		case *parse.ActionNode:
			parseCommandNodes(varMap, TypeString, n.Pipe.Cmds...)
		case *parse.RangeNode:
			parseCommandNodes(varMap, TypeJSON, n.Pipe.Cmds...)
		}
	}

	return varMap
}

func renderer(t *template.Template) func(c *cobra.Command, args []string) {
	return func(c *cobra.Command, args []string) {
		data := make(map[string]interface{})
		c.Flags().VisitAll(func(f *pflag.Flag) {
			switch f.Usage {
			case "String":
				data[f.Name] = f.Value.String()
			case "JSON":
				m := make(map[string]interface{})
				if err := json.Unmarshal([]byte(f.Value.String()), &m); err != nil {
					log.Fatalf("Invalid JSON: %s", err)
				}
				data[f.Name] = m
			}

		})
		if err := t.Execute(os.Stdout, data); err != nil {
			log.Fatalf("Error executing template: %s", err)
		}
	}
}

// TODO streaming templates
func streamTemplate(t *template.Template, output io.Writer, jsonReader io.Reader) error {
	dec := json.NewDecoder(jsonReader)
	tk, err := dec.Token()
	if err != nil {
		return err
	}
	// according to the docs a json.Delim is one [, ], { or }
	// Make sure the token is a delim
	delim, ok := tk.(json.Delim)
	if !ok {
		panic("first token not a delim")
	}
	// Make sure the value of the delim is '['
	if delim != json.Delim('[') {
		panic("first token not a [")
	}

	for dec.More() {
		data := make(map[string]interface{})
		err = dec.Decode(&data)
		if err != nil {
			return err
		}
		if err := t.Execute(output, data); err != nil {
			return err
		}
	}

	return nil
}
