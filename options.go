package czds

type Options struct {
	tokenStore         TokenStore
	accountsAPIBaseURL string
	czdsAPIBaseURL     string
}

type ClientOption func(*Options)

func TokenStoreOpt(store TokenStore) ClientOption {
	return func(opts *Options) {
		opts.tokenStore = store
	}
}

func ICANNAccountsAPIBaseURL(baseURL string) ClientOption {
	return func(opts *Options) {
		opts.accountsAPIBaseURL = baseURL
	}
}

func CZDSAPIBaseURL(baseURL string) ClientOption {
	return func(opts *Options) {
		opts.czdsAPIBaseURL = baseURL
	}
}
