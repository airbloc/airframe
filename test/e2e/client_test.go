package e2e_test

import (
	"context"
	"github.com/airbloc/airframe/afclient"
	"github.com/airbloc/airframe/test/utils"
	"github.com/klaytn/klaytn/crypto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("afclient", func() {
	var endpoint string
	var ctx context.Context
	var cancel context.CancelFunc
	key, _ := crypto.GenerateKey()

	BeforeEach(func() {
		endpoint = testutils.LookupEnv("ENDPOINT", "localhost:9090")
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		// prevent this: the cancel function is not used on all paths (possible context leak)
		_ = cancel
	})
	AfterEach(func() {
		cancel()
	})

	It("should Dial", func() {
		_, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("should create object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		res, err := client.Put(ctx, "testdata", "deadbeef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(res.Created).Should(BeTrue())
	})

	It("should get object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		_, err = client.Put(ctx, "testdata", "b0b0beef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())

		obj, err := client.Get(ctx, "testdata", "b0b0beef")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(obj.Data).Should(Equal(afclient.M{"foo": "bar"}))
		Ω(obj.Owner).Should(Equal(crypto.PubkeyToAddress(key.PublicKey)))
	})

	It("should update object", func() {
		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		resCreate, err := client.Put(ctx, "testdata", "cafebabe", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(resCreate.Created).Should(BeTrue())

		resUpdate, err := client.Put(ctx, "testdata", "cafebabe", afclient.M{"foo": "baz"})
		Ω(err).ShouldNot(HaveOccurred())
		Ω(resUpdate.Created).Should(BeFalse())

		obj, err := client.Get(ctx, "testdata", "cafebabe")
		Ω(err).ShouldNot(HaveOccurred())
		Ω(obj.Data).Should(Equal(afclient.M{"foo": "baz"}))
	})

	It("should not update object by others", func() {
		otherKey, _ := crypto.GenerateKey()

		client, err := afclient.Dial(endpoint, key)
		Ω(err).ShouldNot(HaveOccurred())

		client2, err := afclient.Dial(endpoint, otherKey)
		Ω(err).ShouldNot(HaveOccurred())

		_, err = client.Put(ctx, "testdata", "00abcdef", afclient.M{"foo": "bar"})
		Ω(err).ShouldNot(HaveOccurred())

		// others are trying to modify object
		_, err = client2.Put(ctx, "testdata", "00abcdef", afclient.M{"foo": "baz"})
		Ω(err).Should(Equal(afclient.ErrNotAuthorized))
	})
})
