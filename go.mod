// For internal go apps we consistently use `module app` so that all our
// internal packages can use `import "app/pkg/{package_name}"` to import
// project packages. This allows us to copy internal packages between
// projects without having to edit the import paths.
module app

go 1.24

require (
	github.com/PuerkitoBio/goquery v1.10.3
	github.com/tdewolff/minify/v2 v2.24.4
	github.com/yuin/goldmark v1.7.13
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
	golang.org/x/net v0.39.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	github.com/alecthomas/chroma/v2 v2.2.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/dlclark/regexp2 v1.7.0 // indirect
	github.com/tdewolff/parse/v2 v2.8.4 // indirect
)
