package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
)

func createThumb(dir string, relImagePath string, cropCoords [4]float32) (thumbname string) {
	thumbCounter++
	thumbname = fmt.Sprintf("thumb%05d.jpg", thumbCounter)
	thumbPath := path.Join(dir, thumbname)
	srcPath := path.Join(*base, relImagePath)

	if _, err := os.Stat(srcPath); err != nil && os.IsNotExist(err) {
		// source file doesn't exist -- this happens when there are stale references to files in the
		// .picasa.ini that have been deleted or moved since we were running picasa
		log.Printf("Missing image (stale reference): %s\n", srcPath)
		thumbname = ""
		return
	}
	// coords are [left, top, right, bottom]
	// right-left = "percentage width" - similarly bottom-top
	const LEFT = 0
	const TOP = 1
	const RIGHT = 2
	const BOTTOM = 3
	cmd := exec.Command("magick",
		srcPath,
		"-gravity", "NorthWest",
		"-crop",
		// crop FX expressions require imagemagic7:
		fmt.Sprintf(
			"%d%%x%d%%+%%[fx:w*%f]+%%[fx:h*%f]",
			int64(100.0*(cropCoords[RIGHT]-cropCoords[LEFT])),
			int64(100.0*(cropCoords[BOTTOM]-cropCoords[TOP])),
			cropCoords[LEFT], cropCoords[TOP]),
		"-resize", "200x200",
		thumbPath)

	//log.Printf(cmd.String())
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to spawn imagemagick convert: %s: %v\n", cmd.String(), err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatalf("Failed to wait for imagemagick convert: %s: %v\n", cmd.String(), err)
	}
	return
}

func renderPersonThumb(name string, relImagePath string, cropCoords [4]float32) {
	var err error
	p := path.Join(*base, outPath)
	thumbDir := path.Join(p, "thumbs")
	if err = os.MkdirAll(thumbDir, 0777); err != nil {
		log.Fatalf("Could not create output folder %s: %v\n", thumbDir, err)
	}
	p = path.Join(p, url.QueryEscape(name)+".html")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	thumbName := createThumb(thumbDir, relImagePath, cropCoords)
	if thumbName != "" {
		f.WriteString("<a href='../" + url.QueryEscape(relImagePath) + "'><img style='width:" + thumbWidth + ";' src='thumbs/" + thumbName + "'></img></a>\n")
	}
}

func renderIndex(names []string) {
	var err error
	p := path.Join(*base, outPath)
	p = path.Join(p, "index.html")
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	for _, name := range names {
		f.WriteString("<a href='" + url.QueryEscape(name) + ".html'>" + name + "</a><br/>\n")

	}
}
