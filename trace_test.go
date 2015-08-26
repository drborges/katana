package katana_test

import (
	"github.com/drborges/katana"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTrace(t *testing.T) {
	Convey("Given I have an instance of katana.Trace", t, func() {
		trace := katana.NewTrace()

		Convey("Then I can push items into it", func() {
			trace.Push("1")

			So(trace.String(), should.Equal, "[1]")

			trace.Push("2")
			trace.Push("3")

			So(trace.String(), should.Equal, "[1 -> 2 -> 3]")

			Convey("When I add an type that is already in the trace", func() {
				err := trace.Push("2")

				Convey("Then it returns a cyclic dependency error", func() {
					So(err.Error(), should.Resemble, katana.ErrCyclicDependency{&katana.Trace{
						Types: []string{"1", "2", "3", "2"},
					}}.Error())
				})
			})

			Convey("And I can check whether an item is already in the trace", func() {
				So(trace.Contains("1"), should.BeTrue)
				So(trace.Contains("2"), should.BeTrue)
				So(trace.Contains("3"), should.BeTrue)

				Convey("And I can pop items from it", func() {
					item := trace.Pop()
					So(item, should.Equal, "3")
					So(trace.String(), should.Equal, "[1 -> 2]")

					item = trace.Pop()
					So(item, should.Equal, "2")
					So(trace.String(), should.Equal, "[1]")

					item = trace.Pop()
					So(item, should.Equal, "1")
					So(trace.Empty(), should.BeTrue)
					So(trace.String(), should.Equal, "[]")
				})
			})
		})

		Convey("When I pop items from an empty trace", func() {
			item := trace.Pop()

			Convey("Then it returns an empty item", func() {
				So(item, should.BeEmpty)
			})
		})
	})
}
