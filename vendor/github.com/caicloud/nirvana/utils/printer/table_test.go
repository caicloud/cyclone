/*
Copyright 2018 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package printer

import (
	"fmt"
	"testing"
)

func TestTablePrinter(t *testing.T) {
	lines := [][]interface{}{
		{"Index", "Programing Language", "Percentage (Change)", "URL"},
		{1, "JavaScript", "22.290% (-5.487%)"},
		{2, "Python", "15.823% (+1.355%)"},
		{3, "Java", "10.054% (-0.392%)"},
		{4, "Ruby", "7.144% (+0.293%)"},
		{5, "PHP", "7.003% (-0.494%)"},
		{6, "Go", "6.792% (+1.751%)", "https://golang.org"},
		{7, "C++", "6.653% (+1.445%)"},
		{8, "C", "3.728% (+0.232%)"},
		{9, "C#", "3.406% (-0.009%)"},
		{10, "TypeScript", "3.382% (+0.947%)"},
		{11, "Shell", "2.058% (-0.224%)"},
		{12, "Scala", "1.499% (+0.139%)"},
		{13, "Swift", "1.120% (-0.214%)"},
		{14, "DM", "1.010% (+0.693%)"},
		{15, "Rust", "0.984% (+0.095%)"},
		{16, "Objective-C", "0.846% (-0.153%)"},
		{17, "CoffeeScript", "0.532% (-0.263%)"},
		{18, "Haskell", "0.387% (+0.026%)"},
		{19, "Groovy", "0.386% (+0.114%)"},
		{20, "Lua", "0.383% (-0.006%)"},
	}
	printer := NewTable(25)
	for _, l := range lines {
		printer.AddRow(l...)
	}
	result := printer.String()
	fmt.Print(result)
}
