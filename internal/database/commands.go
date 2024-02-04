package database

const (
	createTableSql          = "create table if not exists points (nickname varchar(50), count smallint)"
	getPointsViewerSql      = "select count from points where nickname = ?"
	getPointsViewerMax20Sql = "select * from points order by count desc limit 20"
	getPointsTopSql         = "select * from points where nickname <> ? order by count desc limit 5"
	insertPointsViewerSql   = "insert into points (nickname, count) values (?, ?)"
	setPointsViewerSql      = "update points set count = ? where nickname = ?"
	addPointsViewerSql      = "update points set count = count + ? where nickname = ?"
)
