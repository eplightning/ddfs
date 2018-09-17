OUTPUT ?= dist
CMDS=ddfs-index ddfs-block ddfs-monitor ddfs-cli

all: $(CMDS) $(OUTPUT)/bootstrap.yml

$(OUTPUT):
	mkdir -p $(OUTPUT)

$(OUTPUT)/bootstrap.yml: configs/bootstrap.yml
	cp configs/bootstrap.yml $(OUTPUT)/bootstrap.yml

$(CMDS): $(OUTPUT)
	go build -o $(OUTPUT)/$@ ./cmd/$@

.PHONY: $(CMDS)

clean:
	rm -rf $(OUTPUT)

.PHONY: clean

proto:
	protoc -I pkg/api/ pkg/api/*.proto --gofast_out=plugins=grpc:pkg/api

.PHONY: proto
