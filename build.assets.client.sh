export GOPROXY=direct

sudo apt-get update
sudo apt-get install gcc-mingw-w64-i686 gcc-multilib

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static"  ./cmd/npc/npc.go

tar -czvf linux_amd64_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_386_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf freebsd_386_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf freebsd_amd64_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf freebsd_arm_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_arm_v7_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_arm_v6_client.tar.gz npc conf/npc.conf conf/multi_account.conf

CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=5 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_arm_v5_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_arm64_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_mips64_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_mips64le_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_mipsle_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf linux_mips_client.tar.gz npc conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf windows_386_client.tar.gz npc.exe conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf windows_amd64_client.tar.gz npc.exe conf/npc.conf conf/multi_account.conf


CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -extldflags -static" ./cmd/npc/npc.go

tar -czvf darwin_amd64_client.tar.gz npc conf/npc.conf conf/multi_account.conf