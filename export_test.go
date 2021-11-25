package cache

import "time"

func SetNowFunc(tm time.Time) (reset func()) {
	backup := nowFunc
	nowFunc = func() time.Time { return tm }
	return func() {
		nowFunc = backup
	}
}
