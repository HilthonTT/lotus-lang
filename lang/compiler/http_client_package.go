package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hilthontt/lotus/object"
)

var defaultHttpClient = &http.Client{Timeout: 30 * time.Second}

func buildResponse(resp *http.Response) object.Object {
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	ok := resp.StatusCode >= 200 && resp.StatusCode < 300

	// Build headers sub-map
	headerPairs := make(map[object.HashKey]object.HashPair)
	for hk, hv := range resp.Header {
		key := &object.String{Value: strings.ToLower(hk)}
		val := &object.String{Value: strings.Join(hv, ", ")}
		headerPairs[key.HashKey()] = object.HashPair{Key: key, Value: val}
	}

	pairs := make(map[object.HashKey]object.HashPair)
	setStr := func(k, v string) {
		key := &object.String{Value: k}
		val := &object.String{Value: v}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: val}
	}
	setObj := func(k string, v object.Object) {
		key := &object.String{Value: k}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: v}
	}

	setObj("status", &object.Integer{Value: int64(resp.StatusCode)})
	setObj("ok", &object.Boolean{Value: ok})
	setStr("body", string(bodyBytes))
	setStr("url", resp.Request.URL.String())
	setObj("headers", &object.Hash{Pairs: headerPairs})
	setStr("error", "")

	return &object.Hash{Pairs: pairs}
}

func buildErrorResponse(err error) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	setObj := func(k string, v object.Object) {
		key := &object.String{Value: k}
		pairs[key.HashKey()] = object.HashPair{Key: key, Value: v}
	}
	setObj("status", &object.Integer{Value: 0})
	setObj("ok", &object.Boolean{Value: false})
	setObj("body", &object.String{Value: ""})
	setObj("headers", &object.Hash{Pairs: map[object.HashKey]object.HashPair{}})
	setObj("url", &object.String{Value: ""})
	setObj("error", &object.String{Value: err.Error()})
	return &object.Hash{Pairs: pairs}
}

func toStringHeaders(obj object.Object) map[string]string {
	result := map[string]string{}
	h, ok := obj.(*object.Hash)
	if !ok {
		return result
	}
	for _, pair := range h.Pairs {
		k, ok1 := pair.Key.(*object.String)
		v, ok2 := pair.Value.(*object.String)
		if ok1 && ok2 {
			result[k.Value] = v.Value
		}
	}
	return result
}

// toRequestBody converts a Lotus value to an io.Reader + content-type.
// Hashes and Arrays are auto-serialized to JSON.
// Strings are sent as-is with text/plain.
func toRequestBody(obj object.Object) (io.Reader, string) {
	switch v := obj.(type) {
	case *object.String:
		return strings.NewReader(v.Value), "text/plain; charset=utf-8"
	case *object.Hash:
		b, _ := json.Marshal(lotusToGoValue(v))
		return bytes.NewReader(b), "application/json"
	case *object.Array:
		b, _ := json.Marshal(lotusToGoValue(v))
		return bytes.NewReader(b), "application/json"
	default:
		return strings.NewReader(obj.Inspect()), "text/plain; charset=utf-8"
	}
}

func doRequest(method, urlStr string, body io.Reader, contentType string, headers object.Object) object.Object {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return buildErrorResponse(err)
	}
	if body != nil && contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if headers != nil {
		for k, v := range toStringHeaders(headers) {
			req.Header.Set(k, v)
		}
	}
	resp, err := defaultHttpClient.Do(req)
	if err != nil {
		return buildErrorResponse(err)
	}
	return buildResponse(resp)
}

func lotusToGoValue(obj object.Object) any {
	switch v := obj.(type) {
	case *object.Integer:
		return v.Value
	case *object.Float:
		return v.Value
	case *object.String:
		return v.Value
	case *object.Boolean:
		return v.Value
	case *object.Nil:
		return nil
	case *object.Hash:
		m := map[string]any{}
		for _, pair := range v.Pairs {
			k := pair.Key.Inspect()
			if s, ok := pair.Key.(*object.String); ok {
				k = s.Value
			}
			m[k] = lotusToGoValue(pair.Value)
		}
		return m
	case *object.Array:
		s := make([]any, len(v.Elements))
		for i, el := range v.Elements {
			s[i] = lotusToGoValue(el)
		}
		return s
	default:
		return v.Inspect()
	}
}

func goValueToLotus(v any) object.Object {
	if v == nil {
		return &object.Nil{}
	}
	switch val := v.(type) {
	case bool:
		return &object.Boolean{Value: val}
	case float64:
		if val == float64(int64(val)) {
			return &object.Integer{Value: int64(val)}
		}
		return &object.Float{Value: val}
	case string:
		return &object.String{Value: val}
	case []any:
		elems := make([]object.Object, len(val))
		for i, el := range val {
			elems[i] = goValueToLotus(el)
		}
		return &object.Array{Elements: elems}
	case map[string]any:
		pairs := make(map[object.HashKey]object.HashPair)
		for k, v2 := range val {
			key := &object.String{Value: k}
			pairs[key.HashKey()] = object.HashPair{Key: key, Value: goValueToLotus(v2)}
		}
		return &object.Hash{Pairs: pairs}
	default:
		return &object.String{Value: fmt.Sprintf("%v", val)}
	}
}

