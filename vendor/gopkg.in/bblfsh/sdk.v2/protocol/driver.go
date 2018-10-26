package protocol

import (
	"bytes"
	"context"
	"errors"

	xcontext "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	serrors "gopkg.in/src-d/go-errors.v1"

	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes/nodesproto"
)

//go:generate protoc --proto_path=$GOPATH/src:. --gogo_out=plugins=grpc:. ./driver.proto

func RegisterDriver(srv *grpc.Server, d driver.Driver) {
	RegisterDriverServer(srv, &driverServer{d: d})
}

func AsDriver(cc *grpc.ClientConn) driver.Driver {
	return &client{c: NewDriverClient(cc)}
}

func toParseErrors(err error) []*ParseError {
	if e, ok := err.(*driver.ErrMulti); ok {
		errs := make([]*ParseError, 0, len(e.Errors))
		for _, e := range e.Errors {
			errs = append(errs, &ParseError{Text: e.Error()})
		}
		return errs
	}
	return []*ParseError{
		{Text: err.Error()},
	}
}

type driverServer struct {
	d driver.Driver
}

func (s *driverServer) Parse(ctx xcontext.Context, req *ParseRequest) (*ParseResponse, error) {
	opts := &driver.ParseOptions{
		Mode:     driver.Mode(req.Mode),
		Language: req.Language,
		Filename: req.Filename,
	}
	var resp ParseResponse
	n, err := s.d.Parse(ctx, req.Content, opts)
	resp.Language = opts.Language // can be set during the call
	if e, ok := err.(*serrors.Error); ok {
		cause := e.Cause()
		if driver.ErrDriverFailure.Is(err) {
			return nil, status.Error(codes.Internal, cause.Error())
		} else if driver.ErrTransformFailure.Is(err) {
			return nil, status.Error(codes.FailedPrecondition, cause.Error())
		} else if driver.ErrModeNotSupported.Is(err) {
			return nil, status.Error(codes.InvalidArgument, cause.Error())
		}
		if !driver.ErrSyntax.Is(err) {
			return nil, err // unknown error
		}
		// partial parse or syntax error; we will send an OK status code, but will fill Errors field
		resp.Errors = toParseErrors(cause)
	}
	buf := bytes.NewBuffer(nil)
	err = nodesproto.WriteTo(buf, n)
	if err != nil {
		return nil, err // unknown error = server failure
	}
	resp.Uast = buf.Bytes()
	return &resp, nil
}

type client struct {
	c DriverClient
}

func (c *client) Parse(ctx context.Context, src string, opts *driver.ParseOptions) (nodes.Node, error) {
	req := &ParseRequest{Content: src}
	if opts != nil {
		req.Mode = Mode(opts.Mode)
		req.Language = opts.Language
		req.Filename = opts.Filename
	}
	resp, err := c.c.Parse(ctx, req)
	if s, ok := status.FromError(err); ok {
		var kind *serrors.Kind
		switch s.Code() {
		case codes.Internal:
			kind = driver.ErrDriverFailure
		case codes.FailedPrecondition:
			kind = driver.ErrTransformFailure
		case codes.InvalidArgument:
			kind = driver.ErrModeNotSupported
		}
		if kind != nil {
			return nil, kind.Wrap(errors.New(s.Message()))
		}
	}
	if err != nil {
		return nil, err // server or network error
	}
	if opts != nil && opts.Language == "" {
		opts.Language = resp.Language
	}
	// it may be still a parsing error
	return resp.Nodes()
}

func (m *ParseResponse) Nodes() (nodes.Node, error) {
	ast, err := nodesproto.ReadTree(bytes.NewReader(m.Uast))
	if err != nil {
		return nil, err
	}
	if len(m.Errors) != 0 {
		var errs []error
		for _, e := range m.Errors {
			errs = append(errs, errors.New(e.Text))
		}
		// syntax error or partial parse - return both UAST and an error
		err = driver.ErrSyntax.Wrap(driver.JoinErrors(errs))
	}
	return ast, err
}