package yamlpatch_test

import (
	yamlpatch "github.com/krishicks/yaml-patch"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Patch", func() {
	Describe("Apply", func() {
		DescribeTable(
			"positive cases",
			func(doc, ops, expectedYAML string) {
				patch, err := yamlpatch.DecodePatch([]byte(ops))
				Expect(err).NotTo(HaveOccurred())

				actualBytes, err := patch.Apply([]byte(doc))
				Expect(err).NotTo(HaveOccurred())

				var actualIface interface{}
				err = yaml.Unmarshal(actualBytes, &actualIface)
				Expect(err).NotTo(HaveOccurred())

				var expectedIface interface{}
				err = yaml.Unmarshal([]byte(expectedYAML), &expectedIface)
				Expect(err).NotTo(HaveOccurred())

				Expect(actualIface).To(Equal(expectedIface))
			},
			Entry("adding an element to an object",
				"---\nfoo: bar",
				"---\n- op: add\n  path: /baz\n  value: qux",
				"---\nfoo: bar\nbaz: qux",
			),
			Entry("adding an element to an array",
				"---\nfoo: [bar,baz]",
				"---\n- op: add\n  path: /foo/1\n  value: qux",
				"---\nfoo: [bar,qux,baz]",
			),
			Entry("removing an element from an object",
				"---\nfoo: bar\nbaz: qux",
				"---\n- op: remove\n  path: /baz",
				"---\nfoo: bar",
			),
			Entry("removing an element from an array",
				"---\nfoo: [bar,qux,baz]",
				"---\n- op: remove\n  path: /foo/1",
				"---\nfoo: [bar,baz]",
			),
			Entry("replacing an element in an object",
				"---\nfoo: bar\nbaz: qux",
				"---\n- op: replace\n  path: /baz\n  value: boo",
				"---\nfoo: bar\nbaz: boo",
			),
			Entry("moving an element in an object",
				"---\nfoo:\n  bar: baz\n  waldo: fred\nqux:\n  corge: grault",
				"---\n- op: move\n  from: /foo/waldo\n  path: /qux/thud",
				"---\nfoo:\n  bar: baz\nqux:\n  corge: grault\n  thud: fred",
			),
			Entry("moving an element in an array",
				"---\nfoo: [all, grass, cows, eat]",
				"---\n- op: move\n  from: /foo/1\n  path: /foo/3",
				"---\nfoo: [all, cows, eat, grass]",
			),
			Entry("adding an object to an object",
				"---\nfoo: bar",
				"---\n- op: add\n  path: /child\n  value:\n    grandchild: {}",
				"---\nfoo: bar\nchild:\n  grandchild: {}",
			),
			Entry("appending an element to an array",
				"---\nfoo: [bar]",
				"---\n- op: add\n  path: /foo/-\n  value: [abc,def]",
				"---\nfoo: [bar, [abc, def]]",
			),
			Entry("removing a nil element from an object",
				"---\nfoo: bar\nqux:\n  baz: 1\n  bar: ~",
				"---\n- op: remove\n  path: /qux/bar",
				"---\nfoo: bar\nqux:\n  baz: 1",
			),
			Entry("adding a nil element to an object",
				"---\nfoo: bar",
				"---\n- op: add\n  path: /baz\n  value: ~",
				"---\nfoo: bar\nbaz: ~",
			),
			Entry("replacing the sole element in an array",
				"---\nfoo: [bar]",
				"---\n- op: replace\n  path: /foo/0\n  value: baz",
				"---\nfoo: [baz]",
			),
			Entry("replacing an element in an array within a root array",
				"---\n- foo: [bar, qux, baz]",
				"---\n- op: replace\n  path: /0/foo/0\n  value: bum",
				"---\n- foo: [bum, qux, baz]",
			),
			Entry("copying an element in an array within a root array with an index",
				"---\n- foo: [bar, qux, baz]\n  bar: [qux, baz]",
				"---\n- op: copy\n  from: /0/foo/0\n  path: /0/bar/0",
				"---\n- foo: [bar, qux, baz]\n  bar: [bar, baz]",
			),
			XEntry("copying an element in an array within a root array to a destination without an index",
				// this is in jsonpatch, but I'd like confirmation from the spec that this is intended
				"---\n- foo: [bar, qux, baz]\n  bar: [qux, baz]",
				"---\n- op: copy\n  from: /0/foo/0\n  path: /0/bar",
				"---\n- foo: [bar, qux, baz]\n  bar: [bar, qux, baz]",
			),
			XEntry("copying an element in an array within a root array to a destination without an index",
				"---\n- foo:\n  bar: [qux, baz]\n  baz:\n    qux: bum",
				"---\n- op: copy\n  from: /0/foo/bar\n  path: /0/baz/bar",
				"---\n- foo: [bar, qux, baz]\n  bar: [bar, qux, baz]",
			),
		)

		DescribeTable(
			"failure cases",
			func(doc, ops string) {
				patch, err := yamlpatch.DecodePatch([]byte(ops))
				Expect(err).NotTo(HaveOccurred())

				_, err = patch.Apply([]byte(doc))
				Expect(err).To(HaveOccurred())
			},
			Entry("adding an element to an object with a bad pointer",
				"---\nfoo: bar",
				"---\n- op: add\n  path: /baz/bat\n  value: qux",
			),
			Entry("removing an element from an object with a bad pointer",
				"---\na:\n  b:\n    d: 1",
				"---\n- op: remove\n  path: /a/b/c",
			),
			Entry("moving an element in an object with a bad pointer",
				"---\na:\n  b:\n    d: 1",
				"---\n- op: move\n  from: /a/b/c\n  path: /a/b/e",
			),
			Entry("removing an element from an array with a bad pointer",
				"---\na:\n  b: [1]",
				"---\n- op: remove\n  path: /a/b/1",
			),
			Entry("moving an element from an array with a bad pointer",
				"---\na:\n  b: [1]",
				"---\n- op: move\n  from: /a/b/1\n  path: /a/b/2",
			),
			Entry("an operation with an invalid pathz field",
				"---\nfoo: bar",
				"---\n- op: add\n  pathz: /baz\n  value: qux",
			),
			Entry("an add operation with an empty path",
				"---\nfoo: bar",
				"---\n- op: add\n  path: ''\n  value: qux",
			),
			Entry("a replace operation on an array with an invalid path",
				"---\nname:\n  foo:\n    bat\n  qux:\n    bum",
				"---\n- op: replace\n  path: /foo/2\n  value: bum",
			),
		)
	})

	Describe("DecodePatch", func() {
		It("returns an empty patch when given nil", func() {
			patch, err := yamlpatch.DecodePatch(nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(patch).To(HaveLen(0))
		})

		It("returns a patch with a single op when given a single op", func() {
			ops := []byte(
				`---
- op: add
  path: /baz
  value: qux`)

			patch, err := yamlpatch.DecodePatch(ops)
			Expect(err).NotTo(HaveOccurred())

			var v interface{}
			v = "qux"
			Expect(patch).To(Equal(yamlpatch.Patch{
				{
					Op:       "add",
					Path:     "/baz",
					RawValue: &v,
				},
			}))
		})
	})
})
