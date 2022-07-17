package httpcache

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/hinha/httpcache/helper"
)

var (
	HeaderAuthorization = "Authorization"
	HeaderCacheControl  = "Cache-Control"

	XFromHache   = "X-HTTPCache"
	XHacheOrigin = "X-HTTPCache-Origin"
)

type CacheHandler struct {
	defaultRoundTripper http.RoundTripper
	iCache              CacheInterface
}

// NewCacheHandlerRoundtrip will create an implementations of cache http roundtripper
func NewCacheHandlerRoundtrip(roundTripper http.RoundTripper, ICache CacheInterface) *CacheHandler {
	if ICache == nil {
		panic("cache storage is not well set")
	}
	return &CacheHandler{
		defaultRoundTripper: roundTripper,
		iCache:              ICache,
	}
}

// RoundTrip the implementation of http.RoundTripper
func (c *CacheHandler) RoundTrip(request *http.Request) (*http.Response, error) {
	cachedResp, cachedItem, cachedErr := getCachedResponse(c.iCache, request)
	if cachedResp != nil && cachedErr == nil {
		buildTheCachedResponseHeader(cachedResp, cachedItem, "CACHE")
		return cachedResp, cachedErr
	}

	resp, err := c.defaultRoundTripper.RoundTrip(request)
	if err != nil {
		return nil, err
	}

	if err := storeRespToCache(c.iCache, request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func storeRespToCache(Icache CacheInterface, request *http.Request, response *http.Response) error {
	cachedResp := CachedResponse{
		RequestMethod: request.Method,
		RequestURI:    getUrl(request),
		CachedTime:    time.Now(),
	}

	dumpedResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		return err
	}
	cachedResp.Response = dumpedResponse

	return Icache.Set(getCacheKey(request), cachedResp)
}

func getCachedResponse(iCache CacheInterface, req *http.Request) (*http.Response, CachedResponse, error) {
	cachedResp, err := iCache.Get(getCacheKey(req))
	if err != nil {
		return nil, CachedResponse{}, err
	}

	cachedResponse := bytes.NewBuffer(cachedResp.Response)
	resp, err := http.ReadResponse(bufio.NewReader(cachedResponse), req)
	if err != nil {
		return nil, cachedResp, err
	}

	validationResult, err := validateTheCacheControl(req, resp)
	if err != nil {
		return nil, cachedResp, err
	}

	if validationResult.OutErr != nil {
		return nil, cachedResp, validationResult.OutErr
	}

	if time.Now().After(validationResult.OutExpirationTime) {
		return nil, cachedResp, fmt.Errorf("cached-item already expired")
	}

	return resp, cachedResp, err
}

func getCacheKey(req *http.Request) (key string) {
	key = fmt.Sprintf("%s %s", req.Method, getUrl(req))
	if req.Header.Get(HeaderAuthorization) != "" {
		// need token bearer
		key = fmt.Sprintf("%s %s", key, strings.TrimSpace(req.Header.Get(HeaderAuthorization)))
	}
	return
}

func validateTheCacheControl(req *http.Request, resp *http.Response) (helper.ObjectResults, error) {
	reqDir, err := helper.ParseRequestCacheControl(req.Header.Get(HeaderCacheControl))
	if err != nil {
		return helper.ObjectResults{}, err
	}

	resDir, err := helper.ParseResponseCacheControl(req.Header.Get(HeaderCacheControl))
	if err != nil {
		return helper.ObjectResults{}, err
	}

	expiry := resp.Header.Get("Expires")
	expiresHeader, err := http.ParseTime(expiry)
	if err != nil && expiry != "" &&
		// https://stackoverflow.com/questions/11357430/http-expires-header-values-0-and-1
		expiry != "-1" && expiry != "0" {
		return helper.ObjectResults{}, err
	}

	dateHeaderStr := resp.Header.Get("Date")
	dateHeader, err := http.ParseTime(dateHeaderStr)
	if err != nil && dateHeaderStr != "" {
		return helper.ObjectResults{}, err
	}

	lastModifiedStr := resp.Header.Get("Last-Modified")
	lastModifiedHeader, err := http.ParseTime(lastModifiedStr)
	if err != nil && lastModifiedStr != "" {
		return helper.ObjectResults{}, err
	}

	obj := helper.Object{
		RespDirectives:         resDir,
		RespHeaders:            resp.Header,
		RespStatusCode:         resp.StatusCode,
		RespExpiresHeader:      expiresHeader,
		RespDateHeader:         dateHeader,
		RespLastModifiedHeader: lastModifiedHeader,
		ReqDirectives:          reqDir,
		ReqHeaders:             req.Header,
		ReqMethod:              req.Method,
		NowUTC:                 time.Now().UTC(),
	}

	validationResult := helper.ObjectResults{}
	helper.CachableObject(&obj, &validationResult)
	helper.ExpirationObject(&obj, &validationResult)

	return validationResult, nil
}

func buildTheCachedResponseHeader(resp *http.Response, cachedResp CachedResponse, origin string) {
	resp.Header.Add("Expires", cachedResp.CachedTime.String())
	resp.Header.Add(XFromHache, "true")
	resp.Header.Add(XHacheOrigin, origin)
}

func getUrl(req *http.Request) string {
	var uri string
	if req.RequestURI != "" {
		uri = req.RequestURI
	} else {
		uri = req.URL.String()
	}
	parse, err := url.QueryUnescape(uri)
	if err == nil {
		uri = parse
	}
	return uri
}
