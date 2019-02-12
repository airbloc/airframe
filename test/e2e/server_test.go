package e2e_test

import (
	"github.com/airbloc/service-template-go/test/utils"
	. "github.com/onsi/ginkgo"
	"gopkg.in/h2non/baloo.v3"
)

var _ = Describe("Server", func() {
	var server *baloo.Client

	BeforeEach(func() {
		endpoint := testutils.LookupEnv("ENDPOINT", "http://localhost:8080")
		server = baloo.New(endpoint)
	})

	It("should response with Hello World!", func() {
		server.
			Get("/").
			Expect(GinkgoT()).
			StatusOk().
			BodyEquals("Hello World!").
			Done()
	})
})
