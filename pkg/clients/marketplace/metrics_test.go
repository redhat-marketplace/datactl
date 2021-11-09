package marketplace

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("marketplace uploaders", func() {
	var (
		err    error
		server *ghttp.Server
		sut    Client

		postReponse               MarketplaceUsageResponse
		getResponse, getResponse2 MarketplaceUsageResponse
		retryPostResponse         MarketplaceUsageResponse

		fileName = "test"

		testId = "testId"

		testBody = []byte("foo")

		config *MarketplaceConfig
	)

	BeforeEach(func() {
		server = ghttp.NewTLSServer()
		caCertPool, _ := x509.SystemCertPool()
		caCertPool.AddCert(server.HTTPTestServer.Certificate())

		config = &MarketplaceConfig{
			URL:   server.URL(),
			Token: "foo",
			TlsConfig: &tls.Config{
				RootCAs: caCertPool,
			},
			polling: 1 * time.Second,
			timeout: 4 * time.Second,
		}
		sut = NewClient(config)

		postReponse = MarketplaceUsageResponse{RequestID: testId}
		retryPostResponse = MarketplaceUsageResponse{
			Details: &MarketplaceUsageResponseDetails{
				Code:      "409",
				Retryable: true,
			},
		}
		getResponse = MarketplaceUsageResponse{
			Status: MktplStatusInProgress,
		}
		getResponse2 = MarketplaceUsageResponse{
			Status: MktplStatusSuccess,
		}
		testBody = []byte("foo")
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("uploading files", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					verifyFileUpload(fileName, testBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &postReponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse2),
				),
			)
		})

		It("should upload a file", func() {
			ctx := context.Background()
			id, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(testId))
		})
	})

	Describe("uploading files", func() {
		BeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					ghttp.RespondWithJSONEncoded(http.StatusTooManyRequests, &retryPostResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					verifyFileUpload(fileName, testBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &postReponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusTooManyRequests, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse2),
				),
			)
		})

		It("should upload a file and retry", func() {
			ctx := context.Background()
			id, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(testId))
		})
	})

	Describe("uploading files", func() {

		BeforeEach(func() {
			getResponse = MarketplaceUsageResponse{
				Status:    MktplStatusFailed,
				Message:   "failed",
				ErrorCode: "100",
			}

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					verifyFileUpload(fileName, testBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &postReponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
			)
		})

		It("should return error on failure", func() {
			ctx := context.Background()
			_, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("uploading files", func() {
		BeforeEach(func() {
			config.polling = 1
			config.timeout = 1
			sut = NewClient(config)
			Expect(err).To(Succeed())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					verifyFileUpload(fileName, testBody),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &postReponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/usage/api/v2/metrics/"+testId),
					ghttp.RespondWithJSONEncoded(http.StatusOK, &getResponse),
				),
			)
		})

		It("should timeout", func() {
			ctx := context.Background()
			_, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("handling conflict", func() {
		BeforeEach(func() {
			sut = NewClient(config)
			Expect(err).To(Succeed())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
					verifyFileUpload(fileName, testBody),
					ghttp.RespondWith(http.StatusConflict, `{"errorCode":"document_conflict","message":"Upload is duplicate of previous submission","details":{"code":"document_conflict","statusCode":409,"retryable":false}}`),
				),
			)
		})

		It("should handle duplicate conflict", func() {
			ctx := context.Background()
			id, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(BeEmpty())
		})
	})

	Describe("handling error", func() {
		BeforeEach(func() {
			sut = NewClient(config)
			Expect(err).To(Succeed())

			for i := 0; i < 4; i++ {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/usage/api/v2/metrics"),
						ghttp.RespondWith(http.StatusInternalServerError, `{"errorCode":"internal_application_error_ocurred","message":"Save usage result not ok or missing value, result {\"lastErrorObject\":{\"n\":0,\"updatedExisting\":false},\"value\":null,\"ok\":1,\"$clusterTime\":{\"clusterTime\":{\"$timestamp\":\"7025668740716953620\"},\"signature\":{\"hash\":\"H2Zbl5S64rst/CWWEsRupwqyUZs=\",\"keyId\":{\"low\":2,\"high\":1628832003,\"unsigned\":false}}},\"operationTime\":{\"$timestamp\":\"7025668740716953620\"}}, Retry UsageStatus.save failed after retry attempts: 3 duration: 3069 ms","details":{"code":"internal_application_error_ocurred","statusCode":500,"retryable":true}}`),
					),
				)
			}
		})

		It("should handle error", func() {
			ctx := context.Background()
			_, err := sut.Metrics().Upload(ctx, fileName, bytes.NewReader(testBody))
			Expect(err).To(HaveOccurred())
		})
	})
})

func verifyFileUpload(fileName string, testBody []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		Expect(req.Header.Get("Content-Type")).To(ContainSubstring("multipart/form-data"))

		err := req.ParseMultipartForm(32 << 20) // maxMemory 32 MB
		Expect(err).To(Succeed())

		file, _, err := req.FormFile(fileName)
		Expect(err).To(Succeed())

		buff := &bytes.Buffer{}
		io.Copy(buff, file)

		Expect(buff.Bytes()).To(Equal(testBody))
		Expect(file.Close()).To(Succeed())
	}
}
