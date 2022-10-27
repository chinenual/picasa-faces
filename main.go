package main

import (
	"flag"
	"gopkg.in/ini.v1"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const outPath = "Picasa-Faces"
const thumbWidth = "100px"

var thumbCounter = -1

var base = flag.String("base", "", "root directory of the source images/.picasa.ini files")

var peopleMap map[string]string = map[string]string{}
var peopleImagesMap map[string][]string = map[string][]string{}

// lots of good info on the format here:
//      https://gist.github.com/fbuchinger/1073823/9986cc61ae67afeca2f4a2f984d7b5d4a818d4f0

// convert a rect64(xxxxx) has to [left, top, right, bottom] percentages
func convertRect64(v string) (result [4]float32) {
	// v is a 64-bit unsigned hex in a "rect64(...)" wrapper
	// can't just convert 4 characters at a time since Picasa sometimes omits leading 0's -
	// so parse the string as a 64bit number, then convert that
	var u uint64
	var err error
	v = strings.ReplaceAll(v, ")", "")
	v = strings.ReplaceAll(v, "rect64(", "")
	if u, err = strconv.ParseUint(v, 16, 64); err != nil {
		log.Fatalf("Can't parse rect64: %s: %v", v, err)
	}

	result[0] = float32((u>>48)&0xffff) / 65535.0
	result[0] = float32((u>>48)&0xffff) / 65535.0
	result[1] = float32((u>>32)&0xffff) / 65535.0
	result[2] = float32((u>>16)&0xffff) / 65535.0
	result[3] = float32(u&0xffff) / 65535.0

	return
}

func process(iniFile string) {
	var relativeDir string
	if path.Dir(iniFile) == path.Dir(*base) {
		relativeDir = "."
	} else {
		relativeDir = path.Join(".", strings.TrimPrefix(path.Dir(iniFile), *base))
	}
	log.Println(relativeDir)
	var opts = ini.LoadOptions{IgnoreInlineComment: true}
	cfg, err := ini.LoadSources(opts, iniFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v\n", err)
	}
	s := cfg.Section("Contacts2")
	var contactMap map[string]string
	if s != nil {
		contactMap = s.KeysHash()
		for k, name := range contactMap {
			// strip trailing semicolons:
			cleaned := strings.ReplaceAll(name, ";", "")
			peopleMap[cleaned] = cleaned
			contactMap[k] = cleaned
		}
		// faces picasa found but we never gave a name to use this hash
		const UNKNOWN_HASH = "ffffffffffffffff"
		peopleMap["UNKNOWN"] = "UNKNOWN"
		contactMap[UNKNOWN_HASH] = "UNKNOWN"
		//log.Printf("Contacts: %v\n", s.KeysHash())
	}
	// loop through the sections in sorted order so that images end up in a reasonable order
	secNames := cfg.SectionStrings()
	sort.Strings(secNames)

	for _, secName := range secNames {
		s := cfg.Section(secName)
		if s.HasKey("faces") {
			l := s.Key("faces").String()
			list := strings.Split(l, ";")
			for _, pair := range list {
				v := strings.Split(pair, ",")
				coords := convertRect64(v[0])
				person := contactMap[v[1]]
				relativeImagePath := path.Join(relativeDir, s.Name())
				peopleImagesMap[person] = append(peopleImagesMap[person], relativeImagePath)

				renderPersonThumb(person, relativeImagePath, coords)
			}
		}
	}
}

func processFiles(dir string) {
	log.Printf("Scanning for .picasa.ini files under %s ...\n", *base)
	var inis []string
	e := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && ".picasa.ini" == info.Name() {
			inis = append(inis, path)
		}
		return nil
	})
	if e != nil {
		log.Fatal(e)
	}
	sort.Strings(inis)
	// process the ini's in sorted order so we can just emit them to the output files as we go and
	// the results will be ordered reasonably
	for _, f := range inis {
		process(f)
	}
}

func main() {
	flag.Parse()
	if *base == "" {
		log.Fatal("-base must be set")
	}

	processFiles(*base)

	var people []string
	for k, _ := range peopleMap {
		people = append(people, k)
	}
	sort.Strings(people)
	//log.Printf("result: %v\n", people)
	renderIndex(people)
}
