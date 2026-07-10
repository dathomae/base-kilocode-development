package domain

type LocalizedString struct {
	Bounded
	Name        string
	Description string
	Tooltip     string
	Country     string
	Language    string
}

type Named struct {
	curName       string
	curDesc       string
	curTooltip    string
	curIndex      int
	localizations []LocalizedString
}

func (n *Named) ChooseLocalization(country string, language string) *LocalizedString {
	return nil
}

func (n *Named) GetLocalizedString(country string, language string) *LocalizedString {
	return nil
}

func (n *Named) SetLocalizedString(country string, language string, name string, description string, tooltip string) {
}
