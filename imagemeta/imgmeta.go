package main

import (
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	imagepathArg = ""
	renameArg    = false
)

func decodeImage(fname string) (*exif.Exif, error) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}

	// Optionally register camera makenote data parsing - currently Nikon and
	// Canon are supported.
	exif.RegisterParsers(mknote.All...)

	return exif.Decode(f)
}

func printUsage() {
	fmt.Println("Options")
	fmt.Println(" -imagepath : image files path")
	fmt.Println(" -rename")
	fmt.Println()
	fmt.Println("Usage : imgmeta -imagepath [dir path] -rename")
}

func isFileExist(f string) bool {
	_, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func addedDuplicatedMark(newFileName string) string {
	name := newFileName

	i := 1
	for {
		if isFileExist(name) == false {
			return name
		}

		name = newFileName + fmt.Sprintf("(%v)", i)
		i++
	}
}

func main() {

	flag.StringVar(&imagepathArg, "imagepath", "", "File-path of image")
	flag.BoolVar(&renameArg, "rename", false, "rename image file name")

	flag.Parse()

	if imagepathArg == "" {
		printUsage()
		os.Exit(1)
	}

	fmt.Println("images path = ", imagepathArg)

	fileInfo, err := ioutil.ReadDir(imagepathArg)
	if err != nil {
		os.Exit(1)
	}

	fmt.Println()

	for _, file := range fileInfo {
		filename := file.Name()
		if strings.EqualFold(filepath.Ext(filename), ".JPG") {
			x, err := decodeImage(filename)
			if err != nil {
				log.Fatal(err)
			}

			lat, long, _ := x.LatLong()
			gps := fmt.Sprintf("lat(%v), long(%v)", lat, long)
			if lat == 0 && long == 0 {
				gps = "No GPS"
			}

			tm, err := x.DateTime()
			if err != nil {
				fmt.Printf("%v - No taken-time, %v\n", filename, gps)
			} else {
				if renameArg == true {
					newFileName := fmt.Sprintf("%v-%02d-%02d %02d:%02d:%02d%v", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second(), filepath.Ext(filename))

					if filename != newFileName {
						if isFileExist(newFileName) == true {
							addedDuplicatedMark(newFileName)
						}

						err := os.Rename(filename, newFileName)
						if err != nil {
							log.Fatal(err)
						}

						fmt.Printf("%v : %v ---> %v\n", filename, gps, newFileName)
					}
					fmt.Printf("%v : %v\n", filename, gps)
				} else {
					fmt.Printf("%v : %v\n", filename, gps)
				}
			}
		}
	}
}
