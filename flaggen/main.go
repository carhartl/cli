package main

import (
	"encoding/json"
	"github.com/urfave/cli"
	"log"
	"os"
	"text/template"
)

type CliFlagInfo struct {
	PackageName string
	Flags       []FlagType
}

type FlagType struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Value          bool   `json:"value"`
	Destination    bool   `json:"dest"`
	Doctail        string `json:"doctail"`
	ContextDefault string `json:"context_default"`
	ContextType    string `json:"context_type"`
	Parser         string `json:"parser"`
	ParserCast     string `json:"parser_cast"`
}

var cliFlagTemplate = `// Code generated by fg; DO NOT EDIT.

package {{ .PackageName }}

import (
	"flag"
	"strconv"
	"time"
)
{{ range $i, $flag := .Flags }}
// {{ $flag.Name }}Flag is a flag with type {{ $flag.Type }}{{ $flag.Doctail }}
type {{ $flag.Name }}Flag struct {
    Name        string
    Usage       string
    EnvVar      string
    FilePath    string
    Required    bool
    Hidden      bool
    {{- if eq $flag.Value true }}
	Value		{{ $flag.Type }}
	{{- end }}
	{{- if eq $flag.Destination true }}
	Destination	*{{ $flag.Type }}
	{{- end }}
}

// String returns a readable representation of this value
// (for usage defaults)
func (f {{ $flag.Name }}Flag) String() string {
    return FlagStringer(f)
}

// GetName returns the name of the flag
func (f {{ $flag.Name }}Flag) GetName() string {
    return f.Name
}

// IsRequired returns whether or not the flag is required
func (f {{ $flag.Name }}Flag) IsRequired() bool {
    return f.Required
}

// {{ $flag.Name }} looks up the value of a local {{ $flag.Name }}Flag, returns
// {{ $flag.ContextDefault }} if not found
func (c *Context) {{ $flag.Name }}(name string) {{ if ne .ContextType "" }} {{ $flag.ContextType }} {{ else }} {{ $flag.Type }} {{- end }} {
    return lookup{{ $flag.Name }}(name, c.flagSet)
}

// Global{{ $flag.Name }} looks up the value of a global {{ $flag.Name }}Flag, returns
// {{ $flag.ContextDefault }} if not found
func (c *Context) Global{{ $flag.Name }}(name string) {{ if ne .ContextType "" }} {{ $flag.ContextType }} {{ else }} {{ $flag.Type }} {{- end }} {
    if fs := lookupGlobalFlagSet(name, c); fs != nil {
        return lookup{{ $flag.Name }}(name, fs)
    }
    return {{ $flag.ContextDefault }}
}

func lookup{{ $flag.Name }}(name string, set *flag.FlagSet) {{ if ne .ContextType "" }} {{ $flag.ContextType }} {{ else }} {{ $flag.Type }} {{- end }} {
    f := set.Lookup(name)
    if f != nil {
        {{ if ne .Parser "" }}parsed, err := {{ $flag.Parser }}{{ else }}parsed, err := f.Value, error(nil){{ end }}
        if err != nil {
            return {{ $flag.ContextDefault }}
        }
        {{ if ne .ParserCast "" }}return {{ $flag.ParserCast }}{{ else }}return parsed{{ end }}
    }
    return {{ $flag.ContextDefault }}
}
{{- end }}`

func main() {
	var packageName, inputPath, outputPath string

	app := cli.NewApp()

	app.Name = "fg"
	app.Usage = "Generate flag type code!"
	app.Version = "v0.1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "package, p",
			Value:       "cli",
			Usage:       "`PACKAGE` for which the flag types will be generated",
			Destination: &packageName,
		},
		cli.StringFlag{
			Name:        "input, i",
			Usage:       "path to the `INPUT JSON FILE` which defines each type to be generated",
			Destination: &inputPath,
		},
		cli.StringFlag{
			Name:        "output, o",
			Usage:       "path to the `OUTPUT GO FILE` which will contain the flag types",
			Destination: &outputPath,
		},
	}

	app.Action = func(c *cli.Context) error {
		var info CliFlagInfo
		info.PackageName = packageName

		inFile, err := os.Open(inputPath)
		if err != nil {
			log.Fatal(err)
		}

		defer inFile.Close()

		decoder := json.NewDecoder(inFile)

		err = decoder.Decode(&info.Flags)
		if err != nil {
			log.Fatal(err)
		}

		tpl := template.Must(template.New("").Parse(cliFlagTemplate))

		outFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatal(err)
		}

		defer outFile.Close()

		err = tpl.Execute(outFile, info)
		if err != nil {
			log.Fatal(err)
		}

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
