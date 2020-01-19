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
type cobracmd func(c *cobra.Command, args []string)

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

var (
	glob   string
	stream bool
	// TODO currently a global func map must be declared for all templates.
	funcMap = template.FuncMap{
		"Title": strings.Title,
	}
)

func init() {
	glob = os.Getenv("TPL_GLOB")
	if glob == "" {
		log.Fatal("TPL_GLOB must not be empty")
	}

	if os.Getenv("STREAM") == "1" {
		stream = true
	}

	var err error
	glob, err = filepath.Abs(glob)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Short: "Template lister and renderer.",
	}

	for _, t := range mustGlobTemplate(glob).Templates() {
		tvars := parseTemplateVars(t)
		if len(tvars) == 0 {
			continue
		}

		cmd := &cobra.Command{
			Use: t.Name(),
			Run: rendererCommand(t),
		}

		if !stream {
			// Declare template variables as command flags.
			// This mode is used to simply render a template once
			// by passing template variables content as arguments.
			for varName, tp := range tvars {
				cmd.Flags().String(varName, "", tp.String())
				cmd.MarkFlagRequired(varName)
			}
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
	varMap := make(map[string]vartype)

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

func rendererCommand(t *template.Template) cobracmd {
	cmd := func(c *cobra.Command, args []string) {
		data := make(map[string]interface{})
		if stream == false {
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
		} else {
			// In stream mode we read data from stdin and render to stdout.
			if err := streamTemplate(os.Stdout, t, os.Stdin); err != nil {
				log.Fatalf("Error streaming template: %s", err)
			}
		}
	}
	return cmd
}

// TODO streaming templates
func streamTemplate(w io.Writer, t *template.Template, jsonReader io.Reader) error {
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
		if err := t.Execute(w, data); err != nil {
			return err
		}
	}

	return nil
}
