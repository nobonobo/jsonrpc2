package jsonrpc2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"sync"
	"testing"
)

type testcase struct {
	id      interface{}
	numArgs int
	name    string
	f       interface{}
	params  interface{}
	reply   interface{}
	resp    string
	req     string
	err     string
}

type Struct struct {
	Text    string `json:"text,omitempty"`
	Number  int    `json:"number,omitempty"`
	Boolean bool   `json:"boolean,omitempty"`
}

type BadStruct struct {
	Text string
}

func (b BadStruct) MarshalJSON() ([]byte, error) {
	return nil, nil
}

type unexported struct{}

var serveTestcases = []testcase{
	// 1 arg, 2 returns
	{
		id:      1,
		numArgs: 1,
		name:    "nil_string",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":1,"result":"string"}`,
		f: func(ctx context.Context) (string, error) {
			return "string", nil
		},
	},
	{
		id:      2,
		numArgs: 1,
		name:    "nil_int",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":2,"result":33}`,
		f: func(ctx context.Context) (int, error) {
			return 33, nil
		},
	},
	{
		id:      "3",
		numArgs: 1,
		name:    "nil_struct",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":"3","result":{"text":"text","number":33,"boolean":true}}`,
		f: func(ctx context.Context) (Struct, error) {
			return Struct{Text: "text", Number: 33, Boolean: true}, nil
		},
	},
	{
		id:      "randomid",
		numArgs: 1,
		name:    "nil_ptrstruct",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":"randomid","result":{"text":"text","boolean":true}}`,
		f: func(ctx context.Context) (*Struct, error) {
			return &Struct{Text: "text", Boolean: true}, nil
		},
	},
	{
		id:      6,
		numArgs: 1,
		name:    "nil_struct_stderror",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":6,"error":{"code":-32000,"message":"something went wrong"}}`,
		f: func(ctx context.Context) (Struct, error) {
			return Struct{}, errors.New("something went wrong")
		},
	},
	{
		id:      7,
		numArgs: 1,
		name:    "nil_struct_liberror",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":7,"error":{"code":-32603,"message":"Internal error"}}`,
		f: func(ctx context.Context) (Struct, error) {
			return Struct{}, ErrInternalError
		},
	},
	{
		id:      8,
		numArgs: 1,
		name:    "nil_struct_liberror2",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":8,"error":{"code":-32602,"message":"Invalid params"}}`,
		f: func(ctx context.Context) (Struct, error) {
			return Struct{}, ErrInvalidParams
		},
	},
	{
		id:      9,
		numArgs: 1,
		name:    "nil_struct_customerror",
		params:  nil,
		resp:    `{"jsonrpc":"2.0","id":9,"error":{"code":-32300,"message":"Something went wrong","data":[1,2,3]}}`,
		f: func(ctx context.Context) (Struct, error) {
			return Struct{}, &Error{Code: -32300, Message: "Something went wrong", Data: []int{1, 2, 3}}
		},
	},
	// 2 args, 2 returns
	{
		id:      "nanoid",
		numArgs: 2,
		name:    "string_string",
		params:  "input",
		resp:    `{"jsonrpc":"2.0","id":"nanoid","result":"input"}`,
		f: func(ctx context.Context, s string) (string, error) {
			return s, nil
		},
	},
	{
		id:      10,
		numArgs: 2,
		name:    "int_int",
		params:  33,
		resp:    `{"jsonrpc":"2.0","id":10,"result":33}`,
		f: func(ctx context.Context, n int) (int, error) {
			return n, nil
		},
	},
	{
		id:      11,
		numArgs: 2,
		name:    "struct_struct",
		params:  Struct{Text: "text", Number: 33},
		resp:    `{"jsonrpc":"2.0","id":11,"result":{"text":"text","number":33}}`,
		f: func(ctx context.Context, s Struct) (Struct, error) {
			return s, nil
		},
	},
	{
		id:      33,
		numArgs: 2,
		name:    "ptrstruct_struct",
		params:  &Struct{Text: "text", Number: 33},
		resp:    `{"jsonrpc":"2.0","id":33,"result":{"text":"text","number":33}}`,
		f: func(ctx context.Context, s *Struct) (Struct, error) {
			return *s, nil
		},
	},
	{
		id:      nil,
		numArgs: 2,
		name:    "notification_struct_struct",
		params:  Struct{Text: "text", Number: 33},
		resp:    ``,
		f: func(ctx context.Context, s Struct) (Struct, error) {
			return s, nil
		},
	},
}

