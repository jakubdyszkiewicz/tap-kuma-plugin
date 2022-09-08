# copy paste from Kuma

go_import_mapping_entries := \
	mesh/v1alpha1/selector.proto=github.com/kumahq/kuma/api/mesh/v1alpha1

comma := ,
empty:=
space := $(empty) $(empty)

go_mapping_with_spaces := $(foreach entry,$(go_import_mapping_entries),M$(entry),)
go_mapping := $(subst $(space),$(empty),$(go_mapping_with_spaces))

.PHONY: protoc
protoc:
	protoc \
	--proto_path=../../kuma/api \
	--proto_path=. \
	--go_opt=paths=source_relative \
	--go_out=plugins=grpc,$(go_mapping):. \
	tap.proto
