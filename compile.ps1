foreach ($cmdlet in "cleanup", "circlefancount") {
    go build -o "bin\$cmdlet.exe" -trimpath -ldflags '-w -s' "cmd\$cmdlet\main.go"
}
