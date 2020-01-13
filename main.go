package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TODO currently a global func map must be declared.
var funcMap = template.FuncMap{
	"Title": strings.Title,
}

func main() {
	rootCmd := &cobra.Command{
		Short: "Template glob lister and renderer.",
	}

	glob := os.Getenv("TPL_GLOB")
	if glob == "" {
		panic("TPL_GLOB must be set")
	}

	absGlob, err := filepath.Abs(glob)
	if err != nil {
		panic(err)
	}

	for _, t := range mustGlobTemplate(absGlob).Templates() {
		cmd := createTemplateCommand(t)
		rootCmd.AddCommand(cmd)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func mustGlobTemplate(glob string) *template.Template {
	tpls, err := template.New("").Funcs(funcMap).ParseGlob(glob)
	if err != nil {
		panic(fmt.Errorf("Error parsing templates from glob: %s", glob))
	}
	return tpls
}

func createTemplateCommand(t *template.Template) *cobra.Command {
	cmd := &cobra.Command{Use: t.Name()}
	// Declare template variables as command flags.
	for tvar := range getTemplateVars(t) {
		cmd.Flags().String(tvar, "", tvar)
		cmd.MarkFlagRequired(tvar)
	}

	// Render the template based on data provided in flags.
	cmd.Run = func(c *cobra.Command, args []string) {
		data := make(map[string]string)
		c.Flags().VisitAll(func(f *pflag.Flag) {
			data[f.Name] = f.Value.String()
		})
		if err := t.Execute(os.Stdout, data); err != nil {
			panic(err)
		}
	}

	return cmd
}

func getTemplateVars(t *template.Template) map[string]struct{} {
	varMap := make(map[string]struct{})
	for _, node := range t.Tree.Root.Nodes {
		if n, ok := node.(*parse.ActionNode); ok {
			for _, c := range n.Pipe.Cmds {
				for _, a := range c.Args {
					if f, ok := a.(*parse.FieldNode); ok {
						for _, ident := range f.Ident {
							if _, exists := varMap[ident]; !exists {
								varMap[ident] = struct{}{}
							}
						}
					}
				}
			}
		}
	}

	return varMap
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
