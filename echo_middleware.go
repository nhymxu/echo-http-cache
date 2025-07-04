package cache

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func (c *Client) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(echoCtx echo.Context) error {
			var err error
			req := echoCtx.Request()
			res := echoCtx.Response()

			if !c.cacheableMethod(req.Method) {
				if err = next(echoCtx); err != nil {
					echoCtx.Error(err)
				}
				return err
			}

			// handle cache
			sortURLParams(req.URL)
			key := generateKey(req.URL.String())

			if req.Method == http.MethodPost && req.Body != nil {
				body, err := io.ReadAll(req.Body)
				defer req.Body.Close()

				if err != nil {
					err := next(echoCtx)
					if err != nil {
						echoCtx.Error(err)
					}
					return err
				}
				reader := io.NopCloser(bytes.NewBuffer(body))
				key = generateKeyWithBody(req.URL.String(), body)
				req.Body = reader
			}

			params := req.URL.Query()
			if _, ok := params[c.refreshKey]; ok {
				delete(params, c.refreshKey)

				req.URL.RawQuery = params.Encode()
				key = generateKey(req.URL.String())

				c.adapter.Release(key)
			} else {
				b, ok := c.adapter.Get(key)
				response := BytesToResponse(b)
				if ok {
					if response.Expiration.After(time.Now()) {
						response.LastAccess = time.Now()
						response.Frequency++
						c.adapter.Set(key, response.Bytes(), response.Expiration)

						//w.WriteHeader(http.StatusNotModified)
						for k, v := range response.Header {
							res.Header().Set(k, strings.Join(v, ","))
						}

						if c.writeExpiresHeader {
							res.Header().Set("Expires", response.Expiration.UTC().Format(http.TimeFormat))
						}

						// write a custom header X-Cache: HIT
						res.Header().Set("X-Cache", "HIT")
						res.WriteHeader(http.StatusOK)
						res.Write(response.Value)

						return nil
					}

					c.adapter.Release(key)
				}
			}

			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(res.Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: res.Writer}
			res.Writer = writer

			if err := next(echoCtx); err != nil {
				echoCtx.Error(err)
			}

			statusCode := writer.statusCode
			value := resBody.Bytes()
			if statusCode < 400 {
				now := time.Now()

				response := Response{
					Value:      value,
					Header:     writer.Header(),
					Expiration: now.Add(c.ttl),
					LastAccess: now,
					Frequency:  1,
				}
				c.adapter.Set(key, response.Bytes(), response.Expiration)
			}
			//for k, v := range writer.Header() {
			//	echoCtx.Response().Header().Set(k, strings.Join(v, ","))
			//}
			//echoCtx.Response().WriteHeader(statusCode)
			//echoCtx.Response().Write(value)
			// end handle cache

			if err = next(echoCtx); err != nil {
				echoCtx.Error(err)
			}

			return nil
		}
	}
}
