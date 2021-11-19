package serviceaccount

import (
	"context"
	"sync"

	"github.com/gotidy/ptr"
	"github.com/sirupsen/logrus"
	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typedv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ServiceAccountClient struct {
	KubernetesInterface kubernetes.Interface
	Token               *Token
	TokenRequestObj     *authv1.TokenRequest
	Client              typedv1.ServiceAccountInterface
	sync.Mutex
}

type Token struct {
	AuthToken           *string
	ExpirationTimestamp metav1.Time
}

func (s *ServiceAccountClient) NewServiceAccountToken(targetServiceAccountName string, audience string, expireSecs int64) (string, metav1.Time, error) {
	s.Lock()
	defer s.Unlock()

	now := metav1.Now().UTC()
	opts := metav1.CreateOptions{}
	tr := s.newTokenRequest(audience, expireSecs)

	if s.Token == nil {
		logrus.Debug("auth token from service account not found")
		return s.getToken(targetServiceAccountName, s.Client, tr, opts)
	}

	if now.UTC().After(s.Token.ExpirationTimestamp.Time) {
		logrus.Debug("service account token is expired")
		return s.getToken(targetServiceAccountName, s.Client, tr, opts)
	}

	return *s.Token.AuthToken, s.Token.ExpirationTimestamp, nil
}

func NewServiceAccountClient(namespace string, kubernetesInterface kubernetes.Interface) *ServiceAccountClient {
	return &ServiceAccountClient{
		Client: kubernetesInterface.CoreV1().ServiceAccounts(namespace),
	}
}

func (s *ServiceAccountClient) newTokenRequest(audience string, expireSeconds int64) *authv1.TokenRequest {
	if len(audience) != 0 {
		return &authv1.TokenRequest{
			Spec: authv1.TokenRequestSpec{
				Audiences:         []string{audience},
				ExpirationSeconds: ptr.Int64(expireSeconds),
			},
		}
	} else {
		return &authv1.TokenRequest{
			Spec: authv1.TokenRequestSpec{
				ExpirationSeconds: ptr.Int64(expireSeconds),
			},
		}
	}
}

func (s *ServiceAccountClient) getToken(targetServiceAccount string, client typedv1.ServiceAccountInterface, tr *authv1.TokenRequest, opts metav1.CreateOptions) (string, metav1.Time, error) {
	tr, err := client.CreateToken(context.TODO(), targetServiceAccount, tr, opts)
	if err != nil {
		return "", metav1.Now(), nil
	}

	s.Token = &Token{
		AuthToken:           ptr.String(tr.Status.Token),
		ExpirationTimestamp: tr.Status.ExpirationTimestamp,
	}

	token := tr.Status.Token
	return token, tr.Status.ExpirationTimestamp, nil
}
