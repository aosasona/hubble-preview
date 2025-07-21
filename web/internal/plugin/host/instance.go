package host

import (
	"bytes"
	"context"
	"time"

	"capnproto.org/go/capnp/v3"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/schema"
	"go.trulyao.dev/seer"
)

const (
	ExportedFnOnCreate = "on_create"
)

const (
	DefaultPluginExecutionDuration = 45 * time.Second // Plugins are allowed to run for this long
)

type Instance struct {
	module  api.Module
	cleanup func() error
}

type OnCreateArgs struct {
	*models.Entry
	URL string
}

func (i *Instance) Close() error {
	return i.cleanup()
}

func (i *Instance) allocateEntry(args *OnCreateArgs) (*bytes.Buffer, error) {
	arena := capnp.SingleSegment(nil)
	msg, seg, err := capnp.NewMessage(arena)
	if err != nil {
		return nil, seer.Wrap("alloc_capnp_message", err)
	}

	// Create the shared entry
	entry, err := schema.NewRootEntry(seg)
	if err != nil {
		return nil, seer.Wrap("new_root_entry", err)
	}

	collection, err := schema.NewCollection(seg)
	if err != nil {
		return nil, seer.Wrap("new_root_collection", err)
	}

	owner, err := schema.NewOwner(seg)
	if err != nil {
		return nil, seer.Wrap("new_root_owner", err)
	}

	queue, err := schema.NewQueue(seg)
	if err != nil {
		return nil, seer.Wrap("new_root_queue", err)
	}

	// Set the owner fields
	if err = owner.SetFirstName(args.AddedBy.FirstName); err != nil {
		return nil, seer.Wrap("set_first_name_in_owner", err)
	}
	if err = owner.SetLastName(args.AddedBy.LastName); err != nil {
		return nil, seer.Wrap("set_last_name_in_owner", err)
	}
	if err = owner.SetUsername(args.AddedBy.Username); err != nil {
		return nil, seer.Wrap("set_username_in_owner", err)
	}

	// Set the collection fields
	if err = collection.SetId(args.Collection.ID.String()); err != nil {
		return nil, seer.Wrap("set_id_in_collection", err)
	}
	if err = collection.SetName(args.Collection.Name); err != nil {
		return nil, seer.Wrap("set_name_in_collection", err)
	}
	if err = collection.SetSlug(args.Collection.Slug); err != nil {
		return nil, seer.Wrap("set_slug_in_collection", err)
	}

	// Set the queue fields
	queue.SetQueuedAt(args.QueuedAt.Unix())
	queue.SetStatus(args.Status.ToCapnpType())

	// Set all the fields
	if err = entry.SetId(args.PublicID.String()); err != nil {
		return nil, seer.Wrap("set_id_in_entry", err)
	}
	if err = entry.SetName(args.Name); err != nil {
		return nil, seer.Wrap("set_name_in_entry", err)
	}
	if err = entry.SetMarkdown(args.Content.String); err != nil {
		return nil, seer.Wrap("set_content_in_entry", err)
	}
	if err = entry.SetPlainText(args.TextContent.String); err != nil {
		return nil, seer.Wrap("set_plain_text_in_entry", err)
	}

	entry.SetVersion(args.Version)
	entry.SetType(args.Type.CapnpType())
	if err = entry.SetCollection(collection); err != nil {
		return nil, seer.Wrap("set_collection_in_entry", err)
	}
	if err = entry.SetOwner(owner); err != nil {
		return nil, seer.Wrap("set_owner_in_entry", err)
	}
	if err = entry.SetQueue(queue); err != nil {
		return nil, seer.Wrap("set_queue_in_entry", err)
	}
	entry.SetCreatedAt(args.CreatedAt.Unix())
	entry.SetFilesizeBytes(args.FilesizeBytes)

	if err = entry.SetUrl(args.URL); err != nil {
		return nil, seer.Wrap("set_url_in_entry", err)
	}

	var buf bytes.Buffer
	if err := capnp.NewEncoder(&buf).Encode(msg); err != nil {
		return nil, seer.Wrap("encode_capnp_message", err)
	}

	return &buf, nil
}

func (i *Instance) OnCreate(ctx context.Context, args *OnCreateArgs) error {
	pluginCtx, cancel := context.WithTimeout(ctx, DefaultPluginExecutionDuration)
	defer cancel()

	// Lookup the on_create function in the module
	onCreateFn := i.module.ExportedFunction("on_create")
	if onCreateFn == nil {
		return seer.New("get_on_create_exported_fn", "on_create function not found in module")
	}

	buf, err := i.allocateEntry(args)
	if err != nil {
		return err
	}

	// Allocate memory for the message in the module
	ptr, err := alloc.Allocate(pluginCtx, i.module, uint64(buf.Len()))
	if err != nil {
		return seer.Wrap("allocate_in_on_create", err)
	}

	// Write the message to the allocated memory
	if !i.module.Memory().Write(uint32(ptr), buf.Bytes()[:buf.Len()]) {
		return seer.New("write_capnp_message_to_memory", "failed to write message to plugin memory")
	}

	result, err := onCreateFn.Call(pluginCtx, ptr, uint64(buf.Len()))
	if err != nil {
		return seer.Wrap("call_on_create_fn", err)
	}

	if len(result) != 1 {
		return seer.New("call_on_create_fn", "unexpected number of results")
	}

	// if the plugin returns 0, it means success
	errPtr := result[0]
	if errPtr == 0 {
		return nil
	}

	// We need to read the error string from the plugin memory otherwise
	outPtr, outSize := alloc.DecodePtrWithSize(errPtr)
	if outSize == 0 {
		return seer.New("read_on_create_error", "size of error string is 0")
	}
	defer alloc.Deallocate(pluginCtx, i.module, outPtr, outSize) // free the memory

	outBuf, ok := i.module.Memory().Read(uint32(outPtr), uint32(outSize))
	if !ok {
		return seer.New(
			"read_capnp_message_from_memory",
			"failed to read message from plugin memory",
		)
	}

	return seer.New("on_create_error", string(outBuf))
}
