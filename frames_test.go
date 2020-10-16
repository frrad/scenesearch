package main

import (
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

	if tabStr != ` 	<table>
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
` {
		t.Errorf("%s", tabStr)
	}
}
