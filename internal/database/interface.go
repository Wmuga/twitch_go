package database

type DBConnection interface {
	SetPoints(username string, points int)
	AddPoints(username string, points int)
	TryRemovePoints(username string, points int) bool
	GetPoints(username string) int
	GetPointsTop20() map[string]int
	GetPointsTop5(owner string) map[string]int
	MassAddPoints(usernames []string, toAdd int)
}
