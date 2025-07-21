@0xb9e89c17977476da;

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
