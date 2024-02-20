package database

type DBConnection interface {
	// Set username.points = {points}
	SetPoints(username string, points int)
	// Adds {points} to username.points
	AddPoints(username string, points int)
	// Removes {points} from username.points if result non negative
	TryRemovePoints(username string, points int) bool
	// Returns username.points
	GetPoints(username string) int
	// Returns top 20 users by points
	GetPointsTop20() map[string]int
	// Returns top 5 users by points excluding owner
	GetPointsTop5(owner string) map[string]int
	// Adds {toAdd} to all users in {usernames}
	MassAddPoints(usernames []string, toAdd int)
}
