package source

import (
	"context"
	"net/url"
	"testing"

	"github.com/reactivex/rxgo/v2"
	. "github.com/smartystreets/goconvey/convey"
)

type dumbSource struct {
	uri *url.URL
}

func (s dumbSource) URI() *url.URL {
	return s.uri
}

func (s dumbSource) NewConsumer(_ context.Context, _ string, _ int64) rxgo.Observable {
	return nil
}

func TestNew(t *testing.T) {
	Convey("Considering a set of source factories with schemes 'foo' and 'bar'", t, func(c C) {
		RegisterFactory("foo", func(uri *url.URL) (Source, error) {
			return dumbSource{uri: uri}, nil
		})
		RegisterFactory("bar", func(uri *url.URL) (Source, error) {
			return dumbSource{uri: uri}, nil
		})
		So(len(sourceFactories), ShouldEqual, 2)

		Convey("When calling New() with scheme 'foo'", func(c C) {
			url, err := url.Parse("foo://xyz?t=0")
			So(err, ShouldBeNil)

			source, err := New(url)

			Convey("Then source is found", func(c C) {
				So(err, ShouldBeNil)
				So(source, ShouldNotBeNil)

				So(source.URI(), ShouldEqual, url)
			})
		})

		Convey("When calling New() with scheme 'barbar'", func(c C) {
			url, err := url.Parse("barbar://xyz?t=0")
			So(err, ShouldBeNil)

			source, err := New(url)

			Convey("Then source is not found", func(c C) {
				So(err, ShouldNotBeNil)
				So(source, ShouldBeNil)

				So(err, ShouldBeError, "scheme 'barbar' for url 'barbar://xyz?t=0' is not supported (available are: foo, bar): incorrect scheme")
			})
		})
	})
}
