@0xfdb92caa514d8128;

enum NetworkMethod {
	get @0;
	post @1;
}

struct NetworkHeader {
	key @0 :Text;
	value @1 :Text;
}

struct NetworkRequest {
	method @0 :NetworkMethod;
	url @1 :Text;
	headers @2 :List(NetworkHeader);
	body @3 :Data;
}

struct NetworkResponse {
	status @0 :Int32;
	headers @1 :List(NetworkHeader);
	body @2 :Data;
}
