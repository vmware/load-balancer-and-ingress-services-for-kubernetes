package session

import "github.com/vmware/alb-sdk/go/clients"

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
	AviSessionGet(url string, response interface{}) error
	AviSessionPost(url string, request interface{}, response interface{}) error
	AviSessionPut(url string, request interface{}, response interface{}) error
	AviSessionDelete(url string, request interface{}, response interface{}) error
}

func (s *AviSessionClient) AviSessionGet(url string, response interface{}) error {
	return s.AviClient.AviSession.Get(url, response)
}

func (s *AviSessionClient) AviSessionPut(url string, request interface{}, response interface{}) error {
	return s.AviClient.AviSession.Put(url, request, response)
}

func (s *AviSessionClient) AviSessionDelete(url string, request interface{}, response interface{}) error {
	return s.AviClient.AviSession.Delete(url, request, response)
}

func (s *AviSessionClient) AviSessionPost(url string, request interface{}, response interface{}) error {
	return s.AviClient.AviSession.Post(url, request, response)
}
