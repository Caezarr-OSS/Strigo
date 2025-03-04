package downloader

// CertConfig contient la configuration des certificats
type CertConfig struct {
	// Enabled indique si la configuration des certificats est activée
	Enabled bool

	// JDKSecurityPath est le chemin relatif vers le fichier cacerts dans le JDK
	// Par exemple : "lib/security/cacerts"
	JDKSecurityPath string

	// SystemCacertsPath est le chemin absolu vers les certificats système
	// Par exemple : "/etc/ssl/certs/java/cacerts"
	SystemCacertsPath string
}
