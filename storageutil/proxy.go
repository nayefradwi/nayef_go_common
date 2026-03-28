package storageutil

type ProxyHandlerConfig struct {
	AllowedContentTypes []string
	MaxUploadSize       int64
}

type ProxyHandler struct {
	objectManager ObjectManager
	config        ProxyHandlerConfig
}

func NewProxyHandler(objectManager ObjectManager) ProxyHandler {
	return ProxyHandler{objectManager: objectManager}
}

func (p ProxyHandler) SetAllowedContentTypes(types []string) ProxyHandler {
	p.config.AllowedContentTypes = types
	return p
}

func (p ProxyHandler) SetMaxUploadSize(size int64) ProxyHandler {
	p.config.MaxUploadSize = size
	return p
}
