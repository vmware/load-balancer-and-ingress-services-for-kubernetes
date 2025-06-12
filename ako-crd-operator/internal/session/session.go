package session

import (
	"context"
	"sync"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/constants"

	"github.com/vmware/alb-sdk/go/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/event"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

//go:generate mockgen -source=session.go -destination=../../test/mock/session_mock.go -package=mock
type AviRestClientPoolFactory interface {
	NewAviRestClientPool(numClients int, ctrlIpAddress, ctrlUsername, ctrlPassword, ctrlAuthToken, controllerVersion, ctrlCAData string, tenant string, protocol string, userHeaders map[string]string) (*utils.AviRestClientPool, string, error)
}

type AviRestClientPoolFactoryImpl struct{}

func (f *AviRestClientPoolFactoryImpl) NewAviRestClientPool(numClients int,
	ctrlIpAddress,
	ctrlUsername,
	ctrlPassword,
	ctrlAuthToken,
	controllerVersion,
	ctrlCAData string,
	tenant string,
	protocol string,
	userHeaders map[string]string) (*utils.AviRestClientPool, string, error) {

	return utils.NewAviRestClientPool(uint32(numClients), ctrlIpAddress, ctrlUsername, ctrlPassword, ctrlAuthToken, controllerVersion, ctrlCAData, tenant, protocol, userHeaders)
}

type Session struct {
	sync                     *sync.Mutex
	aviClientPool            *utils.AviRestClientPool
	aviRestClientPoolFactory AviRestClientPoolFactory
	ctrlProperties           map[string]string
	tenant                   string
	k8sClient                kubernetes.Interface
	eventManager             *event.EventManager
	status                   string
	controllerVersion        string
}

func NewSession(k8sClient kubernetes.Interface, eventManager *event.EventManager) *Session {
	return &Session{
		sync:                     &sync.Mutex{},
		aviRestClientPoolFactory: &AviRestClientPoolFactoryImpl{},
		ctrlProperties:           make(map[string]string),
		k8sClient:                k8sClient,
		status:                   utils.AVIAPI_INITIATING,
		eventManager:             eventManager,
	}
}

func (s *Session) PopulateControllerProperties(ctx context.Context) error {
	var err error
	s.ctrlProperties, err = lib.GetControllerPropertiesFromSecret(s.k8sClient)
	if err != nil {
		return err
	}
	return nil
}

func (s *Session) CreateAviClients(ctx context.Context, numClient int) {
	s.sync.Lock()
	defer s.sync.Unlock()

	log := utils.LoggerFromContext(ctx)

	ctrlUsername := s.ctrlProperties[utils.ENV_CTRL_USERNAME]
	ctrlPassword := s.ctrlProperties[utils.ENV_CTRL_PASSWORD]
	ctrlAuthToken := s.ctrlProperties[utils.ENV_CTRL_AUTHTOKEN]
	ctrlCAData := s.ctrlProperties[utils.ENV_CTRL_CADATA]
	ctrlIpAddress := lib.GetControllerIP()
	if ctrlUsername == "" || (ctrlPassword == "" && ctrlAuthToken == "") || ctrlIpAddress == "" {
		var passwordLog, authTokenLog string
		if ctrlPassword != "" {
			passwordLog = constants.Sensitive
		}
		if ctrlAuthToken != "" {
			authTokenLog = constants.Sensitive
		}
		s.eventManager.PodEventf(
			corev1.EventTypeWarning,
			lib.AKOShutdown, "Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s)",
			ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress,
		)
		if ctrlIpAddress == "" {
			log.Fatalf("Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s). Update the controller IP in ConfigMap : avi-k8s-config", ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress)
		}
		log.Fatalf("Avi Controller information missing (username: %s, password: %s, authToken: %s, controller: %s). Update them in avi-secret.", ctrlUsername, passwordLog, authTokenLog, ctrlIpAddress)
	}
	var err error
	// TODO: inject interface of AviClient instead directly using AviRestClient
	var aviRestClientPool *utils.AviRestClientPool

	aviRestClientPool, s.controllerVersion, err = s.aviRestClientPoolFactory.NewAviRestClientPool(
		numClient,
		ctrlIpAddress,
		ctrlUsername,
		ctrlPassword,
		ctrlAuthToken,
		"",
		ctrlCAData,
		s.tenant,
		"",
		nil,
	)
	if err != nil {
		s.status = utils.AVIAPI_DISCONNECTED
	} else {
		s.status = utils.AVIAPI_CONNECTED
		// set the controller version in avisession obj
		if aviRestClientPool != nil && aviRestClientPool.AviClient != nil {
			for _, client := range aviRestClientPool.AviClient {
				SetVersion := session.SetVersion(s.controllerVersion)
				SetVersion(client.AviSession)
			}
		}
	}
	s.aviClientPool = aviRestClientPool
}

func (s *Session) UpdateAviClients(ctx context.Context, numClient int) error {
	if err := s.PopulateControllerProperties(ctx); err != nil {
		return err
	}
	s.CreateAviClients(ctx, numClient)
	return nil
}

func (s *Session) GetAviClients() *utils.AviRestClientPool {
	return s.aviClientPool
}
