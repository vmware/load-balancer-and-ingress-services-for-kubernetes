package session

import (
	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"
)

//go:generate mockgen -source=aviclient.go -destination=../../test/mock/aviclient_mock.go -package=mock
type AviSessionClient struct {
	AviClient *clients.AviClient
}

func NewAviSessionClient(aviClient *clients.AviClient) AviClientInterface {
	return &AviSessionClient{
		AviClient: aviClient,
	}
}

type AviClientInterface interface {
	// currently only required functions are defined here
	GetAviSession() *session.AviSession
	AviSessionGet(url string, response interface{}, params ...session.ApiOptionsParams) error
	AviSessionGetCollectionRaw(url string, params ...session.ApiOptionsParams) (session.AviCollectionResult, error)
	AviSessionPost(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error
	AviSessionPut(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error
	AviSessionDelete(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error
}

func (s *AviSessionClient) AviSessionGet(url string, response interface{}, params ...session.ApiOptionsParams) error {
	return s.AviClient.AviSession.Get(url, response, params...)
}

func (s *AviSessionClient) AviSessionPut(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error {
	return s.AviClient.AviSession.Put(url, request, response, params...)
}

func (s *AviSessionClient) AviSessionDelete(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error {
	return s.AviClient.AviSession.Delete(url, request, response)
}

func (s *AviSessionClient) AviSessionPost(url string, request interface{}, response interface{}, params ...session.ApiOptionsParams) error {
	return s.AviClient.AviSession.Post(url, request, response, params...)
}

func (s *AviSessionClient) GetAviSession() *session.AviSession {
	return s.AviClient.AviSession
}

func (s *AviSessionClient) AviSessionGetCollectionRaw(url string, params ...session.ApiOptionsParams) (session.AviCollectionResult, error) {
	return s.AviClient.AviSession.GetCollectionRaw(url, params...)
}
