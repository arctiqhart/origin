package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	g "github.com/onsi/ginkgo"
	o "github.com/onsi/gomega"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	"github.com/openshift/library-go/pkg/oauth/oauthdiscovery"

	exutil "github.com/openshift/origin/test/extended/util"
	"github.com/openshift/origin/test/extended/util/ibmcloud"
)

var _ = g.Describe("[sig-auth][Feature:OAuthServer] well-known endpoint", func() {
	defer g.GinkgoRecover()
	var (
		oc             = exutil.NewCLI("oauth-well-known")
		oauthRoute     = "oauth-openshift"
		oauthNamespace = "openshift-authentication"
	)

	g.It("should be reachable", func() {
		metadataJSON, err := oc.Run("get").Args("--raw", "/.well-known/oauth-authorization-server").Output()
		o.Expect(err).NotTo(o.HaveOccurred())

		metadata := &oauthdiscovery.OauthAuthorizationServerMetadata{}
		err = json.Unmarshal([]byte(metadataJSON), metadata)
		o.Expect(err).NotTo(o.HaveOccurred())

		// If not running on an IBM ROKS cluster,
		// compare to openshift-authentication route
		// (On a ROKS cluster the openshift-authentication route does not live in the cluster)
		if e2e.TestContext.Provider != ibmcloud.ProviderName {
			route, err := oc.AdminRouteClient().RouteV1().Routes(oauthNamespace).Get(context.Background(), oauthRoute, metav1.GetOptions{})
			o.Expect(err).NotTo(o.HaveOccurred())
			u, err := url.Parse("https://" + route.Spec.Host)
			o.Expect(err).NotTo(o.HaveOccurred())
			u.Path = path.Join(u.Path, "oauth/authorize")
			authEndpointFromRoute := u.String()
			o.Expect(metadata.AuthorizationEndpoint).To(o.Equal(authEndpointFromRoute))
		}
		tlsClientConfig, err := rest.TLSConfigFor(oc.AdminConfig())
		o.Expect(err).NotTo(o.HaveOccurred())

		rt := http.Transport{
			TLSClientConfig: tlsClientConfig,
		}

		req, err := http.NewRequest(http.MethodHead, metadata.Issuer, nil)
		o.Expect(err).NotTo(o.HaveOccurred())

		resp, err := rt.RoundTrip(req)
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(resp.StatusCode).To(o.Equal(200))
	})
})
