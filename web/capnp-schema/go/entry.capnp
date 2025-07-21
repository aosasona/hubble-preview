# Code generated from shared schema. DO NOT EDIT.
using Go = import "/go.capnp";
@0x89f9fd74f165e666;
$Go.package("schema");
$Go.import("go.trulyao.dev/hubble/web/schema");

enum Status {
	queued @0;
	processing @1;
	completed @2;
	failed @3;
	canceled @4;
	paused @5;
}

enum Type {
	link @0;
	audio @1;
	video @2;
	image @3;
	pdf @4;
	interchange @5;
	epub @6;
	wordDocument @7;
	presentation @8;
	spreadsheet @9;
	html @10;
	markdown @11;
	plainText @12;
	archive @13;
	code @14;
	comment @15;
	other @16;
}

struct Queue {
	status @0 :Status;
	queuedAt @1 :Int64;
}

struct Collection  {
	id @0 :Text;
	name @1 :Text;
	slug @2 :Text;
}

struct Owner {
	username @0 :Text;
	firstName @1 :Text;
	lastName @2 :Text;
}

struct Entry {
	id @0 :Text;
	name @1 :Text;
	markdown @2 :Text;
	plainText @11 :Text; # this is the plain text content of the entry, used for search indexing
	version @3 :Int32;
	type @4 :Type;
	collection @5 :Collection;
	owner @6 :Owner;
	createdAt @7 :Int64;
	queue @8 :Queue;
	filesizeBytes @9 :Int64;
	# In the case of a link, this is the URL, but in the case of a file, this is a pre-signed Minio URL to download the file.
	url @10 :Text;
}

struct UpdateEntryRequest {
	id @0 :Text;
	name @1 :Text;
	markdownContent @2 :Text; # this is the markdown content of the entry, used for rendering
	plainTextContent @4 :Text; # this is the plain text content of the entry, used for search indexing
	checksum @3 :Text;
}

struct ChunkResult {
	chunks @0 :List(Text);
}

struct EntryChunk {
	id @0 :Int32;
	entryId @1 :Text;
	index @2 :Int32;
	minimumVersion @3 :Int32;
	content @4 :Text;
	language @5 :Text;
	createdAt @6 :Int64;
}

struct CreateChunkRequest {
	entryId @0 :Text;
	index @1 :Int32;
	minimumVersion @2 :Int32;
	content @3 :Text;
	language @4 :Text;
}

struct CreateChunksRequest {
	chunks @0 :List(CreateChunkRequest);
}
