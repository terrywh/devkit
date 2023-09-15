//go:build windows
// +build windows

package entity

func (o *Shell) ApplyDefaults() {
	if len(o.Cmd) < 1 {
		o.Cmd = []string{"C:\\Windows\\System32\\cmd.exe"}
	}
	if o.Row < 16 {
		o.Row = 16
	}
	if o.Col < 96 {
		o.Col = 96
	}
}
