package support

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	helmsupport "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/helm"

	daprApi "github.com/dapr-sandbox/dapr-kubernetes-operator/api/operator/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/go-logr/logr/testr"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/onsi/gomega"
	"github.com/rs/xid"

	supportclient "github.com/dapr-sandbox/dapr-kubernetes-operator/test/support/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	corev1 "k8s.io/api/core/v1"

	olmV1 "github.com/operator-framework/api/pkg/operators/v1"
	olmV1Alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

const (
	TestTimeoutMini   = 5 * time.Second
	TestTimeoutShort  = 1 * time.Minute
	TestTimeoutMedium = 2 * time.Minute
	TestTimeoutLong   = 5 * time.Minute

	DefaultEventuallyPollingInterval   = 500 * time.Millisecond
	DefaultEventuallyTimeout           = TestTimeoutLong
	DefaultConsistentlyDuration        = 500 * time.Millisecond
	DefaultConsistentlyPollingInterval = 500 * time.Millisecond
)

func init() {
	if err := daprApi.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := olmV1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := olmV1Alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
}

type Test interface {
	gomega.Gomega

	T() *testing.T
	Ctx() context.Context

	ID() string
	Cleanup(f func() []runtime.Object)

	Client() *supportclient.Client
	Helm() *helmsupport.Helm
	HTTPClient() *http.Client

	NewTestNamespace(opts ...Option[*corev1.Namespace]) *corev1.Namespace
}

type Option[T any] interface {
	applyTo(to T) error
}

func With(t *testing.T) Test {
	t.Helper()

	lr := testr.New(t)
	klog.SetLogger(lr.WithName("client"))

	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		withDeadline, cancel := context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
		ctx = withDeadline
	}
	answer := &T{
		WithT:   gomega.NewWithT(t),
		id:      xid.New().String(),
		t:       t,
		ctx:     ctx,
		http:    cleanhttp.DefaultClient(),
		cleanup: make([]func() []runtime.Object, 0),
	}

	answer.SetDefaultEventuallyPollingInterval(DefaultEventuallyPollingInterval)
	answer.SetDefaultEventuallyTimeout(DefaultEventuallyTimeout)
	answer.SetDefaultConsistentlyDuration(DefaultConsistentlyDuration)
	answer.SetDefaultConsistentlyPollingInterval(DefaultConsistentlyPollingInterval)

	t.Cleanup(func() {
		t.Log("Run Test cleanup")

		allerr := make([]error, 0)

		for i := len(answer.cleanup) - 1; i >= 0; i-- {
			objects := answer.cleanup[i]()

			for i := range objects {
				err := runCleanup(answer, objects[i])
				if err != nil {
					allerr = append(allerr, err)
				}
			}
		}

		if len(allerr) != 0 {
			t.Fatal(errors.Join(allerr...))
		}

		t.Log("Test cleanup done")
	})

	return answer
}

type T struct {
	*gomega.WithT

	id         string
	t          *testing.T
	client     *supportclient.Client
	clientOnce sync.Once
	helm       *helmsupport.Helm
	helmOnce   sync.Once
	http       *http.Client
	cleanup    []func() []runtime.Object

	//nolint:containedctx
	ctx context.Context
}

func (t *T) ID() string {
	return t.id
}

func (t *T) T() *testing.T {
	return t.t
}

func (t *T) Ctx() context.Context {
	return t.ctx
}

func (t *T) Client() *supportclient.Client {
	t.clientOnce.Do(func() {
		c, err := supportclient.New(t.t)
		if err != nil {
			t.T().Fatalf("Error creating client: %v", err)
		}
		t.client = c
	})
	return t.client
}

func (t *T) Helm() *helmsupport.Helm {
	t.helmOnce.Do(func() {
		h, err := helmsupport.New(helmsupport.WithLog(func(s string, i ...interface{}) {
			t.T().Logf("[helm] "+s, i...)
		}))

		if err != nil {
			t.T().Fatalf("Error creating helm client: %v", err)
		}

		t.helm = h
	})

	return t.helm
}

func (t *T) HTTPClient() *http.Client {
	return t.http
}

func (t *T) Cleanup(f func() []runtime.Object) {
	t.cleanup = append(t.cleanup, f)
}

func (t *T) NewTestNamespace(options ...Option[*corev1.Namespace]) *corev1.Namespace {
	t.T().Helper()

	namespace := createTestNamespace(t, options...)

	t.Cleanup(func() []runtime.Object {
		deleteTestNamespace(t, namespace)
		return nil
	})

	return namespace
}
