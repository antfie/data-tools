package main

import (
	"data-tools/config"
	"data-tools/models"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initDb(config *config.Config) *gorm.DB {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(getLogLevel(config)),
	}

	return connect(config.DBPath, gormConfig)
}

func getLogLevel(config *config.Config) logger.LogLevel {
	if config.IsDebug {
		return logger.Info
	}

	return logger.Silent
}

func testDB() *gorm.DB {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	return connect("file::memory:?cache=shared", gormConfig)
}

func connect(dsn string, gormConfig *gorm.Config) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)

	if err != nil {
		panic("failed to connect to the database")
	}

	err = db.AutoMigrate(&models.Path{}, &models.Root{}, &models.File{})

	if err != nil {
		panic("failed to migrate the database")
	}

	return db
}

const fileAbsolutePathCTEQuery = `
(
	WITH RECURSIVE path_cte(parent_path_id, level, name) AS
	(
		SELECT	f1.path_id,
				f1.level,
				f1.name
		FROM 	files f1
		WHERE 	f1.ignored = 0
		AND 	id = f.id
		
		UNION
		
		SELECT	p1.parent_path_id,
				p1.level,
				p1.name
		FROM 	files f2
		JOIN 	paths p1 ON f2.path_id = p1.id
		WHERE 	f2.ignored = 0
		AND 	f2.id = f.id
		
		UNION
		
		SELECT 	p2.parent_path_id,
				p2.level,
				p2.name
		FROM 	paths p2
		JOIN	path_cte ON p2.id = path_cte.parent_path_id
		WHERE 	p2.ignored = 0
	),
	path_ordered AS (
		SELECT 	 	*
		FROM 		path_cte
		ORDER BY	level
	)
	SELECT	group_concat(name, '/')
	FROM	path_ordered
) absolute_path
`

func QueryUnHashedFilePathsWithLimit() string {
	return fmt.Sprintf(`
SELECT		f.id,
			%s
FROM		files f
WHERE 		f.file_hash_id IS NULL
AND 		absolute_path IS NOT NULL
AND			f.ignored = 0
ORDER BY	f.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryUnSizedFileHashesWithLimit() string {
	return fmt.Sprintf(`
SELECT		fh.id,
			%s
FROM 		file_hashes fh
JOIN 		files f ON f.file_hash_id = fh.id
WHERE		fh.size IS NULL
AND			fh.ignored = 0
AND   		f.ignored = 0
ORDER BY	fh.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryUnTypedFileHashesWithLimit() string {
	return fmt.Sprintf(`
SELECT		fh.id,
			%s
FROM 		file_hashes fh
JOIN 		files f ON f.file_hash_id = fh.id
WHERE		fh.file_type_id IS NULL
AND			fh.ignored = 0
AND   		f.ignored = 0
ORDER BY	fh.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetFileHashesToZapWithLimit() string {
	return fmt.Sprintf(`
SELECT		fh.id,
      		fh.hash,
			%s
FROM 		file_hashes fh
JOIN  		(
				SELECT		f.id,
							max(f.file_hash_id) AS file_hash_id -- for deterministic result order
				FROM 		files f
				WHERE		f.deleted_at IS NULL
				AND			f.zapped = 0
				AND			f.ignored = 0
				GROUP BY 	f.file_hash_id
  			) f ON f.file_hash_id = fh.id
WHERE		fh.zapped = 0
AND			fh.ignored = 0
ORDER BY	fh.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}

func QueryGetZappedFileHashesToUnZapWithLimit() string {
	return fmt.Sprintf(`
SELECT		f.id,
      		fh.hash,
			%s
FROM 		files f
JOIN 		file_hashes fh ON f.file_hash_id = fh.id
WHERE		f.deleted_at IS NULL
AND			f.zapped = 1
AND			f.ignored = 0
AND			fh.zapped = 1
AND			fh.ignored = 0
AND			f.id NOT IN ?
ORDER BY	fh.id -- for deterministic result order
LIMIT 		?
`, fileAbsolutePathCTEQuery)
}
