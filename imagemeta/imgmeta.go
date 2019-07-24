package main

import (
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/mknote"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

	defer f.Close()

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

func getNewName(path string, name string, ext string) string {
	nameExceptExt := filepath.Join(path, name)
	fullname := nameExceptExt + ext

	i := 1
	for {
		if isFileExist(fullname) == false {
			return fullname
		}

		fullname = nameExceptExt + fmt.Sprintf("(%v)", i) + ext
		i++
	}
}

func isJpgFile(filename string) bool {
	extName := filepath.Ext(filename)
	return strings.EqualFold(extName, ".JPG") || strings.EqualFold(extName, ".JPEG")
}

func isMp4File(filename string) bool {
	extName := filepath.Ext(filename)
	return strings.EqualFold(extName, ".MP4")
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
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
		fullpath := filepath.Join(imagepathArg, filename)

		// 이미지 파일 처리
		if isJpgFile(filename) {
			x, err := decodeImage(fullpath)
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
					newPureFileName := fmt.Sprintf("%v-%02d-%02d %02d.%02d.%02d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
					newFileName := newPureFileName + filepath.Ext(filename)

					if filename != newFileName {
						newFullPathName := getNewName(imagepathArg, newPureFileName, filepath.Ext(filename))
						//copy(fullpath, newFullPathName)
						err := os.Rename(fullpath, newFullPathName)
						if err != nil {
							log.Fatal(err)
						}

						_, newFileName := filepath.Split(newFileName)
						fmt.Printf("%v : %v ---> %v\n", filename, gps, newFileName)
					} else {
						fmt.Printf("%v : %v\n", filename, gps)
					}
				} else {
					fmt.Printf("%v : %v\n", filename, gps)
				}
			}
		} else if isMp4File(filename) {
			// 20190602_140042.mp4 형태인 경우
			match, _ := regexp.MatchString("^\\d{8}_\\d{6}\\.", filename)
			if renameArg == true && match == true {
				year := filename[:4]
				month := filename[4:6]
				day := filename[6:8]
				hour := filename[9:11]
				min := filename[11:13]
				sec := filename[13:15]

				newPureFileName := fmt.Sprintf("%v-%v-%v %v.%v.%v", year, month, day, hour, min, sec)

				newFullPathName := getNewName(imagepathArg, newPureFileName, filepath.Ext(filename))
				err := os.Rename(fullpath, newFullPathName)
				if err != nil {
					log.Fatal(err)
				}
				_, newFileName := filepath.Split(newFullPathName)

				fmt.Printf("%v ---> %v\n", filename, newFileName)
			}
		}
	}
}
