package drive

type Drive interface {
	Upload(string)
	Download(string) error
	Login(string, string) error
	Info(string) error
}
