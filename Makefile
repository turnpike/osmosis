
.PHONY: clean buildpath osmosis proto docs

all: osmosis proto

clean:
	rm -rf ./build

buildpath:
	mkdir -p build

osmosis: clean buildpath
	@echo
	@echo "=========== Build Osmosis ================"
	@echo
	go build -o ./build/osmosisd ./cmd/osmosisd
	@echo
	@echo "=========== Build Complete ==============="
	@echo

proto:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-proto.sh
	@echo
	@echo "=========== Generate Complete ============"
	@echo

docs:
	@echo
	@echo "=========== Generate Message ============"
	@echo
	./scripts/generate-docs.sh

	statik -src=client/docs/static -dest=client/docs -f -m
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
    	echo "\033[92mSwagger docs are in sync\033[0m";\
    fi
	@echo
	@echo "=========== Generate Complete ============"
	@echo


###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto-all: proto-gen proto-lint proto-check-breaking proto-format
.PHONY: proto-all proto-gen proto-gen-docker proto-lint proto-check-breaking proto-format

proto-gen:
	@./scripts/protocgen.sh

proto-gen-docker:
	@echo "Generating Protobuf files"
	docker run -v $(shell pwd):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

proto-format:
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-lint:
	@buf check lint --error-format=json

proto-check-breaking:
	@buf check breaking --against '.git#branch=master'

proto-lint-docker:
	@$(DOCKER_BUF) check lint --error-format=json

proto-check-breaking-docker:
	@$(DOCKER_BUF) check breaking --against-input $(HTTPS_GIT)#branch=master

GOGO_PROTO_URL   = https://raw.githubusercontent.com/regen-network/protobuf/cosmos
COSMOS_PROTO_URL   = https://raw.githubusercontent.com/cosmos/cosmos-sdk/master/proto/cosmos

GOGO_PROTO_TYPES    = third_party/proto/gogoproto
COSMOS_PROTO_TYPES    = third_party/proto/cosmos

proto-update-deps:
	@mkdir -p $(GOGO_PROTO_TYPES)
	@curl -sSL $(GOGO_PROTO_URL)/gogoproto/gogo.proto > $(GOGO_PROTO_TYPES)/gogo.proto

	@mkdir -p $(COSMOS_PROTO_TYPES)/base/query/v1beta1/
	@curl -sSL $(COSMOS_PROTO_URL)/base/query/v1beta1/pagination.proto > $(COSMOS_PROTO_TYPES)/base/query/v1beta1/pagination.proto
