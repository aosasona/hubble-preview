# Code generated from shared schema. DO NOT EDIT.
using Go = import "/go.capnp";
@0xba385a997f42f53b;
$Go.package("schema");
$Go.import("go.trulyao.dev/hubble/web/schema");

struct StoreGetResponse {
	value @0 :Text;
	exists @1 :Bool;
}

struct StoreSetRequest {
	key @0 :Text;
	value @1 :Text;
}

struct StoreKvPair {
	key @0 :Text;
	value @1 :Text;
}

struct StoreAllResponse {
	pairs @0 :List(StoreKvPair);
}
