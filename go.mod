module github.com/bgpfix/bgpipe

go 1.21.0

require (
	github.com/bgpfix/bgpfix v0.0.0
	github.com/spf13/pflag v1.0.5
)

require (
	github.com/knadh/koanf/maps v0.1.1 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
)

require (
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/knadh/koanf/providers/posflag v0.1.0
	github.com/knadh/koanf/v2 v2.0.1
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/puzpuzpuz/xsync/v2 v2.4.1 // indirect
	github.com/rs/zerolog v1.30.0
	golang.org/x/sync v0.3.0
	golang.org/x/sys v0.11.0
)

replace github.com/bgpfix/bgpfix => ../bgpfix
