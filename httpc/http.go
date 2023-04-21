package httpc

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/xpwu/go-log/log"
	"github.com/xpwu/go-reqid/reqid"
	"github.com/xpwu/go-xnet/xhttp"
	"io"
	"io/ioutil"
	"net/http"
)

type option struct {
	method     string
	body       io.Reader
	header     http.Header
	resF       func(response *http.Response) error
	resHeaderF func(header http.Header) error
}

type Option func(o *option) error

func WithHeader(header http.Header) Option {
	return func(o *option) error {
		o.header = header
		return nil
	}
}

func WithBody(body io.Reader) Option {
	return func(o *option) error {
		o.body = body
		o.method = http.MethodPost
		return nil
	}
}

func WithBytesBody(bytes_ []byte) Option {
	return func(o *option) error {
		o.body = bytes.NewReader(bytes_)
		o.method = http.MethodPost
		return nil
	}
}

func WithStructBodyToJson(struct_ interface{}) Option {
	return func(o *option) error {
		js, err := json.Marshal(struct_)
		if err != nil {
			return err
		}

		o.body = bytes.NewReader(js)
		o.method = http.MethodPost
		return nil
	}
}

func WithStructBodyToXml(struct_ interface{}) Option {
	return func(o *option) error {
		js, err := xml.Marshal(struct_)
		if err != nil {
			return err
		}

		o.body = bytes.NewReader(js)
		o.method = http.MethodPost
		return nil
	}
}

func WithResponse(res **http.Response) Option {
	return func(o *option) error {
		o.resF = func(response *http.Response) error {
			*res = response
			return nil
		}
		return nil
	}
}

func readResponse(resp *http.Response) (response []byte, err error) {

	if resp.StatusCode != http.StatusOK {
		err = errors.New(resp.Status)
		return
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	response, err = ioutil.ReadAll(resp.Body)

	return
}

func WithBytesResponse(res *[]byte) Option {
	return func(o *option) error {
		o.resF = func(response *http.Response) error {
			body, err := readResponse(response)
			if err != nil {
				return err
			}

			*res = body
			return nil
		}
		return nil
	}
}

func WithStructResponseFromJson(struct_ interface{}) Option {
	return func(o *option) error {
		o.resF = func(response *http.Response) error {
			body, err := readResponse(response)
			if err != nil {
				return err
			}

			err = json.Unmarshal(body, struct_)
			if err != nil {
				return err
			}

			return nil
		}
		return nil
	}
}

func WithStructResponseFromXml(struct_ interface{}) Option {
	return func(o *option) error {
		o.resF = func(response *http.Response) error {
			body, err := readResponse(response)
			if err != nil {
				return err
			}

			err = xml.Unmarshal(body, struct_)
			if err != nil {
				return err
			}

			return nil
		}
		return nil
	}
}

func WithMethod(method string) Option {
	return func(o *option) error {
		o.method = method
		return nil
	}
}

func WithResponseHeader(resHeader *http.Header) Option {
	return func(o *option) error {
		o.resHeaderF = func(header http.Header) error {
			*resHeader = header
			return nil
		}
		return nil
	}
}

func Send(ctx context.Context, url string, options ...Option) (err error) {

	opt := &option{
		header: make(http.Header),
		resF: func(response *http.Response) error {
			if response.StatusCode != http.StatusOK {
				return errors.New(response.Status)
			}
			return nil
		},
		method: http.MethodGet,
		resHeaderF: func(header http.Header) error {
			return nil
		},
	}
	for _, f := range options {
		if err := f(opt); err != nil {
			return err
		}
	}

	rid := opt.header.Get(reqid.HeaderKey)
	if rid == "" {
		// 新生成一个id，默认情况下不使用ctx里面的reqid，如果需要ctx里面的reqid，上层可以通过reqid.HeaderKey头信息设置
		rid = reqid.RandomID()
		opt.header.Set(reqid.HeaderKey, rid)
	}

	ctx, logger := log.WithCtx(ctx)

	logPrefix := fmt.Sprintf("send reqid:%s to url(%s). ", rid, url)
	logger.PushPrefix(logPrefix)
	defer logger.PopPrefix()

	logger.Debug("Start")

	request, err := http.NewRequestWithContext(ctx, opt.method, url, opt.body)
	if err != nil {
		logger.Error(err)
		return err
	}

	request.Header = opt.header

	response, err := xhttp.DefaultClient.Do(request)
	if err != nil {
		logger.Error(err)
		return err
	}

	// 需要先读body，没有错误的情况再读header
	err = opt.resF(response)
	if err != nil {
		return err
	}

	err = opt.resHeaderF(response.Header)
	if err != nil {
		return err
	}

	logger.Info("End")

	return nil
}
