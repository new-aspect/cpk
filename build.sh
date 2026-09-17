go build -ldflags "-X main.Version=v0.1.0" -o cpk .
sudo mv cpk /usr/local/bin/cpk