package api

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("source", func() {
	It("should parse correct", func() {
		text, err := DataService.MarshalText()
		Expect(err).To(Succeed())
		Expect(text).To(Equal([]byte(DataService.String())))
		var obj SourceType
		err = (&obj).UnmarshalText(text)
		Expect(err).To(Succeed())
		Expect(obj).To(Equal(DataService))
	})
})
