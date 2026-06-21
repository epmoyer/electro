// For internal go apps we consistently use `module app` so that all our
// internal packages can use `import "app/pkg/{package_name}"` to import
// project packages. This allows us to copy internal packages between
// projects without having to edit the import paths.
module app

go 1.24.4

require (
	github.com/FurqanSoftware/goldmark-d2 v0.0.0-20250906161746-6305edf4a24a
	github.com/PuerkitoBio/goquery v1.10.3
	github.com/epmoyer/quicklog/v2 v2.1.0
	github.com/tdewolff/minify/v2 v2.24.4
	github.com/yuin/goldmark v1.7.13
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
	golang.org/x/net v0.43.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	oss.terrastruct.com/d2 v0.7.1
)

require (
	github.com/alecthomas/chroma/v2 v2.20.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/andybalholm/cascadia v1.3.3 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/dop251/goja v0.0.0-20250630131328-58d95d85e994 // indirect
	github.com/epmoyer/callsite v1.0.0 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4+incompatible // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/pprof v0.0.0-20250903194437-c28834ac2320 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mazznoer/csscolorparser v0.1.6 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/tdewolff/parse/v2 v2.8.4 // indirect
	golang.org/x/exp v0.0.0-20250819193227-8b4c13bb791b // indirect
	golang.org/x/image v0.30.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/xerrors v0.0.0-20240903120638-7835f813f4da // indirect
	oss.terrastruct.com/util-go v0.0.0-20250213174338-243d8661088a // indirect
)
