package main

import (
	"html/template"
	"testing"
)

func TestFrameTable(t *testing.T) {
	x := frameRange{
		StartOffset: 0,
		EndOffset:   100,

		File:  "asdf.mp4",
		Shots: 3,
		Width: 123,
	}

	tabStr := x.Table()

	expected := `<table>
<tr>
    
    <td>0</td>
    
    <td>50</td>
    
    <td>100</td>
    
</tr>

<tr>
    
    <td><img src="/frame?offset=0&amp;file=asdf.mp4" width="123px"></td>
    
    <td><img src="/frame?offset=50&amp;file=asdf.mp4" width="123px"></td>
    
    <td><img src="/frame?offset=100&amp;file=asdf.mp4" width="123px"></td>
    
</tr>
</table>
`

	if tabStr != template.HTML(expected) {
		if len(tabStr) != len(expected) {
			t.Error("different len", len(tabStr), len(expected))
		}

		maxLen := len(expected)
		if len(tabStr) > maxLen {
			maxLen = len(tabStr)
		}

		for i := 0; i < maxLen; i++ {
			if i+1 > len(tabStr) {
				t.Error("expected:", i, expected[i])
				continue
			}

			if i+1 > len(expected) {
				t.Error("tabStr:", i, tabStr[i])
				continue
			}

			if tabStr[i] != expected[i] {
				t.Error(i, tabStr[i], expected[i])
			}
		}

		t.Errorf("%s", tabStr)
	}
}
