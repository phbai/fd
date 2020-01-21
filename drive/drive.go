package drive

type Drive interface {
	Upload(string) error
	Download(string) error
	Login(string, string) error
	Info(string) error
}
