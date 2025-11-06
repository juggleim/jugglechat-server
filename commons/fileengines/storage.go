package fileengines

type PreSignedURLGenerator interface {
	PreSignedURL(fileType string, dir string) (url string, err error)
}
