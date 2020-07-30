package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/frrad/scenesearch/lib/frame"
	"github.com/frrad/scenesearch/lib/util"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	offset := 4 * time.Second

	for i := 0; i < 10; i++ {
		offset += 100 * time.Millisecond

		f, err := frame.ExtractFrame(offset)
		if err != nil {
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}

		fname := fmt.Sprintf("frame%d.jpg", i)

		err = ioutil.WriteFile(fname, b, 0766)
		if err != nil {
			log.Fatal(err)
		}

		if x, err := util.ExecDebug("xdg-open", fname); err != nil {
			log.Fatal(err, x)
		}

		if err != nil {
			log.Fatal(err)
		}

	}

}
