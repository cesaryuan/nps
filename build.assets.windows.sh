export GOPROXY=direct

sudo apt-get update
sudo apt-get install gcc-mingw-w64-i686 gcc-multilib

CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf windows_386_client.tar.gz npc.exe conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf windows_amd64_client.tar.gz npc.exe conf/npc.conf conf/multi_account.conf