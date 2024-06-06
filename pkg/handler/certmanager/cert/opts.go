package cert

type Opts func(*opts) error

type opts struct {
	ecdsaCurve string // P224, P256 (default), P384, P521
	rsaBits    int    // 2048 (default), 4096
}
