# picasa-faces

A simple go program to parse .picasa.ini files and create HTML pages with thumbnail for each face and links to original image.

## Build

Requires ImageMagick 7+ for the thumbnail creation.
```
go build
```

## Run
```
./picasa-faces -base /path/to/images/
```
This will recursively visit all subdirectories under the base path, looking for .picasa.ini files.  The results will be in <base-path>/Picasa-Faces:
```
/path/to/images/Picasa-Faces
    index.html
    Joe Schmoe.html
    Jane Doe.html
    ...
    thumbs
       thumb00001.jpg
       thumb00002.jpg
       ...
```
The index contain links to each person's thumbnail page; each of those contains thumbnails cropped to the Picasa identified face.

## License

Do what you will - this is released as open source under the MIT license.