var serveErrTestcases = []testcase{
	{
		numArgs: 1,
		name:    "parse_error",
		req:     `invalid_json`,
		resp:    `{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"Parse error"}}`,
		f: func(ctx context.Context) (string, error) {
			return "string", nil
		},
	},
	{
		numArgs: 2,
		name:    "method_not_found",
		req:     `{"jsonrpc":"2.0","id":1,"method":"garbage_text","params":[]}`,
		resp:    `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"Method not found"}}`,
		f: func(ctx context.Context, s string) (string, error) {
			return "string", nil
		},
	},
	{
		numArgs: 2,
		name:    "method_not_found_without_id",
		req:     `{"jsonrpc":"2.0","method":"garbage_text","params":[]}`,
		resp:    `{"jsonrpc":"2.0","id":null,"error":{"code":-32601,"message":"Method not found"}}`,
		f: func(ctx context.Context, s string) (string, error) {
			return "string", nil
		},
	},
	{
		numArgs: 2,
		name:    "missing_method",
		req:     `{"jsonrpc":"2.0","id":1,"params":[]}`,
		resp:    `{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"Invalid Request"}}`,
		f: func(ctx context.Context, s string) (string, error) {
			return "string", nil
		},
	},
	{
		numArgs: 2,
		name:    "invalid_params",
		req:     `{"jsonrpc":"2.0","id":1,"method":"invalid_params","params":[1,2]}`,
		resp:    `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params"}}`,
		f: func(ctx context.Context, s string) (string, error) {
			return "string", nil
		},
	},
	{
		numArgs: 2,
		name:    "invalid_params_struct",
		req:     `{"jsonrpc":"2.0","id":1,"method":"invalid_params_struct","params":{}}`,
		resp:    `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"Invalid params"}}`,
		f: func(ctx context.Context, s Struct) (Struct, error) {
			return Struct{}, nil
		},
	},
	{
		numArgs: 2,
		name:    "invalid_output",
		req:     `{"jsonrpc":"2.0","id":1,"method":"invalid_output","params":"input"}`,
		resp:    `{"jsonrpc":"2.0","id":1,"error":{"code":-32603,"message":"Internal error"}}`,
		f: func(ctx context.Context, s string) (BadStruct, error) {
			return BadStruct{}, nil
		},
	},
}

var handleFuncErrTestcases = []testcase{
	{
		name: "invalid_handler_type",
		err:  "jsonrpc: invalid handler type: expected func, got string",
		f:    "invalid",
	},
	{
		name: "invalid_num_args",
		err:  "jsonrpc: invalid number of args: expected 2, got 0",
		f: func() (string, error) {
			return "", nil
		},
	},
	{
		name: "invalid_first_arg_type",
		err:  "jsonrpc: invalid first arg type: expected context.Context, got string",
		f: func(s string) (string, error) {
			return "", nil
		},
	},
	{
		name: "invalid_second_arg_type",
		err:  "jsonrpc: invalid second arg type: expected exported or builtin",
		f: func(ctx context.Context, params unexported) (string, error) {
			return "", nil
		},
	},
	{
		name: "invalid_num_returns",
		err:  "jsonrpc: invalid number of returns: expected 2, got 3",
		f: func(ctx context.Context, params string) (string, string, string) {
			return "", "", ""
		},
	},
	{
		name: "invalid_first_return_type",
		err:  "jsonrpc: invalid first return type: expected exported or builtin",
		f: func(ctx context.Context, params string) (unexported, error) {
			return unexported{}, nil
		},
	},
	{
		name: "invalid_second_return_type",
		err:  "jsonrpc: invalid second return type: expected error, got string",
		f: func(ctx context.Context, params string) (string, string) {
			return "", ""
		},
	},
}

