package boundhost

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/schema"
)

/*
Makes a network request and returns the response.

NOTE: this is required since WASM modules cannot make network requests directly.

Signature: fn(request: Capnp::NetworkRequest) -> Capnp::NetworkResponse

Exported as: `network_request`
*/
func (b *BoundHost) makeRequest(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	logger := b.HostFnLogger(spec.PermNetworkRequest)

	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		logger.error(ErrFailedToReadMemory)
		return 0
	}
	defer b.dealloc(ctx, m, logger, offset, byteCount) // DEALLOCATE MEMORY

	msg, err := capnp.Unmarshal(buf)
	if err != nil {
		logger.error(err, "failed to unmarshal capnp message")
		return 0
	}

	request, err := schema.ReadRootNetworkRequest(msg)
	if err != nil {
		logger.error(err, "failed to read network request")
		return 0
	}

	method := request.Method()
	body, _ := request.Body()
	link, _ := request.Url()
	headers, _ := request.Headers()

	if link == "" {
		logger.error(errors.New("empty URL"), "failed to make network request")
		return 0
	}

	req, err := http.NewRequest(strings.ToUpper(method.String()), link, bytes.NewBuffer(body))
	if err != nil {
		logger.error(err, "failed to create network request")
		return 0
	}

	// Write the headers to a http.Header
	for i := range headers.Len() {
		key, _ := headers.At(i).Key()
		value, _ := headers.At(i).Value()
		req.Header.Add(key, value)
	}

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.error(err, "failed to send network request")
		return 0
	}
	defer resp.Body.Close() //nolint:errcheck

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.error(err, "failed to read network response body")
		return 0
	}

	// Encode the capnp message for the response
	arena := capnp.SingleSegment(nil)
	msg, seg, err := capnp.NewMessage(arena)
	if err != nil {
		logger.error(err, "failed to create capnp message")
		return 0
	}

	response, err := schema.NewRootNetworkResponse(seg)
	if err != nil {
		logger.error(err, "failed to create network response root message")
		return 0
	}
	response.SetStatus(int32(resp.StatusCode))
	if err = response.SetBody(responseBody); err != nil {
		logger.error(err, "failed to set network response body")
	}

	// Set the headers
	headersList, err := response.NewHeaders(int32(len(resp.Header)))
	if err != nil {
		logger.error(err, "failed to create network response headers")
		return 0
	}

	j := 0
	for k, v := range resp.Header {
		header, err := schema.NewNetworkHeader(seg)
		if err != nil {
			logger.error(err, "failed to create network header container")
			return 0
		}
		_ = header.SetKey(k)
		_ = header.SetValue(strings.Join(v, ","))

		if err := headersList.Set(j, header); err != nil {
			logger.error(err, "failed to set network response header at index "+strconv.Itoa(j))
			return 0
		}
		j++
	}

	if err := response.SetHeaders(headersList); err != nil {
		logger.error(err, "failed to set network response headers")
		return 0
	}

	// Serialize the message to a byte slice
	var output bytes.Buffer
	if err := capnp.NewEncoder(&output).Encode(msg); err != nil {
		logger.error(err, "failed to encode capnp message")
		return 0
	}

	// Allocate memory for the message in the module
	ptr, err := alloc.Allocate(ctx, m, uint64(output.Len()))
	if err != nil {
		logger.error(err, "failed to allocate memory for network response")
		return 0
	}

	// Write the message to the allocated memory
	if !m.Memory().Write(uint32(ptr), output.Bytes()[:output.Len()]) {
		logger.error(err, "failed to write network response to memory")
		return 0
	}

	return alloc.EncodePtrWithSize(ptr, uint64(output.Len()))
}
