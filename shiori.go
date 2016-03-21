package shiori

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Method is SHIORI Request Method
type Method int

const (
	// InvalidMethod is reserved value for error case
	InvalidMethod Method = iota
	// GET is GET SHIORI/x.x
	GET
	// NOTIFY is NOTIFY SHIORI/x.x
	NOTIFY
)

func (method Method) String() string {
	switch method {
	case GET:
		return "GET"
	case NOTIFY:
		return "NOTIFY"
	default:
		return ""
	}
}

// InvalidMethodError is invalid method error
type InvalidMethodError string

func (err InvalidMethodError) Error() string {
	return "InvalidMethodError: " + string(err)
}

// ToMethod converts method string into Method type
func ToMethod(method string) (Method, error) {
	switch method {
	case "GET":
		return GET, nil
	case "NOTIFY":
		return NOTIFY, nil
	default:
		return InvalidMethod, InvalidMethodError(method)
	}
}

// Protocol is SHIORI
type Protocol int

const (
	// InvalidProtocol is reserved value for error case
	InvalidProtocol Protocol = iota
	// SHIORI is SHIORI Protocol
	SHIORI
)

func (protocol Protocol) String() string {
	return "SHIORI"
}

// Request is SHIORI/x.x Request Message
type Request struct {
	Method   Method
	Protocol Protocol
	Version  string
	Headers  RequestHeaders
}

// Charset header
func (request *Request) Charset() string {
	return (*request).Headers["Charset"]
}

// Sender header
func (request *Request) Sender() string {
	return (*request).Headers["Sender"]
}

// Reference gets Reference* header
func (request *Request) Reference(i int) string {
	return (*request).Headers["Reference"+strconv.Itoa(i)]
}

func (request Request) String() string {
	return fmt.Sprintf("%s %s/%s\r\n%s\r\n", request.Method, request.Protocol, request.Version, request.Headers)
}

// Response is SHIORI/x.x Response Message
type Response struct {
	Code     int
	Protocol Protocol
	Version  string
	Headers  ResponseHeaders
}

// Message makes Response Message from Response Code
func (response *Response) Message() string {
	switch (*response).Code {
	case 200:
		return "OK"
	default:
		return ""
	}
}

// Charset header
func (response *Response) Charset() string {
	return (*response).Headers["Charset"]
}

// Sender header
func (response *Response) Sender() string {
	return (*response).Headers["Sender"]
}

// Value header
func (response *Response) Value(i int) string {
	return (*response).Headers["Value"]
}

// Reference gets Reference* header
func (response *Response) Reference(i int) string {
	return (*response).Headers["Reference"+strconv.Itoa(i)]
}

func (response Response) String() string {
	return fmt.Sprintf("%s/%s %d %s\r\n%s\r\n", response.Protocol, response.Version, response.Code, response.Message(), response.Headers)
}

// Headers is SHIORI Message Headers
type Headers map[string]string

// RequestHeaders is SHIORI Request Message Headers
type RequestHeaders Headers

// ResponseHeaders is SHIORI Response Message Headers
type ResponseHeaders Headers

func (headers Headers) String() string {
	headersStr := ""
	for key, value := range headers {
		headersStr += key + ": " + value + "\r\n"
	}
	return headersStr
}
func (headers RequestHeaders) String() string {
	return Headers(headers).String()
}
func (headers ResponseHeaders) String() string {
	return Headers(headers).String()
}

var requestLineRe = regexp.MustCompile(`^(.+) SHIORI/(\d+\.\d+)$`)

// ParseRequestError is Request parsing error
type ParseRequestError string

func (err ParseRequestError) Error() string {
	return "ParseRequestError: " + string(err)
}

// ParseRequest converts SHIORI/x.x Request Message into Request type
func ParseRequest(requestStr string) (Request, error) {
	request := Request{Protocol: SHIORI}
	lines := strings.Split(requestStr, "\r\n")
	requestLine := lines[0]
	headerLines := lines[1:]
	requestLineResult := requestLineRe.FindStringSubmatch(requestLine)
	if requestLineResult == nil {
		return request, ParseRequestError("request line parse failed: " + requestLine)
	}
	var err error
	request.Method, err = ToMethod(requestLineResult[1])
	if err != nil {
		return request, err
	}
	request.Version = requestLineResult[2]
	headers, err := ParseHeaderLines(headerLines)
	request.Headers = RequestHeaders(headers)
	if err != nil {
		return request, err
	}
	return request, nil
}

// ParseResponseError is Response parsing error
type ParseResponseError string

func (err ParseResponseError) Error() string {
	return "ParseResponseError: " + string(err)
}

var statusLineRe = regexp.MustCompile(`^SHIORI/(\d+\.\d+) (\d+) (.+)$`)

// ParseResponse converts SHIORI/x.x Response Message into Response type
func ParseResponse(responseStr string) (Response, error) {
	response := Response{Protocol: SHIORI}
	lines := strings.Split(responseStr, "\r\n")
	statusLine := lines[0]
	headerLines := lines[1:]
	statusLineResult := statusLineRe.FindStringSubmatch(statusLine)
	if statusLineResult == nil {
		return response, ParseResponseError("status line parse failed: " + statusLine)
	}
	response.Version = statusLineResult[1]
	var err error
	response.Code, err = strconv.Atoi(statusLineResult[2])
	if err != nil {
		return response, err
	}
	headers, err := ParseHeaderLines(headerLines)
	response.Headers = ResponseHeaders(headers)
	if err != nil {
		return response, err
	}
	return response, nil
}

// ParseHeaderError is Header parsing error
type ParseHeaderError string

func (err ParseHeaderError) Error() string {
	return "ParseHeaderError: " + string(err)
}

var headerRe = regexp.MustCompile(`^([^:]+): (.*)$`)

// ParseHeaderLines converts header lines into Headers type
func ParseHeaderLines(headerLines []string) (Headers, error) {
	headers := Headers{}
	for _, line := range headerLines {
		if line == "" {
			break
		}
		headerResult := headerRe.FindStringSubmatch(line)
		if headerResult == nil {
			return headers, ParseHeaderError("header line parse failed: " + line)
		}
		headers[headerResult[1]] = headerResult[2]
	}
	return headers, nil
}