func httpClientPackage() *object.Package {
	return &object.Package{
		Name: "HttpClient",
		Functions: map[string]object.PackageFunction{

			// HttpClient.get(url) -> Response
			// HttpClient.get(url, headers) -> Response
			"get": func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return buildErrorResponse(fmt.Errorf("get: url required"))
				}
				url, ok := args[0].(*object.String)
				if !ok {
					return buildErrorResponse(fmt.Errorf("get: url must be a string"))
				}
				var headers object.Object
				if len(args) >= 2 {
					headers = args[1]
				}
				return doRequest(http.MethodGet, url.Value, nil, "", headers)
			},

			// HttpClient.post(url, body) -> Response
			// HttpClient.post(url, body, headers) -> Response
			"post": func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return buildErrorResponse(fmt.Errorf("post: url and body required"))
				}
				url, ok := args[0].(*object.String)
				if !ok {
					return buildErrorResponse(fmt.Errorf("post: url must be a string"))
				}
				body, ct := toRequestBody(args[1])
				var headers object.Object
				if len(args) >= 3 {
					headers = args[2]
				}
				return doRequest(http.MethodPost, url.Value, body, ct, headers)
			},

			// HttpClient.put(url, body) -> Response
			// HttpClient.put(url, body, headers) -> Response
			"put": func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return buildErrorResponse(fmt.Errorf("put: url and body required"))
				}
				url, ok := args[0].(*object.String)
				if !ok {
					return buildErrorResponse(fmt.Errorf("put: url must be a string"))
				}
				body, ct := toRequestBody(args[1])
				var headers object.Object
				if len(args) >= 3 {
					headers = args[2]
				}
				return doRequest(http.MethodPut, url.Value, body, ct, headers)
			},

			// HttpClient.patch(url, body) -> Response
			// HttpClient.patch(url, body, headers) -> Response
			"patch": func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return buildErrorResponse(fmt.Errorf("patch: url and body required"))
				}
				url, ok := args[0].(*object.String)
				if !ok {
					return buildErrorResponse(fmt.Errorf("patch: url must be a string"))
				}
				body, ct := toRequestBody(args[1])
				var headers object.Object
				if len(args) >= 3 {
					headers = args[2]
				}
				return doRequest(http.MethodPatch, url.Value, body, ct, headers)
			},

			// HttpClient.delete(url) -> Response
			// HttpClient.delete(url, headers) -> Response
			"delete": func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return buildErrorResponse(fmt.Errorf("delete: url required"))
				}
				url, ok := args[0].(*object.String)
				if !ok {
					return buildErrorResponse(fmt.Errorf("delete: url must be a string"))
				}
				var headers object.Object
				if len(args) >= 2 {
					headers = args[1]
				}
				return doRequest(http.MethodDelete, url.Value, nil, "", headers)
			},

			// HttpClient.request(method, url) -> Response
			// HttpClient.request(method, url, body) -> Response
			// HttpClient.request(method, url, body, headers) -> Response
			"request": func(args ...object.Object) object.Object {
				if len(args) < 2 {
					return buildErrorResponse(fmt.Errorf("request: method and url required"))
				}
				method, ok1 := args[0].(*object.String)
				url, ok2 := args[1].(*object.String)
				if !ok1 || !ok2 {
					return buildErrorResponse(fmt.Errorf("request: method and url must be strings"))
				}
				var body io.Reader
				var ct string
				if len(args) >= 3 && args[2].Type() != object.NIL_OBJ {
					body, ct = toRequestBody(args[2])
				}
				var headers object.Object
				if len(args) >= 4 {
					headers = args[3]
				}
				return doRequest(strings.ToUpper(method.Value), url.Value, body, ct, headers)
			},

			// HttpClient.json(response) -> value
			// Parses response["body"] as JSON. Also accepts a plain string.
			"json": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				var bodyStr string
				switch v := args[0].(type) {
				case *object.Hash:
					// Extract body field from response hash
					key := &object.String{Value: "body"}
					if pair, ok := v.Pairs[key.HashKey()]; ok {
						if s, ok := pair.Value.(*object.String); ok {
							bodyStr = s.Value
						}
					}
				case *object.String:
					bodyStr = v.Value
				default:
					return &object.Nil{}
				}
				bodyStr = strings.TrimSpace(bodyStr)
				if bodyStr == "" {
					return &object.Nil{}
				}
				var raw any
				if err := json.Unmarshal([]byte(bodyStr), &raw); err != nil {
					return &object.Nil{}
				}
				return goValueToLotus(raw)
			},

			// HttpClient.setTimeout(ms) -> nil
			// Sets the global timeout for all subsequent requests.
			"setTimeout": func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return &object.Nil{}
				}
				ms, ok := args[0].(*object.Integer)
				if !ok {
					return &object.Nil{}
				}
				defaultHttpClient.Timeout = time.Duration(ms.Value) * time.Millisecond
				return &object.Nil{}
			},

			// HttpClient.buildUrl(base, params) -> string
			// Builds a URL with query parameters from a map.
			// HttpClient.buildUrl("https://api.example.com/search", {"q": "lotus", "limit": "10"})
			// → "https://api.example.com/search?limit=10&q=lotus"
			"buildUrl": func(args ...object.Object) object.Object {
				if len(args) < 1 {
					return &object.String{Value: ""}
				}
				base, ok := args[0].(*object.String)
				if !ok {
					return &object.String{Value: ""}
				}
				if len(args) < 2 {
					return base
				}
				params := toStringHeaders(args[1])
				if len(params) == 0 {
					return base
				}
				var parts []string
				for k, v := range params {
					parts = append(parts, k+"="+v)
				}
				sep := "?"
				if strings.Contains(base.Value, "?") {
					sep = "&"
				}
				return &object.String{Value: base.Value + sep + strings.Join(parts, "&")}
			},
		},
	}
}
