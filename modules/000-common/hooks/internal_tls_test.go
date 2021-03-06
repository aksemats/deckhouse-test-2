/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hooks

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

type tlsTest struct {
	ca  string
	crt string
	key string
}

type tlsTestFixtures struct {
	cert  tlsTest
	state string
}

func setupHookTest(cert tlsTest) tlsTestFixtures {
	encode := func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}

	state := fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: module-name-internal-tls
  namespace: d8-module-name
data:
  ca.crt: %s
  tls.crt: %s
  tls.key: %s
`, encode(cert.ca), encode(cert.crt), encode(cert.key))

	return tlsTestFixtures{
		cert:  cert,
		state: state,
	}
}

var (
	secretCreatedFixture = setupHookTest(tlsTest{
		ca:  "a",
		crt: "b",
		key: "c",
	})

	secretChangedFixture = setupHookTest(tlsTest{
		ca:  "x",
		crt: "y",
		key: "z",
	})
)

var _ = Describe("Modules :: common :: hooks :: internal_tls", func() {
	f := HookExecutionConfigInit(`{"moduleName":{"internal":{}}}`, "{}")

	Context("For empty cluster", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("executes successful with empty state", func() {
			Expect(f).To(ExecuteSuccessfully())
		})

		Context("when secret created", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSet(secretCreatedFixture.state))
				f.RunHook()
			})

			It("stores tls into values", func() {
				Expect(f).To(ExecuteSuccessfully())
				assertTLSStoredIntoValues(f, secretCreatedFixture.cert)
			})

			Context("when secret changed", func() {
				BeforeEach(func() {
					f.BindingContexts.Set(f.KubeStateSet(secretChangedFixture.state))
					f.RunHook()
				})

				It("stores new tls data into values", func() {
					Expect(f).To(ExecuteSuccessfully())
					assertTLSStoredIntoValues(f, secretChangedFixture.cert)
				})
			})
		})

		Context("when fire onBeforeHelm event", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.GenerateBeforeHelmContext())
				f.RunHook()
			})

			It("generates and stores new tls data to values", func() {
				Expect(f).To(ExecuteSuccessfully())

				assertExistsTLSInValues(f)
			})

			It("stores tls which is signed by CA for all passed SANs", func() {
				Expect(f).To(ExecuteSuccessfully())

				certFields := assertExistsTLSInValues(f)
				assertCaSignTLS(certFields, "module.d8-module-name")
				assertCaSignTLS(certFields, "127.0.0.1")
			})

		})
	})

	Context("For cluster with secret", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(secretCreatedFixture.state))
			f.RunHook()
		})

		It("stores tls data in values", func() {
			Expect(f).To(ExecuteSuccessfully())
			assertTLSStoredIntoValues(f, secretCreatedFixture.cert)
		})

		Context("when delete secret", func() {
			BeforeEach(func() {
				f.BindingContexts.Set(f.KubeStateSet(""))
				f.RunHook()
			})

			It("generates new tls and stores new tls data to values", func() {
				Expect(f).To(ExecuteSuccessfully())

				cert := assertExistsTLSInValues(f)
				assertNotEqualsCerts(cert, secretCreatedFixture.cert)
			})

			It("stores tls which is signed by CA for all passed SANs", func() {
				Expect(f).To(ExecuteSuccessfully())

				certFields := assertExistsTLSInValues(f)

				assertCaSignTLS(certFields, "module.d8-module-name")
				assertCaSignTLS(certFields, "127.0.0.1")
			})
		})
	})

})

func assertNotEqualsCerts(a tlsTest, b tlsTest) {
	Expect(a.ca).To(Not(Equal(b.ca)))
	Expect(a.crt).To(Not(Equal(b.crt)))
	Expect(a.key).To(Not(Equal(b.key)))
}

func assertTLSStoredIntoValues(f *HookExecutionConfig, cert tlsTest) {
	Expect(f).To(ExecuteSuccessfully())

	certFromValues := assertExistsTLSInValues(f)

	Expect(certFromValues.ca).To(Equal(cert.ca))
	Expect(certFromValues.crt).To(Equal(cert.crt))
	Expect(certFromValues.key).To(Equal(cert.key))
}

func assertCaSignTLS(certFields tlsTest, dnsName string) {
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM([]byte(certFields.ca))
	Expect(ok).To(BeTrue())

	block, _ := pem.Decode([]byte(certFields.crt))
	Expect(block).ShouldNot(BeNil())

	cert, err := x509.ParseCertificate(block.Bytes)
	Expect(err).ShouldNot(HaveOccurred())

	opts := x509.VerifyOptions{
		DNSName: dnsName,
		Roots:   certPool,
	}

	_, err = cert.Verify(opts)
	Expect(err).ShouldNot(HaveOccurred())
}

func assertExistsTLSInValues(f *HookExecutionConfig) tlsTest {
	ca := f.ValuesGet("moduleName.internal.moduleNameCA")
	crt := f.ValuesGet("moduleName.internal.moduleNamePem")
	key := f.ValuesGet("moduleName.internal.moduleNameKey")

	Expect(ca.Exists()).To(BeTrue())
	Expect(crt.Exists()).To(BeTrue())
	Expect(key.Exists()).To(BeTrue())

	cert := tlsTest{
		ca:  ca.String(),
		crt: crt.String(),
		key: key.String(),
	}

	return cert
}
