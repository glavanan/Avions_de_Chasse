package main

import (
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
	"regexp"
	"net/http"
	"io/ioutil"
	"os"
	"log"
	"path"
	"strings"
	)

var (
	ok = false
	determined = make(chan struct{})
)

func downloadImage(url string) {
		resp, err := http.Get(url)
		defer resp.Body.Close()

		if err != nil {
			log.Fatal("Trouble making GET photo request!")
		}

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Trouble reading reesponse body!")
		}

		filename := path.Base(url)
		if filename == "" {
			log.Fatalf("Trouble deriving file name for %s", url)
		}

//		_, err = os.Stat("/storage/emulated/0/Pictures/avions")
//		if err != nil {
//				log.Fatal("Trouble creating dir avions! -- ", err)
//		}
		err = os.Mkdir("/storage/emulated/0/Pictures/avions", 0775)
		if err != nil {
			log.Printf("Trouble creating dir avions! -- ", err)
		}


		fo, err := os.Create("storage/emulated/0/Pictures/avions/" + filename)
		if err != nil {
			log.Fatal("Trouble creating file! -- ", err)
		}
		fo.Write(contents)
		fo.Close()
}

func scrapper() (bool){

	rep, err := http.Get("http://www.mensquare.com/avionsdechasse/avions/?type=dernier")
	if err != nil {
		log.Fatal("Trouble to make get request : %s", err)
	} else {

		defer rep.Body.Close()
		contents, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			log.Fatal("Trouble reading response body")
		}

		str := string(contents)
		str_split := strings.Split(str, "\n")

		re, err := regexp.Compile("img class=\"lazy\".*data-original=\"(.*)\" />")
		if err != nil {
			log.Fatal("Trouble to create regexp")
		}

		for i := 0; len(str_split) > i ; i++ {
			if re.MatchString(str_split[i]) {
				line := strings.Split(str_split[i], "\"")
				for j := 0; len(line) > j; j++ {
					if line[j] == " data-original=" {
						carre, _ := regexp.Compile("_carre")
						res := carre.ReplaceAll([]byte(line[j+1]), []byte(""))
						downloadImage(string(res))
						return true
					}
				}
			}
		}
	}
	return false
}

func paintScreen(glctx gl.Context, sz size.Event) {
	if ok {
		glctx.ClearColor(0, 1, 0, 1)
	} else {
		glctx.ClearColor(1, 1, 1, 1)
	}
	glctx.Clear(gl.COLOR_BUFFER_BIT)
}

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		det, sz := determined, size.Event{}
		for {
			select {
				case <-det:
					paintScreen(glctx, sz)
					a.Publish()
					det = nil
				case e := <-a.Events():
					switch e := a.Filter(e).(type) {
					case lifecycle.Event:
						glctx, _ = e.DrawContext.(gl.Context)
					case size.Event:
					 sz = e
					case paint.Event:
						paintScreen(glctx, sz)
						a.Publish()
					case touch.Event:
						ok = scrapper()
						a.Send(paint.Event{})
				}
			}
		}
	})
}


