// For internal go apps we consistently use `module app` so that all our
// internal packages can use `import "app/pkg/{package_name}"` to import
// project packages. This allows us to copy internal packages between
// projects without having to edit the import paths.
module app

go 1.24

require (
	github.com/yuin/goldmark v1.7.13
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)
