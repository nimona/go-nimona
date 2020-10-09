module nimona.io/examples

go 1.14

require (
	github.com/arl/statsviz v0.2.1 // indirect
	github.com/gdamore/tcell v1.4.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/rivo/tview v0.0.0-20200915114512-42866ecf6ca6
	nimona.io v0.0.0
)

replace nimona.io => ../

replace github.com/mitchellh/mapstructure => github.com/geoah/mapstructure v1.3.4-rc4
