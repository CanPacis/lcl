package i18n

import "time"

func proc_year(t time.Time) string {
	return strconv_ext_i.Itoa(time_ext_i.Year(t))
}
