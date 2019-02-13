package e2e_test

import (
	"github.com/airbloc/airframe/afclient"
	"github.com/airbloc/airframe/test/utils"
	"github.com/ethereum/go-ethereum/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("afclient", func() {
	var endpoint string
	key, _ := crypto.GenerateKey()

	BeforeEach(func() {
		endpoint = testutils.LookupEnv("ENDPOINT", "localhost:9090")
	})

	It("should Dial", func() {
		_, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("should create object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		res, err := client.Put("testdata", "deadbeef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(res.Created).Should(BeTrue())
	})

	It("should get object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		_, err = client.Put("testdata", "b0b0beef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())

		obj, err := client.Get("/testdata/b0b0beef")
		Ω(obj.Data).Should(Equal(afclient.M{"foo": "bar"}))
		Ω(obj.Owner).Should(Equal(crypto.PubkeyToAddress(key.PublicKey)))
	})

	It("should update object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		resCreate, err := client.Put("testdata", "cafebabe", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(resCreate.Created).Should(BeTrue())

		resUpdate, err := client.Put("testdata", "cafebabe", afclient.M{"foo": "baz"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(resUpdate.Created).Should(BeFalse())

		obj, err := client.Get("/testdata/cafebabe")
		Ω(obj.Data).Should(Equal(afclient.M{"foo": "baz"}))
	})

	It("should not update object by others", func() {
		otherKey, _ := crypto.GenerateKey()

		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		client2, err := afclient.Dial(endpoint, otherKey)
		Ω(err).ShouldNot(HaveOccurred())

		_, err = client.Put("testdata", "00abcdef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())

		// others are trying to modify object
		_, err = client2.Put("testdata", "00abcdef", afclient.M{"foo": "baz"})
		Ω(err).Should(HaveOccurred())
	})
})
