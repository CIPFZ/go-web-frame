package seed

// MenuOrder is shared by fresh database seeding and the one-time order upgrade.
// Leave gaps for business modules before the status and about pages.
var MenuOrder = map[string]int{
	"/dashboard/workplace": 10,
	"/sys":                 20,
	"/state":               90,
	"/about":               100,
	"/account/settings":    999,
	"/sys/user":            10,
	"/sys/authority":       20,
	"/sys/menu":            30,
	"/sys/api":             40,
	"/sys/api-token":       50,
	"/sys/notice":          60,
	"/sys/operation":       70,
}
