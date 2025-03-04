package core

// CertConfig contient la configuration des certificats
type CertConfig struct {
	Enabled           bool
	JDKSecurityPath   string
	SystemCacertsPath string
}

// DownloadOptions contient les options pour le téléchargement et l'installation
type DownloadOptions struct {
	DownloadURL   string
	CacheDir      string
	InstallPath   string
	SDKType       string
	Distribution  string
	Version       string
	KeepCache     bool
	CertConfig    CertConfig
}
