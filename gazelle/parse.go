// Copyright 2019 The Bazel Authors. All rights reserved.
// Modifications copyright (C) 2021 BenchSci Analytics Inc.
// Modifications copyright (C) 2018 Ecosia GmbH

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package js

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func ParseJS(data []byte) ([]string, error) {

	imports := make([]string, 0)

	for _, match := range jsImportPattern.FindAllSubmatch(data, -1) {
		switch {
		case match[IMPORT] != nil:
			unquoted, err := unquoteImportString(match[IMPORT])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[IMPORT], err)
			}
			imports = append(imports, unquoted)

		case match[REQUIRE] != nil:
			unquoted, err := unquoteImportString(match[REQUIRE])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[REQUIRE], err)
			}
			imports = append(imports, strings.ToLower(unquoted))

		case match[EXPORT] != nil:
			unquoted, err := unquoteImportString(match[EXPORT])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[EXPORT], err)
			}
			imports = append(imports, strings.ToLower(unquoted))

		case match[JEST_MOCK] != nil:
			unquoted, err := unquoteImportString(match[JEST_MOCK])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[JEST_MOCK], err)
			}
			imports = append(imports, strings.ToLower(unquoted))

		case match[JEST_REQUIRE_ACTUAL] != nil:
			unquoted, err := unquoteImportString(match[JEST_REQUIRE_ACTUAL])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[JEST_REQUIRE_ACTUAL], err)
			}
			imports = append(imports, strings.ToLower(unquoted))
		case match[REQUIRE_RESOLVE] != nil:
			unquoted, err := unquoteImportString(match[REQUIRE_RESOLVE])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[REQUIRE_RESOLVE], err)
			}
			imports = append(imports, strings.ToLower(unquoted))
		case match[JEST_CREATE_MOCK_FROM_MODULE] != nil:
			unquoted, err := unquoteImportString(match[JEST_CREATE_MOCK_FROM_MODULE])
			if err != nil {
				return nil, fmt.Errorf("unquoting string literal %s from js, %v", match[JEST_CREATE_MOCK_FROM_MODULE], err)
			}
			imports = append(imports, strings.ToLower(unquoted))

		default:
			// Comment matched. Nothing to extract.
		}
	}
	sort.Strings(imports)

	return imports, nil
}

// unquoteImportString takes a string that has a complex quoting around it
// and returns a string without the complex quoting.
func unquoteImportString(quoted []byte) (string, error) {
	// Adjust quotes so that Unquote is happy. We need a double quoted string
	// without unescaped double quote characters inside.
	noQuotes := bytes.Split(quoted[1:len(quoted)-1], []byte{'"'})
	if len(noQuotes) != 1 {
		for i := 0; i < len(noQuotes)-1; i++ {
			if len(noQuotes[i]) == 0 || noQuotes[i][len(noQuotes[i])-1] != '\\' {
				noQuotes[i] = append(noQuotes[i], '\\')
			}
		}
		quoted = append([]byte{'"'}, bytes.Join(noQuotes, []byte{'"'})...)
		quoted = append(quoted, '"')
	}
	if quoted[0] == '\'' {
		quoted[0] = '"'
		quoted[len(quoted)-1] = '"'
	}

	result, err := strconv.Unquote(string(quoted))
	if err != nil {
		return "", fmt.Errorf("unquoting string literal %s from js: %v", quoted, err)
	}
	return result, err
}

const (
	IMPORT                       = 1
	REQUIRE                      = 2
	EXPORT                       = 3
	JEST_MOCK                    = 4
	JEST_REQUIRE_ACTUAL          = 5
	REQUIRE_RESOLVE              = 6
	JEST_CREATE_MOCK_FROM_MODULE = 7
)

var jsImportPattern = compileJsImportPattern()

func compileJsImportPattern() *regexp.Regexp {
	charactersPattern := ".+"
	stringLiteralPattern := `'(?:` + charactersPattern + `|")*'|"(?:` + charactersPattern + `|')*"`
	importPattern := `(?m)^import\s(?:(?:.|\n)+?from )??(?P<import>` + stringLiteralPattern + `).*?`
	requirePattern := `(?m)^\s*?(?:const .+ = )?require\((?P<require>` + stringLiteralPattern + `)\).*`
	exportPattern := `(?m)^export\s(?:(?:.|\n)+?from )??(?P<export>` + stringLiteralPattern + `).*?`
	jestMockPattern := `(?m)^\s*?(?:const .+ = )?jest.mock\((?P<jestMock>` + stringLiteralPattern + `,*)*`
	jestRequireActualPattern := `(?m)^\s*?(?:return )?jest.requireActual\((?P<jestRequireActual>` + stringLiteralPattern + `).*?`
	requireResolvePattern := `(?m)^\s*?(?:const .+ = )?require.resolve\((?P<requireResolve>` + stringLiteralPattern + `)\).*`
	jestCreateMockFromModulePattern := `(?m)^\s*?(?:const .+ = )?jest.createMockFromModule\((?P<createMockFromModule>` + stringLiteralPattern + `)\).*`

	return regexp.MustCompile(strings.Join([]string{importPattern, requirePattern, exportPattern, jestMockPattern, jestRequireActualPattern, requireResolvePattern, jestCreateMockFromModulePattern}, "|"))
}
