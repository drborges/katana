package katana_test

import (
	"github.com/drborges/katana"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestStack(t *testing.T) {
	Convey("Given I have an instance of katana.Stack", t, func() {
		stack := katana.NewStack()

		Convey("Then I can push items into it", func() {
			stack.Push("1")

			So(stack.String(), should.Equal, "[1]")

			stack.Push("2").Push("3")

			So(stack.String(), should.Equal, "[1 2 3]")

			Convey("And I can reset the stack", func() {
				stack.Reset()
				So(stack.Empty(), should.BeTrue)
				So(stack.String(), should.Equal, "[]")
			})

			Convey("And I can check whether an item is in the stack", func() {
				So(stack.Contains("1"), should.BeTrue)
				So(stack.Contains("2"), should.BeTrue)
				So(stack.Contains("3"), should.BeTrue)

				Convey("And I can pop items from it", func() {
					item := stack.Pop()
					So(item, should.Equal, "3")
					So(stack.String(), should.Equal, "[1 2]")

					item = stack.Pop()
					So(item, should.Equal, "2")
					So(stack.String(), should.Equal, "[1]")

					item = stack.Pop()
					So(item, should.Equal, "1")
					So(stack.Empty(), should.BeTrue)
					So(stack.String(), should.Equal, "[]")
				})
			})
		})

		Convey("When I pop items from an empty stack", func() {
			item := stack.Pop()

			Convey("Then it returns an empty item", func() {
				So(item, should.BeEmpty)
			})
		})
	})
}