func TestHandleFunc(t *testing.T) {
	server := NewServer()

	for _, tc := range serveTestcases {
		t.Run(tc.name, func(t *testing.T) {
			err := server.HandleFunc(tc.name, tc.f)
			if err != nil {
				t.Errorf("method %v registration failed: %v", tc.name, err)
			}
			v, ok := server.handler.Load(tc.name)
			if !ok {
				t.Errorf("method %v not registered", tc.name)
			}
			htype, ok := v.(handlerType)
			if !ok {
				t.Errorf("handler with wrong type")
			}
			if htype.numArgs != tc.numArgs {
				t.Errorf("handlerType with incorrect numArgs: \ngot: %v\nwant: %v\n", htype.numArgs, tc.numArgs)
			}
		})
	}
}

func TestHandleFuncErr(t *testing.T) {
	server := NewServer()

	for _, tc := range handleFuncErrTestcases {
		t.Run(tc.name, func(t *testing.T) {
			err := server.HandleFunc(tc.name, tc.f)
			if err.Error() != tc.err {
				t.Errorf("invalid registration error:\ngot: %v\nwant: %v\n", err, tc.err)
			}
		})
	}
}

func TestServeErr(t *testing.T) {
	server := NewServer()

	for _, tc := range serveErrTestcases {
		server.HandleFunc(tc.name, tc.f)
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "locahost:8080", bytes.NewReader([]byte(tc.req)))
			rw := httptest.NewRecorder()
			server.ServeHTTP(rw, req)
			want := tc.resp

			if got := rw.Body.String(); got != want {
				t.Errorf("invalid jsonrpc response: \ngot: %v\nwant: %v\n", got, want)
			}
		})
	}
}

func TestServeSync(t *testing.T) {
	type request struct {
		VersionTag string      `json:"jsonrpc"`
		Method     string      `json:"method"`
		ID         interface{} `json:"id,omitempty"`
		Params     interface{} `json:"params"`
	}

	server := NewServer()
	for _, h := range serveTestcases {
		server.HandleFunc(h.name, h.f)
	}

	for _, tc := range serveTestcases {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(&request{
				VersionTag: "2.0",
				Method:     tc.name,
				ID:         tc.id,
				Params:     tc.params,
			})
			if err != nil {
				t.Errorf("marshaling req body: %v", err)
			}
			req := httptest.NewRequest("POST", "locahost:8080", bytes.NewReader(body))
			rw := httptest.NewRecorder()
			server.ServeHTTP(rw, req)
			want := tc.resp

			if got := rw.Body.String(); got != want {
				t.Errorf("invalid jsonrpc response: \ngot: %v\nwant: %v\n", got, want)
			}
		})
	}
}

func TestServeAsync(t *testing.T) {
	type request struct {
		VersionTag string      `json:"jsonrpc"`
		Method     string      `json:"method"`
		ID         interface{} `json:"id,omitempty"`
		Params     interface{} `json:"params"`
	}

	server := NewServer()
	for _, h := range serveTestcases {
		server.HandleFunc(h.name, h.f)
	}

	var wg sync.WaitGroup
	wg.Add(len(serveTestcases))
	for _, tc := range serveTestcases {
		go func(tc testcase, id interface{}) {
			t.Run(tc.name, func(t *testing.T) {
				body, err := json.Marshal(&request{
					VersionTag: "2.0",
					Method:     tc.name,
					ID:         id,
					Params:     tc.params,
				})
				if err != nil {
					t.Errorf("marshaling req body: %v", err)
				}
				req := httptest.NewRequest("POST", "locahost:8080", bytes.NewReader(body))
				rw := httptest.NewRecorder()
				server.ServeHTTP(rw, req)
				want := tc.resp

				if got := rw.Body.String(); got != want {
					t.Errorf("invalid jsonrpc response: \ngot: %v\nwant: %v\n", got, want)
				}
				wg.Done()
			})
		}(tc, tc.id)
	}
	wg.Wait()
}
