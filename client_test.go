package czds_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	czds "github.com/martinsirbe/go-icann-czds-client"
)

const (
	testEmail     = "test-email"
	testPassword  = "test-password"
	testGoodToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ik1hcnRpbnMgSXJiZSIsImlhdCI6MTUxNjIzOTAyMiwiZXhwIjo4ODg4ODg4ODg4OH0.NPp6gHGl-DFrD6Bk5VGd2VcTFCcKztecm4d3U2AR_yk"
)

func TestGetZoneFile(t *testing.T) {
	for name, tc := range map[string]struct {
		setupICANNAccountsAPIMock func() *httptest.Server
		setupCZDSAPIMock          func() *httptest.Server
		expectedZoneFileDetails   map[string][]string
		errAssert                 assert.ErrorAssertionFunc
	}{
		"Success": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/downloads/com.zone", r.URL.Path)

					w.WriteHeader(http.StatusOK)

					_, err := w.Write([]byte(`test-1.com.	10800	in	ns	test-dns-1.com.
test-1.com.	10800	in	ns	test-dns-2.com.
test-2.com.	10800	in	ns	test-dns-3.com.
test-3.com.	10800	in	ns	test-dns-4.com.`))
					require.NoError(t, err)
				}))
				return ts
			},
			expectedZoneFileDetails: map[string][]string{
				"test-1.com.": {
					"10800,in,ns,test-dns-1.com.",
					"10800,in,ns,test-dns-2.com.",
				},
				"test-2.com.": {
					"10800,in,ns,test-dns-3.com.",
				},
				"test-3.com.": {
					"10800,in,ns,test-dns-4.com.",
				},
			},
			errAssert: assert.NoError,
		},
		"Success_ZoneFileAsGzip": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					var reqBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/downloads/com.zone", r.URL.Path)

					w.Header().Set("Content-Type", "application/x-gzip")

					var buffer bytes.Buffer
					gz := gzip.NewWriter(&buffer)
					_, err := gz.Write([]byte(`test-1.com.	10800	in	ns	test-dns-1.com.
test-1.com.	10800	in	ns	test-dns-2.com.
test-2.com.	10800	in	ns	test-dns-3.com.
test-3.com.	10800	in	ns	test-dns-4.com.`))
					require.NoError(t, err)
					require.NoError(t, gz.Close())

					_, err = w.Write(buffer.Bytes())
					require.NoError(t, err)
				}))
				return ts
			},
			expectedZoneFileDetails: map[string][]string{
				"test-1.com.": {
					"10800,in,ns,test-dns-1.com.",
					"10800,in,ns,test-dns-2.com.",
				},
				"test-2.com.": {
					"10800,in,ns,test-dns-3.com.",
				},
				"test-3.com.": {
					"10800,in,ns,test-dns-4.com.",
				},
			},
			errAssert: assert.NoError,
		},
		"Fail_FetchJWTReturnsHTTP500": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					w.WriteHeader(http.StatusInternalServerError)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fail()
				}))
				return ts
			},
			errAssert: assert.Error,
		},
		"Fail_FetchJWTReturnsBadResponse": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					_, err := w.Write([]byte("not a json response"))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fail()
				}))
				return ts
			},
			errAssert: assert.Error,
		},
		"Fail_ExpiredToken": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					const expiredToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ik1hcnRpbnMgSXJiZSIsImlhdCI6MTUxNjIzOTAyMiwiZXhwIjo4ODg4ODg4OH0.f1MBGBBvza_-DLyoXv_oujVZfQWoOEFyC4-I0_0MZQQ"
					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, expiredToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Fail()
				}))
				return ts
			},
			errAssert: assert.Error,
		},
		"Fail_APIReturnsHTTP500": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				return ts
			},
			errAssert: assert.Error,
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockAccountsAPI := tc.setupICANNAccountsAPIMock()
			defer mockAccountsAPI.Close()

			mockCZDSAPI := tc.setupCZDSAPIMock()
			defer mockCZDSAPI.Close()

			client := czds.NewClient(testEmail, testPassword,
				czds.ICANNAccountsAPIBaseURL(mockAccountsAPI.URL),
				czds.CZDSAPIBaseURL(mockCZDSAPI.URL))

			records, err := client.GetZoneFile(context.Background(), "com")
			tc.errAssert(t, err)
			if err != nil {
				return
			}

			assert.Len(t, records, len(tc.expectedZoneFileDetails))

			for expectedDomainName, expectedRecords := range tc.expectedZoneFileDetails {
				assert.ElementsMatch(t, expectedRecords, records[expectedDomainName])
			}
		})
	}
}

func TestListTLDs(t *testing.T) {
	for name, tc := range map[string]struct {
		setupICANNAccountsAPIMock func() *httptest.Server
		setupCZDSAPIMock          func() *httptest.Server
		expectedTLDs              []czds.TLD
		errAssert                 assert.ErrorAssertionFunc
	}{
		"Success": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/tlds", r.URL.Path)

					_, err := w.Write([]byte(`[{"tld":"dev","ulable":"dev","currentStatus":"approved","sftp":false},
{"tld":"tech","ulable":"tech","currentStatus":"approved","sftp":false},
{"tld":"com","ulable":"com","currentStatus":"pending","sftp":false}]`))
					require.NoError(t, err)
				}))
				return ts
			},
			expectedTLDs: []czds.TLD{
				{Name: "dev", Ulable: "dev", CurrentStatus: "approved", SFTP: false},
				{Name: "tech", Ulable: "tech", CurrentStatus: "approved", SFTP: false},
				{Name: "com", Ulable: "com", CurrentStatus: "pending", SFTP: false},
			},
			errAssert: assert.NoError,
		},
		"Fail_APIReturnsHTTP500": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/tlds", r.URL.Path)

					w.WriteHeader(http.StatusInternalServerError)
				}))
				return ts
			},
			errAssert: assert.Error,
		},
		"Fail_BadAPIResponse": {
			setupICANNAccountsAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodPost, r.Method)
					require.Equal(t, "/authenticate", r.URL.Path)

					type authRequestBody struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}

					var reqBody authRequestBody
					require.NoError(t, json.NewDecoder(r.Body).Decode(&reqBody))
					require.Equal(t, reqBody.Username, testEmail)
					require.Equal(t, reqBody.Password, testPassword)

					testResponse := fmt.Sprintf(`{"accessToken":%q,"message":"Authentication Successful"}`, testGoodToken)
					_, err := w.Write([]byte(testResponse))
					require.NoError(t, err)
				}))
				return ts
			},
			setupCZDSAPIMock: func() *httptest.Server {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/tlds", r.URL.Path)

					_, err := w.Write([]byte(`bad response`))
					require.NoError(t, err)
				}))
				return ts
			},
			errAssert: assert.Error,
		},
	} {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			mockAccountsAPI := tc.setupICANNAccountsAPIMock()
			defer mockAccountsAPI.Close()

			mockCZDSAPI := tc.setupCZDSAPIMock()
			defer mockCZDSAPI.Close()

			client := czds.NewClient(testEmail, testPassword,
				czds.ICANNAccountsAPIBaseURL(mockAccountsAPI.URL),
				czds.CZDSAPIBaseURL(mockCZDSAPI.URL))

			tlds, err := client.ListTLDs(context.Background())
			tc.errAssert(t, err)
			if err != nil {
				return
			}

			assert.ElementsMatch(t, tc.expectedTLDs, tlds)
		})
	}
}
