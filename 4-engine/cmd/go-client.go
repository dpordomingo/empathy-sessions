package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/src-d/docs-sources/enry/data"
	"gopkg.in/bblfsh/client-go.v3"
	"gopkg.in/src-d/enry.v1"

	"gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

func main() {
	client, err := bblfsh.NewClient(fmt.Sprintf("0.0.0.0:%s", os.Args[1]))
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}

	req := client.NewParseRequest().Content(string(content))

	// It is needed the Lang or Filename
	if len(os.Args) > 3 {
		if os.Args[3] != "unk" {
			req.Language(os.Args[3])
		} else {
			lang, _ := enry.GetLanguageByClassifier(content, enryLangs())
			req.Language(lang)
		}
	} else {
		req.Filename(os.Args[2])
	}

	res, lang, err := req.UAST()
	if err != nil {
		fmt.Printf("language: '%s'\n\n", lang)
		panic(err)
	}

	data, err := uastyml.Marshal(res)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n\n\nlanguage: '%s'\n\n", string(data), lang)
}

func enryLangs() []string {
	langs := make([]string, len(data.ExtensionsByLanguage))
	i := 0

	for lang := range data.ExtensionsByLanguage {
		langs[i] = lang
		i++
	}

	return langs
}
